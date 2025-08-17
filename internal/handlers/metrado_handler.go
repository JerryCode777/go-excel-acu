package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"goexcel/internal/database/repositories"
	"goexcel/internal/models"
)

// MetradoHandler maneja las peticiones HTTP relacionadas con metrados
type MetradoHandler struct {
	metradoRepo *repositories.MetradoRepository
}

// NewMetradoHandler crea una nueva instancia del handler de metrados
func NewMetradoHandler(metradoRepo *repositories.MetradoRepository) *MetradoHandler {
	return &MetradoHandler{
		metradoRepo: metradoRepo,
	}
}

// CrearMetrado crea o actualiza un metrado específico
func (h *MetradoHandler) CrearMetrado(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	var metradoReq models.MetradoRequest
	if err := json.NewDecoder(r.Body).Decode(&metradoReq); err != nil {
		http.Error(w, fmt.Sprintf("Error decodificando JSON: %v", err), http.StatusBadRequest)
		return
	}

	metrado, err := h.metradoRepo.CrearMetrado(proyectoID, metradoReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creando metrado: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MetradoResponse{
		Success: true,
		Message: "Metrado creado/actualizado exitosamente",
		Data:    metrado,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ActualizarMetradosLote actualiza múltiples metrados en una sola petición
func (h *MetradoHandler) ActualizarMetradosLote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	var metradosReq models.MetradosLoteRequest
	if err := json.NewDecoder(r.Body).Decode(&metradosReq); err != nil {
		http.Error(w, fmt.Sprintf("Error decodificando JSON: %v", err), http.StatusBadRequest)
		return
	}

	err = h.metradoRepo.ActualizarMetrados(proyectoID, metradosReq.Metrados)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error actualizando metrados: %v", err), http.StatusInternalServerError)
		return
	}

	// Obtener metrados actualizados
	metrados, err := h.metradoRepo.ObtenerMetradosPorProyecto(proyectoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obteniendo metrados actualizados: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MetradoResponse{
		Success:  true,
		Message:  fmt.Sprintf("%d metrados actualizados exitosamente", len(metradosReq.Metrados)),
		Metrados: metrados,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerMetradosPorProyecto obtiene todos los metrados de un proyecto
func (h *MetradoHandler) ObtenerMetradosPorProyecto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	metrados, err := h.metradoRepo.ObtenerMetradosPorProyecto(proyectoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obteniendo metrados: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MetradoResponse{
		Success:  true,
		Message:  fmt.Sprintf("%d metrados encontrados", len(metrados)),
		Metrados: metrados,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerMetradoSimple obtiene metrados en formato simple (mapa clave-valor)
func (h *MetradoHandler) ObtenerMetradoSimple(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	metrados, err := h.metradoRepo.ObtenerMetradosSimples(proyectoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obteniendo metrados: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("%d metrados encontrados", len(metrados)),
		"metrados": metrados,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerMetradoPorPartida obtiene el metrado de una partida específica
func (h *MetradoHandler) ObtenerMetradoPorPartida(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]
	partidaCodigo := vars["partida_codigo"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	metrado, err := h.metradoRepo.ObtenerMetradoPorPartida(proyectoID, partidaCodigo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obteniendo metrado: %v", err), http.StatusNotFound)
		return
	}

	response := models.MetradoResponse{
		Success: true,
		Message: "Metrado encontrado",
		Data:    metrado,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EliminarMetrado elimina un metrado específico
func (h *MetradoHandler) EliminarMetrado(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]
	partidaCodigo := vars["partida_codigo"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	err = h.metradoRepo.EliminarMetrado(proyectoID, partidaCodigo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error eliminando metrado: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MetradoResponse{
		Success: true,
		Message: "Metrado eliminado exitosamente",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ObtenerResumenProyecto obtiene el resumen financiero de un proyecto
func (h *MetradoHandler) ObtenerResumenProyecto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	resumen, err := h.metradoRepo.ObtenerResumenProyecto(proyectoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obteniendo resumen: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.ResumenProyectoResponse{
		Success: true,
		Message: "Resumen del proyecto obtenido exitosamente",
		Data:    resumen,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CalcularCostoTotalProyecto calcula el costo total del proyecto
func (h *MetradoHandler) CalcularCostoTotalProyecto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proyectoIDStr := vars["proyecto_id"]

	proyectoID, err := uuid.Parse(proyectoIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID de proyecto inválido: %v", err), http.StatusBadRequest)
		return
	}

	costoTotal, err := h.metradoRepo.CalcularCostoTotalProyecto(proyectoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calculando costo total: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Costo total calculado exitosamente",
		"costo_total": costoTotal,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}