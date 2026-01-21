package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/romanzipp/feedback/internal/services"
)

type AdminHandler struct {
	templates    *template.Template
	shareService *services.ShareService
	fileService  *services.FileService
}

func NewAdminHandler(templates *template.Template, shareService *services.ShareService, fileService *services.FileService) *AdminHandler {
	return &AdminHandler{
		templates:    templates,
		shareService: shareService,
		fileService:  fileService,
	}
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	shares, err := h.shareService.List()
	if err != nil {
		http.Error(w, "Failed to load shares", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Token":  token,
		"Shares": shares,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) NewShareForm(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	data := map[string]interface{}{
		"Token": token,
	}

	if err := h.templates.ExecuteTemplate(w, "share_form", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) CreateShare(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	share, err := h.shareService.Create(name, description)
	if err != nil {
		http.Error(w, "Failed to create share", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/"+token+"/shares/"+strconv.Itoa(share.ID), http.StatusSeeOther)
}

func (h *AdminHandler) ShareDetail(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	shareID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	share, err := h.shareService.GetByID(shareID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load share", http.StatusInternalServerError)
		return
	}

	files, err := h.fileService.GetByShareID(shareID)
	if err != nil {
		http.Error(w, "Failed to load files", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Token": token,
		"Share": share,
		"Files": files,
	}

	if err := h.templates.ExecuteTemplate(w, "share_detail", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	shareID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := h.fileService.Save(shareID, header); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/"+token+"/shares/"+strconv.Itoa(shareID), http.StatusSeeOther)
}

func (h *AdminHandler) DeleteShare(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	shareID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.shareService.Delete(shareID); err != nil {
		http.Error(w, "Failed to delete share", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/"+token, http.StatusSeeOther)
}

func (h *AdminHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	file, err := h.fileService.GetByID(fileID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.fileService.Delete(fileID); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	token := chi.URLParam(r, "token")
	http.Redirect(w, r, "/admin/"+token+"/shares/"+strconv.Itoa(file.ShareID), http.StatusSeeOther)
}
