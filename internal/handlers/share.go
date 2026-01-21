package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/romanzipp/feedback/internal/database"
	"github.com/romanzipp/feedback/internal/middleware"
	"github.com/romanzipp/feedback/internal/services"
)

type ShareHandler struct {
	templates    *template.Template
	shareService *services.ShareService
	fileService  *services.FileService
	store        *sessions.CookieStore
}

func NewShareHandler(templates *template.Template, shareService *services.ShareService, fileService *services.FileService, store *sessions.CookieStore) *ShareHandler {
	return &ShareHandler{
		templates:    templates,
		shareService: shareService,
		fileService:  fileService,
		store:        store,
	}
}

func (h *ShareHandler) View(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	username := middleware.GetUsername(r)

	share, err := h.shareService.GetByHash(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load share", http.StatusInternalServerError)
		return
	}

	files, err := h.fileService.GetByShareID(share.ID)
	if err != nil {
		http.Error(w, "Failed to load files", http.StatusInternalServerError)
		return
	}

	// Load comments for each file
	filesWithComments := make([]database.FileWithComments, 0, len(files))
	for _, file := range files {
		comments, err := h.fileService.GetComments(file.ID)
		if err != nil {
			http.Error(w, "Failed to load comments", http.StatusInternalServerError)
			return
		}
		filesWithComments = append(filesWithComments, database.FileWithComments{
			File:     file,
			Comments: comments,
		})
	}

	data := map[string]interface{}{
		"Share":    share,
		"Files":    filesWithComments,
		"Username": username,
		"Hash":     hash,
	}

	if err := h.templates.ExecuteTemplate(w, "public/share.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ShareHandler) SetUsername(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	session, _ := h.store.Get(r, "user-session")
	session.Values["username"] = username
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/share/"+hash, http.StatusSeeOther)
}
