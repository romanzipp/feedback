package services

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/romanzipp/feedback/internal/database"
)

type FileService struct {
	db      *sql.DB
	dataDir string
}

func NewFileService(db *sql.DB, dataDir string) *FileService {
	return &FileService{
		db:      db,
		dataDir: dataDir,
	}
}

func (s *FileService) Save(shareID int, fileHeader *multipart.FileHeader) (*database.File, error) {
	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Generate random hash for file access
	fileHash, err := GenerateHash(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file hash: %w", err)
	}

	// Generate unique filename
	fileID := uuid.New().String()
	storageName := fmt.Sprintf("%s_%s", fileID, fileHeader.Filename)

	// Create share directory
	shareDir := filepath.Join(s.dataDir, "uploads", fmt.Sprintf("%d", shareID))
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create share directory: %w", err)
	}

	// Full storage path
	storagePath := filepath.Join(shareDir, storageName)

	// Create destination file
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Detect MIME type
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Save to database
	result, err := s.db.Exec(
		"INSERT INTO files (share_id, hash, filename, storage_path, mime_type, size_bytes) VALUES (?, ?, ?, ?, ?, ?)",
		shareID, fileHash, fileHeader.Filename, storagePath, mimeType, size,
	)
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(storagePath)
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetByID(int(id))
}

func (s *FileService) GetByID(id int) (*database.File, error) {
	file := &database.File{}
	err := s.db.QueryRow(
		"SELECT id, share_id, hash, filename, storage_path, mime_type, size_bytes, uploaded_at FROM files WHERE id = ?",
		id,
	).Scan(&file.ID, &file.ShareID, &file.Hash, &file.Filename, &file.StoragePath, &file.MimeType, &file.SizeBytes, &file.UploadedAt)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *FileService) GetByHash(hash string) (*database.File, error) {
	file := &database.File{}
	err := s.db.QueryRow(
		"SELECT id, share_id, hash, filename, storage_path, mime_type, size_bytes, uploaded_at FROM files WHERE hash = ?",
		hash,
	).Scan(&file.ID, &file.ShareID, &file.Hash, &file.Filename, &file.StoragePath, &file.MimeType, &file.SizeBytes, &file.UploadedAt)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *FileService) GetByShareID(shareID int) ([]database.File, error) {
	rows, err := s.db.Query(
		"SELECT id, share_id, hash, filename, storage_path, mime_type, size_bytes, uploaded_at FROM files WHERE share_id = ? ORDER BY uploaded_at DESC",
		shareID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []database.File
	for rows.Next() {
		var f database.File
		err := rows.Scan(&f.ID, &f.ShareID, &f.Hash, &f.Filename, &f.StoragePath, &f.MimeType, &f.SizeBytes, &f.UploadedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, nil
}

func (s *FileService) Delete(id int) error {
	// Get file info first
	file, err := s.GetByID(id)
	if err != nil {
		return err
	}

	// Delete from database
	result, err := s.db.Exec("DELETE FROM files WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}

	// Delete physical file
	if err := os.Remove(file.StoragePath); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to delete file %s: %v\n", file.StoragePath, err)
	}

	return nil
}

func (s *FileService) GetComments(fileID int) ([]database.Comment, error) {
	rows, err := s.db.Query(
		"SELECT id, file_id, username, content, created_at FROM comments WHERE file_id = ? ORDER BY created_at ASC",
		fileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []database.Comment
	for rows.Next() {
		var c database.Comment
		err := rows.Scan(&c.ID, &c.FileID, &c.Username, &c.Content, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func (s *FileService) AddComment(fileID int, username, content string) (*database.Comment, error) {
	result, err := s.db.Exec(
		"INSERT INTO comments (file_id, username, content) VALUES (?, ?, ?)",
		fileID, username, content,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	comment := &database.Comment{}
	err = s.db.QueryRow(
		"SELECT id, file_id, username, content, created_at FROM comments WHERE id = ?",
		id,
	).Scan(&comment.ID, &comment.FileID, &comment.Username, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}

	return comment, nil
}
