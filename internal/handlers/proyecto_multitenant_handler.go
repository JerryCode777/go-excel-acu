package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"goexcel/internal/auth"
	"goexcel/internal/database/repositories"
	"goexcel/internal/models"
)

type ProyectoMultiTenantHandler struct {
	proyectoRepo *repositories.ProyectoRepository
}

func NewProyectoMultiTenantHandler(proyectoRepo *repositories.ProyectoRepository) *ProyectoMultiTenantHandler {
	return &ProyectoMultiTenantHandler{
		proyectoRepo: proyectoRepo,
	}
}

// GetMisProyectos returns projects owned by the authenticated user
func (h *ProyectoMultiTenantHandler) GetMisProyectos(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	
	if limit <= 0 {
		limit = 20
	}

	log.Printf("📋 Obteniendo proyectos del usuario: %s", user.Email)

	projects, err := h.proyectoRepo.GetByUsuario(user.ID, limit, offset)
	if err != nil {
		log.Printf("❌ Error obteniendo proyectos del usuario: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching projects: %v", err), http.StatusInternalServerError)
		return
	}

	// Convertir a response format
	var projectResponses []models.ProyectoResponse
	for _, p := range projects {
		descripcion := ""
		if p.Descripcion != nil {
			descripcion = *p.Descripcion
		}
		
		projectResponses = append(projectResponses, models.ProyectoResponse{
			ID:          p.ID.String(),
			Nombre:      p.Nombre,
			Descripcion: descripcion,
			Moneda:      p.Moneda,
			Visibility:  p.Visibility,
			LikesCount:  p.LikesCount,
			VistasCount: p.VistasCount,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response := ProjectListResponse{
		Success:  true,
		Projects: projectResponses,
		Total:    len(projectResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProyectosPublicos returns public and featured projects
func (h *ProyectoMultiTenantHandler) GetProyectosPublicos(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	
	if limit <= 0 {
		limit = 20
	}

	log.Printf("📋 Obteniendo proyectos públicos")

	projects, err := h.proyectoRepo.GetPublicos(limit, offset)
	if err != nil {
		log.Printf("❌ Error obteniendo proyectos públicos: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching public projects: %v", err), http.StatusInternalServerError)
		return
	}

	// Convertir a response format
	var projectResponses []models.ProyectoResponse
	for _, p := range projects {
		descripcion := ""
		if p.Descripcion != nil {
			descripcion = *p.Descripcion
		}
		
		projectResponse := models.ProyectoResponse{
			ID:          p.ID.String(),
			Nombre:      p.Nombre,
			Descripcion: descripcion,
			Moneda:      p.Moneda,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// Agregar información del usuario si existe
		if p.Usuario != nil {
			projectResponse.Usuario = &models.UsuarioResponse{
				ID:     p.Usuario.ID.String(),
				Nombre: p.Usuario.Nombre,
				Email:  p.Usuario.Email,
			}
		}

		projectResponses = append(projectResponses, projectResponse)
	}

	responseData := ProjectListResponse{
		Success:  true,
		Projects: projectResponses,
		Total:    len(projectResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// GetProyectosDestacados returns featured projects for homepage
func (h *ProyectoMultiTenantHandler) GetProyectosDestacados(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 6 // Para mostrar en homepage
	}

	log.Printf("📋 Obteniendo proyectos destacados")

	projects, err := h.proyectoRepo.GetFeatured(limit)
	if err != nil {
		log.Printf("❌ Error obteniendo proyectos destacados: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching featured projects: %v", err), http.StatusInternalServerError)
		return
	}

	// Convertir a response format
	var projectResponses []models.ProyectoResponse
	for _, p := range projects {
		descripcion := ""
		if p.Descripcion != nil {
			descripcion = *p.Descripcion
		}
		
		projectResponse := models.ProyectoResponse{
			ID:          p.ID.String(),
			Nombre:      p.Nombre,
			Descripcion: descripcion,
			Moneda:      p.Moneda,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// Agregar información del usuario si existe
		if p.Usuario != nil {
			projectResponse.Usuario = &models.UsuarioResponse{
				ID:     p.Usuario.ID.String(),
				Nombre: p.Usuario.Nombre,
				Email:  p.Usuario.Email,
			}
		}

		projectResponses = append(projectResponses, projectResponse)
	}

	responseData := ProjectListResponse{
		Success:  true,
		Projects: projectResponses,
		Total:    len(projectResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// ToggleLikeProject toggles like status for a project
func (h *ProyectoMultiTenantHandler) ToggleLikeProject(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID := vars["id"]

	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	log.Printf("❤️ Toggle like para proyecto %s por usuario %s", projectID, user.Email)

	liked, err := h.proyectoRepo.ToggleLike(proyectoUUID, user.ID)
	if err != nil {
		log.Printf("❌ Error toggle like: %v", err)
		http.Error(w, fmt.Sprintf("Error updating like: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"liked":   liked,
		"message": func() string {
			if liked {
				return "Like agregado"
			}
			return "Like removido"
		}(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProyectoVisibility updates project visibility (user can only update their own projects)
func (h *ProyectoMultiTenantHandler) UpdateProyectoVisibility(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID := vars["id"]

	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	var req struct {
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Validar visibilidad (usuarios no pueden marcar como featured, solo admin)
	if req.Visibility != "private" && req.Visibility != "public" {
		http.Error(w, "Visibilidad debe ser: private o public", http.StatusBadRequest)
		return
	}

	log.Printf("👁️ Usuario %s actualizando visibilidad de proyecto %s a %s", user.Email, projectID, req.Visibility)

	// Usuario normal solo puede cambiar sus propios proyectos
	if err := h.proyectoRepo.UpdateVisibility(proyectoUUID, req.Visibility, &user.ID); err != nil {
		log.Printf("❌ Error actualizando visibilidad: %v", err)
		http.Error(w, fmt.Sprintf("Error updating visibility: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Visibilidad actualizada",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProjectWithLikeStatus returns a project with like status for the authenticated user
func (h *ProyectoMultiTenantHandler) GetProjectWithLikeStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	// Usuario opcional (puede estar autenticado o no)
	user := auth.GetUserFromContext(r.Context())

	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if user != nil {
		userID = &user.ID
	}

	proyecto, err := h.proyectoRepo.GetByIDWithLikeStatus(proyectoUUID, userID)
	if err != nil {
		log.Printf("❌ Error obteniendo proyecto: %v", err)
		http.Error(w, "Proyecto no encontrado", http.StatusNotFound)
		return
	}

	// Incrementar vistas
	h.proyectoRepo.IncrementViews(proyectoUUID)

	response := map[string]interface{}{
		"success": true,
		"proyecto": map[string]interface{}{
			"id":          proyecto.ID.String(),
			"nombre":      proyecto.Nombre,
			"descripcion": func() string {
				if proyecto.Descripcion != nil {
					return *proyecto.Descripcion
				}
				return ""
			}(),
			"moneda":       proyecto.Moneda,
			"visibility":   proyecto.Visibility,
			"likes_count":  proyecto.LikesCount,
			"vistas_count": proyecto.VistasCount,
			"is_liked":     proyecto.IsLiked,
			"created_at":   proyecto.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   proyecto.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	// Agregar información del usuario propietario si existe
	if proyecto.Usuario != nil {
		response["usuario"] = map[string]interface{}{
			"id":     proyecto.Usuario.ID.String(),
			"nombre": proyecto.Usuario.Nombre,
			"email":  proyecto.Usuario.Email,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPublicProjectDetails returns complete details of a public project including partidas
func (h *ProyectoMultiTenantHandler) GetPublicProjectDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	// Verificar que el proyecto sea público o destacado
	proyecto, err := h.proyectoRepo.GetByID(proyectoUUID)
	if err != nil {
		log.Printf("❌ Error obteniendo proyecto: %v", err)
		http.Error(w, "Proyecto no encontrado", http.StatusNotFound)
		return
	}

	// Solo permitir acceso a proyectos públicos o destacados
	if proyecto.Visibility != "public" && proyecto.Visibility != "featured" {
		http.Error(w, "Proyecto no público", http.StatusForbidden)
		return
	}

	// Incrementar contador de vistas
	if err := h.proyectoRepo.IncrementViews(proyectoUUID); err != nil {
		log.Printf("⚠️ Error incrementando vistas: %v", err)
	}

	// Por ahora, devolver un array vacío de partidas
	// TODO: Implementar obtención real de partidas desde la base de datos
	partidas := []map[string]interface{}{}

	// Crear respuesta con formato similar al ProjectDetailResponse
	response := map[string]interface{}{
		"success": true,
		"project": map[string]interface{}{
			"id":           proyecto.ID.String(),
			"nombre":       proyecto.Nombre,
			"descripcion":  func() string {
				if proyecto.Descripcion != nil {
					return *proyecto.Descripcion
				}
				return ""
			}(),
			"moneda":       proyecto.Moneda,
			"visibility":   proyecto.Visibility,
			"likes_count":  proyecto.LikesCount,
			"vistas_count": proyecto.VistasCount,
			"created_at":   proyecto.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   proyecto.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
		"partidas": partidas,
		"stats": map[string]interface{}{
			"total_partidas": len(partidas),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Estructuras de respuesta compartidas con proyecto_handler.go