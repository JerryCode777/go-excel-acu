package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"goexcel/config"
	"goexcel/internal/auth"
	"goexcel/internal/database"
	"goexcel/internal/database/repositories"
	apiHandlers "goexcel/internal/handlers"
)

type Server struct {
	config           *config.Config
	db               *database.DB
	router           *mux.Router
	proyectoHandler  *apiHandlers.ProyectoHandler
	authHandler      *apiHandlers.AuthHandler
	adminHandler     *apiHandlers.AdminHandler
	multiTenantHandler *apiHandlers.ProyectoMultiTenantHandler
	jwtService       *auth.JWTService
	authMiddleware   *auth.AuthMiddleware
}

func NewServer(cfg *config.Config, db *database.DB) *Server {
	// Inicializar repositorios
	usuarioRepo := repositories.NewUsuarioRepository(db)
	organizacionRepo := repositories.NewOrganizacionRepository(db)
	proyectoRepo := repositories.NewProyectoRepository(db)

	// Inicializar servicios de auth
	jwtService := auth.NewJWTService(cfg.JWT.Secret, "PresupuestosAI")
	authMiddleware := auth.NewAuthMiddleware(jwtService)

	// Inicializar handlers
	server := &Server{
		config:             cfg,
		db:                 db,
		router:             mux.NewRouter(),
		proyectoHandler:    apiHandlers.NewProyectoHandler(db, cfg),
		authHandler:        apiHandlers.NewAuthHandler(usuarioRepo, jwtService),
		adminHandler:       apiHandlers.NewAdminHandler(usuarioRepo, organizacionRepo, proyectoRepo),
		multiTenantHandler: apiHandlers.NewProyectoMultiTenantHandler(proyectoRepo),
		jwtService:         jwtService,
		authMiddleware:     authMiddleware,
	}

	server.setupRoutes()
	server.setupMiddleware()

	return server
}

func (s *Server) setupRoutes() {
	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", s.healthCheck).Methods("GET")

	// Auth routes (públicas)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", s.authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", s.authHandler.Login).Methods("POST")
	auth.HandleFunc("/logout", s.authHandler.Logout).Methods("POST")

	// Auth routes (requieren autenticación)
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(s.middlewareAdapter(s.authMiddleware.RequireAuth))
	authProtected.HandleFunc("/refresh", s.authHandler.RefreshToken).Methods("POST")
	authProtected.HandleFunc("/profile", s.authHandler.GetProfile).Methods("GET")
	authProtected.HandleFunc("/profile", s.authHandler.UpdateProfile).Methods("PUT")
	authProtected.HandleFunc("/change-password", s.authHandler.ChangePassword).Methods("POST")

	// Public project routes (no auth required)
	publicProjects := api.PathPrefix("/public").Subrouter()
	publicProjects.Use(s.middlewareAdapter(s.authMiddleware.OptionalAuth)) // Auth opcional para likes
	publicProjects.HandleFunc("/projects", s.multiTenantHandler.GetProyectosPublicos).Methods("GET")
	publicProjects.HandleFunc("/projects/featured", s.multiTenantHandler.GetProyectosDestacados).Methods("GET")
	publicProjects.HandleFunc("/projects/{id}", s.multiTenantHandler.GetProjectWithLikeStatus).Methods("GET")
	publicProjects.HandleFunc("/projects/{id}/details", s.multiTenantHandler.GetPublicProjectDetails).Methods("GET")

	// User project routes (require auth)
	userProjects := api.PathPrefix("/my").Subrouter()
	userProjects.Use(s.middlewareAdapter(s.authMiddleware.RequireAuth))
	userProjects.HandleFunc("/projects", s.multiTenantHandler.GetMisProyectos).Methods("GET")
	userProjects.HandleFunc("/projects", s.proyectoHandler.CreateProject).Methods("POST")
	userProjects.HandleFunc("/projects/{id}/visibility", s.multiTenantHandler.UpdateProyectoVisibility).Methods("PUT")
	userProjects.HandleFunc("/projects/{id}/like", s.multiTenantHandler.ToggleLikeProject).Methods("POST")

	// Legacy project routes (protected, para compatibilidad)
	projects := api.PathPrefix("/projects").Subrouter()
	projects.Use(s.middlewareAdapter(s.authMiddleware.RequireAuth))
	projects.HandleFunc("", s.proyectoHandler.GetProjects).Methods("GET")
	projects.HandleFunc("", s.proyectoHandler.CreateProject).Methods("POST")
	projects.HandleFunc("/{id}", s.proyectoHandler.GetProject).Methods("GET")
	projects.HandleFunc("/{id}", s.proyectoHandler.UpdateProject).Methods("PUT")
	projects.HandleFunc("/{id}", s.proyectoHandler.DeleteProject).Methods("DELETE")
	projects.HandleFunc("/{id}/export", s.proyectoHandler.ExportProject).Methods("GET")
	projects.HandleFunc("/{id}/acu", s.proyectoHandler.GetProjectACU).Methods("GET")
	projects.HandleFunc("/{id}/hierarchy", s.proyectoHandler.GetProjectHierarchy).Methods("GET")
	projects.HandleFunc("/{id}/titles", s.proyectoHandler.GetProjectTitles).Methods("GET")
	projects.HandleFunc("/{id}/titles", s.proyectoHandler.UpdateProjectTitles).Methods("PUT")

	// Admin routes (require admin role)
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(s.middlewareAdapter(s.authMiddleware.RequireRole("admin")))
	admin.HandleFunc("/dashboard", s.adminHandler.GetDashboardStats).Methods("GET")
	admin.HandleFunc("/users", s.adminHandler.GetAllUsuarios).Methods("GET")
	admin.HandleFunc("/users/{id}", s.adminHandler.GetUsuario).Methods("GET")
	admin.HandleFunc("/users/{id}/deactivate", s.adminHandler.DeactivateUsuario).Methods("POST")
	admin.HandleFunc("/projects", s.adminHandler.GetAllProyectos).Methods("GET")
	admin.HandleFunc("/projects/{id}/visibility", s.adminHandler.UpdateProyectoVisibility).Methods("PUT")
	admin.HandleFunc("/projects/featured", s.adminHandler.GetFeaturedProyectos).Methods("GET")
	admin.HandleFunc("/organizations", s.adminHandler.GetAllOrganizaciones).Methods("GET")

	// ACU validation (public)
	api.HandleFunc("/validate-acu", s.proyectoHandler.ValidateACU).Methods("POST")

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

// middlewareAdapter adapta nuestros middlewares para trabajar con mux.MiddlewareFunc
func (s *Server) middlewareAdapter(middleware func(http.HandlerFunc) http.HandlerFunc) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return middleware(next.ServeHTTP)
	}
}