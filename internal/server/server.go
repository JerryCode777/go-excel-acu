package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jerryandersonh/goexcel/config"
	"github.com/jerryandersonh/goexcel/internal/database"
	apiHandlers "github.com/jerryandersonh/goexcel/internal/handlers"
)

type Server struct {
	config  *config.Config
	db      *database.DB
	router  *mux.Router
	handler *apiHandlers.ProyectoHandler
}

func NewServer(cfg *config.Config, db *database.DB) *Server {
	server := &Server{
		config: cfg,
		db:     db,
		router: mux.NewRouter(),
	}

	server.handler = apiHandlers.NewProyectoHandler(db, cfg)
	server.setupRoutes()
	server.setupMiddleware()

	return server
}

func (s *Server) setupRoutes() {
	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", s.healthCheck).Methods("GET")

	// Projects routes
	projects := api.PathPrefix("/projects").Subrouter()
	projects.HandleFunc("", s.handler.GetProjects).Methods("GET")
	projects.HandleFunc("", s.handler.CreateProject).Methods("POST")
	projects.HandleFunc("/{id}", s.handler.GetProject).Methods("GET")
	projects.HandleFunc("/{id}", s.handler.UpdateProject).Methods("PUT")
	projects.HandleFunc("/{id}", s.handler.DeleteProject).Methods("DELETE")
	projects.HandleFunc("/{id}/export", s.handler.ExportProject).Methods("GET")
	projects.HandleFunc("/{id}/acu", s.handler.GetProjectACU).Methods("GET")
	projects.HandleFunc("/{id}/hierarchy", s.handler.GetProjectHierarchy).Methods("GET")
	projects.HandleFunc("/{id}/titles", s.handler.GetProjectTitles).Methods("GET")
	projects.HandleFunc("/{id}/titles", s.handler.UpdateProjectTitles).Methods("PUT")

	// ACU validation
	api.HandleFunc("/validate-acu", s.handler.ValidateACU).Methods("POST")

	// Static files and React app (for production)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/build/")))
}

func (s *Server) setupMiddleware() {
	// No necesitamos aplicar middleware al router directamente
	// El middleware se aplicará en el Start() method
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	dbStatus := "connected"
	if err := s.db.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "2.0.0",
		"database":  dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) Start() error {
	addr := s.config.GetServerAddress()
	log.Printf("🚀 Servidor iniciando en %s", addr)
	log.Printf("📡 API disponible en http://%s/api/v1", addr)
	log.Printf("🌐 Health check en http://%s/api/v1/health", addr)

	// Aplicar middleware
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins(s.config.CORS.AllowedOrigins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Logging
	finalHandler := handlers.LoggingHandler(log.Writer(), corsHandler(s.router))

	srv := &http.Server{
		Handler:      finalHandler,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}

func (s *Server) Stop() error {
	log.Println("🛑 Deteniendo servidor...")
	return s.db.Close()
}