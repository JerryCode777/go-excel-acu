package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"goexcel/internal/database/repositories"
	"goexcel/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AdminHandler struct {
	usuarioRepo      *repositories.UsuarioRepository
	organizacionRepo *repositories.OrganizacionRepository
	proyectoRepo     *repositories.ProyectoRepository
}

func NewAdminHandler(
	usuarioRepo *repositories.UsuarioRepository,
	organizacionRepo *repositories.OrganizacionRepository,
	proyectoRepo *repositories.ProyectoRepository,
) *AdminHandler {
	return &AdminHandler{
		usuarioRepo:      usuarioRepo,
		organizacionRepo: organizacionRepo,
		proyectoRepo:     proyectoRepo,
	}
}

// Dashboard stats
func (h *AdminHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Por ahora estadísticas básicas, se puede expandir
	stats := map[string]interface{}{
		"message": "Dashboard de administración",
		"stats": map[string]int{
			"usuarios":      0, // Se implementará con consultas específicas
			"proyectos":     0,
			"organizaciones": 0,
		},
	}

	h.respondWithJSON(w, http.StatusOK, stats)
}

// Gestión de usuarios
func (h *AdminHandler) GetAllUsuarios(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	
	if limit <= 0 {
		limit = 50
	}

	usuarios, err := h.usuarioRepo.GetAll(limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error obteniendo usuarios")
		return
	}

	// Limpiar passwords
	for i := range usuarios {
		usuarios[i].PasswordHash = ""
	}

	h.respondWithJSON(w, http.StatusOK, usuarios)
}

func (h *AdminHandler) GetUsuario(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	usuario, err := h.usuarioRepo.GetByID(userID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	usuario.PasswordHash = ""
	h.respondWithJSON(w, http.StatusOK, usuario)
}

func (h *AdminHandler) DeactivateUsuario(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	if err := h.usuarioRepo.Delete(userID); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error desactivando usuario")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Usuario desactivado"})
}

// Gestión de proyectos destacados
func (h *AdminHandler) GetAllProyectos(w http.ResponseWriter, r *http.Request) {
	// Obtener todos los proyectos públicos y privados (usar método existente por ahora)
	limit := 1000 // Obtener todos los proyectos
	offset := 0
	proyectos, err := h.proyectoRepo.GetPublicos(limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error obteniendo proyectos")
		return
	}

	// Convertir a response format
	var projectResponses []models.ProyectoResponse
	for _, p := range proyectos {
		descripcion := ""
		if p.Descripcion != nil {
			descripcion = *p.Descripcion
		}
		
		projectResponse := models.ProyectoResponse{
			ID:          p.ID.String(),
			Nombre:      p.Nombre,
			Descripcion: descripcion,
			Moneda:      p.Moneda,
			Visibility:  p.Visibility,
			LikesCount:  p.LikesCount,
			VistasCount: p.VistasCount,
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

	response := map[string]interface{}{
		"success":  true,
		"projects": projectResponses,
		"total":    len(projectResponses),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *AdminHandler) UpdateProyectoVisibility(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "ID de proyecto inválido")
		return
	}

	var req struct {
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Validar visibilidad
	if req.Visibility != "private" && req.Visibility != "public" && req.Visibility != "featured" {
		h.respondWithError(w, http.StatusBadRequest, "Visibilidad debe ser: private, public o featured")
		return
	}

	// Admin puede actualizar cualquier proyecto (sin usuarioID)
	if err := h.proyectoRepo.UpdateVisibility(proyectoID, req.Visibility, nil); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error actualizando visibilidad")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Visibilidad actualizada"})
}

func (h *AdminHandler) GetFeaturedProyectos(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	proyectos, err := h.proyectoRepo.GetFeatured(limit)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error obteniendo proyectos destacados")
		return
	}

	h.respondWithJSON(w, http.StatusOK, proyectos)
}

// Gestión de organizaciones
func (h *AdminHandler) GetAllOrganizaciones(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	
	if limit <= 0 {
		limit = 50
	}

	organizaciones, err := h.organizacionRepo.GetAll(limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error obteniendo organizaciones")
		return
	}

	h.respondWithJSON(w, http.StatusOK, organizaciones)
}

// Helper methods
func (h *AdminHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AdminHandler) respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}