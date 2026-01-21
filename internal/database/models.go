package database

import "time"

type Share struct {
	ID          int
	Hash        string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type File struct {
	ID          int
	ShareID     int
	Hash        string
	Filename    string
	StoragePath string
	MimeType    string
	SizeBytes   int64
	UploadedAt  time.Time
}

type Comment struct {
	ID        int
	FileID    int
	Username  string
	Content   string
	CreatedAt time.Time
}

type ShareWithStats struct {
	Share
	FileCount    int
	CommentCount int
}

type FileWithComments struct {
	File
	Comments []Comment
}
