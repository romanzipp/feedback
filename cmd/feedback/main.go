package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/romanzipp/feedback/internal/config"
	"github.com/romanzipp/feedback/internal/database"
	"github.com/romanzipp/feedback/internal/handlers"
	"github.com/romanzipp/feedback/internal/middleware"
	"github.com/romanzipp/feedback/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	uploadsDir := filepath.Join(cfg.DataDir, "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Open database
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database initialized successfully")

	// Initialize services
	shareService := services.NewShareService(db)
	fileService := services.NewFileService(db, cfg.DataDir)

	// Initialize session store
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	// Load templates
	funcMap := template.FuncMap{
		"hasPrefix": func(s, prefix string) bool {
			return len(s) >= len(prefix) && s[:len(prefix)] == prefix
		},
	}

	// Admin templates
	adminTmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("web/templates/layouts/*.html"))
	adminTmpl = template.Must(adminTmpl.ParseGlob("web/templates/admin/*.html"))

	// Public templates
	publicTmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("web/templates/layouts/*.html"))
	publicTmpl = template.Must(publicTmpl.ParseGlob("web/templates/public/*.html"))

	// Initialize handlers
	adminHandler := handlers.NewAdminHandler(adminTmpl, shareService, fileService)
	shareHandler := handlers.NewShareHandler(publicTmpl, shareService, fileService, store)
	fileHandler := handlers.NewFileHandler(fileService)
	commentHandler := handlers.NewCommentHandler(fileService)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.UserSession(store))

		r.Get("/share/{hash}", shareHandler.View)
		r.Post("/share/{hash}/name", shareHandler.SetUsername)
		r.Post("/api/files/{hash}/comments", commentHandler.Create)
	})

	// File download (no auth needed if you have the hash)
	r.Get("/files/{hash}", fileHandler.Download)

	// Admin routes
	r.Route("/admin/{token}", func(r chi.Router) {
		r.Use(middleware.AdminAuth(cfg.AdminToken))

		r.Get("/", adminHandler.Dashboard)
		r.Get("/shares/new", adminHandler.NewShareForm)
		r.Post("/shares", adminHandler.CreateShare)
		r.Get("/shares/{id}", adminHandler.ShareDetail)
		r.Post("/shares/{id}/upload", adminHandler.UploadFile)
		r.Post("/shares/{id}/delete", adminHandler.DeleteShare)
		r.Post("/files/{id}/delete", adminHandler.DeleteFile)
	})

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	log.Printf("Admin URL: http://%s/admin/%s", addr, cfg.AdminToken)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
