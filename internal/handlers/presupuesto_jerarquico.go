package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	
	"goexcel/internal/database/repositories"
	"goexcel/internal/models"
)

// PresupuestoJerarquicoHandler maneja las peticiones HTTP para presupuestos jerárquicos
type PresupuestoJerarquicoHandler struct {
	repo *repositories.PresupuestoRepository
}

// NewPresupuestoJerarquicoHandler crea una nueva instancia del handler
func NewPresupuestoJerarquicoHandler(repo *repositories.PresupuestoRepository) *PresupuestoJerarquicoHandler {
	return &PresupuestoJerarquicoHandler{repo: repo}
}

// CrearPresupuesto crea un nuevo presupuesto jerárquico
func (h *PresupuestoJerarquicoHandler) CrearPresupuesto(w http.ResponseWriter, r *http.Request) {
	var req models.PresupuestoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	presupuesto, err := h.repo.CrearPresupuesto(req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.PresupuestoResponse{
		Success: true,
		Message: "Presupuesto creado exitosamente",
		Data:    presupuesto,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerPresupuesto obtiene un presupuesto por ID
func (h *PresupuestoJerarquicoHandler) ObtenerPresupuesto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	presupuesto, err := h.repo.ObtenerPresupuesto(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := models.PresupuestoResponse{
		Success:     true,
		Presupuesto: presupuesto,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CrearSubpresupuesto crea un nuevo subpresupuesto
func (h *PresupuestoJerarquicoHandler) CrearSubpresupuesto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presupuestoIDStr := vars["presupuesto_id"]

	presupuestoID, err := uuid.Parse(presupuestoIDStr)
	if err != nil {
		http.Error(w, "Invalid presupuesto ID format", http.StatusBadRequest)
		return
	}

	var req models.SubpresupuestoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	subpresupuesto, err := h.repo.CrearSubpresupuesto(presupuestoID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Subpresupuesto creado exitosamente",
		"data":    subpresupuesto,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CrearTitulo crea un nuevo título jerárquico
func (h *PresupuestoJerarquicoHandler) CrearTitulo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presupuestoIDStr := vars["presupuesto_id"]

	presupuestoID, err := uuid.Parse(presupuestoIDStr)
	if err != nil {
		http.Error(w, "Invalid presupuesto ID format", http.StatusBadRequest)
		return
	}

	var req models.TituloRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	titulo, err := h.repo.CrearTitulo(presupuestoID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.TituloResponse{
		Success: true,
		Message: "Título creado exitosamente",
		Data:    titulo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CrearPartidaJerarquica crea una nueva partida en la estructura jerárquica
func (h *PresupuestoJerarquicoHandler) CrearPartidaJerarquica(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presupuestoIDStr := vars["presupuesto_id"]

	presupuestoID, err := uuid.Parse(presupuestoIDStr)
	if err != nil {
		http.Error(w, "Invalid presupuesto ID format", http.StatusBadRequest)
		return
	}

	var req models.PartidaJerarquicaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	partida, err := h.repo.CrearPartidaJerarquica(presupuestoID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.PartidaJerarquicaResponse{
		Success: true,
		Message: "Partida creada exitosamente",
		Data:    partida,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerEstructuraJerarquica obtiene la estructura jerárquica completa
func (h *PresupuestoJerarquicoHandler) ObtenerEstructuraJerarquica(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presupuestoIDStr := vars["presupuesto_id"]

	presupuestoID, err := uuid.Parse(presupuestoIDStr)
	if err != nil {
		http.Error(w, "Invalid presupuesto ID format", http.StatusBadRequest)
		return
	}

	estructura, err := h.repo.ObtenerEstructuraJerarquica(presupuestoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"estructura": estructura,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerPartidasJerarquicas obtiene todas las partidas con información jerárquica
func (h *PresupuestoJerarquicoHandler) ObtenerPartidasJerarquicas(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presupuestoIDStr := vars["presupuesto_id"]

	presupuestoID, err := uuid.Parse(presupuestoIDStr)
	if err != nil {
		http.Error(w, "Invalid presupuesto ID format", http.StatusBadRequest)
		return
	}

	partidas, err := h.repo.ObtenerPartidasJerarquicas(presupuestoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.PartidaJerarquicaResponse{
		Success:  true,
		Partidas: partidas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerResumenJerarquico obtiene las estadísticas del presupuesto
func (h *PresupuestoJerarquicoHandler) ObtenerResumenJerarquico(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presupuestoIDStr := vars["presupuesto_id"]

	presupuestoID, err := uuid.Parse(presupuestoIDStr)
	if err != nil {
		http.Error(w, "Invalid presupuesto ID format", http.StatusBadRequest)
		return
	}

	resumen, err := h.repo.ObtenerResumenJerarquico(presupuestoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    resumen,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ProcesarACUJerarquico procesa un archivo ACU en formato jerárquico
func (h *PresupuestoJerarquicoHandler) ProcesarACUJerarquico(w http.ResponseWriter, r *http.Request) {
	// Obtener el contenido ACU del request
	var request struct {
		Contenido string `json:"contenido"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implementar el procesamiento del ACU jerárquico
	// Esto incluiría:
	// 1. Parsear el contenido ACU con el ACUJerarquicoParser
	// 2. Crear el presupuesto y su estructura jerárquica en la base de datos
	// 3. Retornar el ID del presupuesto creado

	response := map[string]interface{}{
		"success": true,
		"message": "ACU jerárquico procesado exitosamente (implementación pendiente)",
		"data":    nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetupPresupuestoJerarquicoRoutes configura las rutas para presupuestos jerárquicos
func SetupPresupuestoJerarquicoRoutes(router *mux.Router, handler *PresupuestoJerarquicoHandler) {
	// Rutas para presupuestos
	router.HandleFunc("/api/v1/presupuestos", handler.CrearPresupuesto).Methods("POST")
	router.HandleFunc("/api/v1/presupuestos/{id}", handler.ObtenerPresupuesto).Methods("GET")
	
	// Rutas para subpresupuestos
	router.HandleFunc("/api/v1/presupuestos/{presupuesto_id}/subpresupuestos", handler.CrearSubpresupuesto).Methods("POST")
	
	// Rutas para títulos
	router.HandleFunc("/api/v1/presupuestos/{presupuesto_id}/titulos", handler.CrearTitulo).Methods("POST")
	
	// Rutas para partidas jerárquicas
	router.HandleFunc("/api/v1/presupuestos/{presupuesto_id}/partidas", handler.CrearPartidaJerarquica).Methods("POST")
	
	// Rutas para obtener datos jerárquicos
	router.HandleFunc("/api/v1/presupuestos/{presupuesto_id}/estructura", handler.ObtenerEstructuraJerarquica).Methods("GET")
	router.HandleFunc("/api/v1/presupuestos/{presupuesto_id}/partidas", handler.ObtenerPartidasJerarquicas).Methods("GET")
	router.HandleFunc("/api/v1/presupuestos/{presupuesto_id}/resumen", handler.ObtenerResumenJerarquico).Methods("GET")
	
	// Ruta para procesar ACU jerárquico
	router.HandleFunc("/api/v1/presupuestos/procesar-acu", handler.ProcesarACUJerarquico).Methods("POST")
}