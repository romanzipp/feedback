package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/romanzipp/feedback/internal/services"
)

type FileHandler struct {
	fileService *services.FileService
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	fileHash := chi.URLParam(r, "hash")
	if fileHash == "" {
		http.NotFound(w, r)
		return
	}

	file, err := h.fileService.GetByHash(fileHash)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Open file
	f, err := os.Open(file.StoragePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Set headers
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+file.Filename+"\"")

	// Serve file
	http.ServeContent(w, r, file.Filename, file.UploadedAt, f)
}
