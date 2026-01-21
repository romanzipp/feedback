package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/romanzipp/feedback/internal/middleware"
	"github.com/romanzipp/feedback/internal/services"
	"golang.org/x/time/rate"
)

type CommentHandler struct {
	fileService *services.FileService
	limiter     *rate.Limiter
}

func NewCommentHandler(fileService *services.FileService) *CommentHandler {
	return &CommentHandler{
		fileService: fileService,
		limiter:     rate.NewLimiter(1, 5), // 1 request per second, burst of 5
	}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	if !h.limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	username := middleware.GetUsername(r)
	if username == "" {
		http.Error(w, "Username not set", http.StatusUnauthorized)
		return
	}

	fileID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	comment, err := h.fileService.AddComment(fileID, username, content)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}
