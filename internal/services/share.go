package services

import (
	"database/sql"
	"fmt"

	"github.com/romanzipp/feedback/internal/database"
)

type ShareService struct {
	db *sql.DB
}

func NewShareService(db *sql.DB) *ShareService {
	return &ShareService{db: db}
}

func (s *ShareService) Create(name, description string) (*database.Share, error) {
	// Generate unique hash
	var hash string
	for {
		var err error
		hash, err = GenerateHash(12)
		if err != nil {
			return nil, err
		}

		// Check if hash already exists
		var exists bool
		err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM shares WHERE hash = ?)", hash).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
	}

	result, err := s.db.Exec(
		"INSERT INTO shares (hash, name, description) VALUES (?, ?, ?)",
		hash, name, description,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetByID(int(id))
}

func (s *ShareService) GetByID(id int) (*database.Share, error) {
	share := &database.Share{}
	err := s.db.QueryRow(
		"SELECT id, hash, name, description, created_at, updated_at FROM shares WHERE id = ?",
		id,
	).Scan(&share.ID, &share.Hash, &share.Name, &share.Description, &share.CreatedAt, &share.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return share, nil
}

func (s *ShareService) GetByHash(hash string) (*database.Share, error) {
	share := &database.Share{}
	err := s.db.QueryRow(
		"SELECT id, hash, name, description, created_at, updated_at FROM shares WHERE hash = ?",
		hash,
	).Scan(&share.ID, &share.Hash, &share.Name, &share.Description, &share.CreatedAt, &share.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return share, nil
}

func (s *ShareService) List() ([]database.ShareWithStats, error) {
	rows, err := s.db.Query(`
		SELECT
			s.id, s.hash, s.name, s.description, s.created_at, s.updated_at,
			COUNT(DISTINCT f.id) as file_count,
			COUNT(DISTINCT c.id) as comment_count
		FROM shares s
		LEFT JOIN files f ON s.id = f.share_id
		LEFT JOIN comments c ON f.id = c.file_id
		GROUP BY s.id
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []database.ShareWithStats
	for rows.Next() {
		var s database.ShareWithStats
		err := rows.Scan(
			&s.ID, &s.Hash, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt,
			&s.FileCount, &s.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}

	return shares, nil
}

func (s *ShareService) Delete(id int) error {
	result, err := s.db.Exec("DELETE FROM shares WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("share not found")
	}

	return nil
}
