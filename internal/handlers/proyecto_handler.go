package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jerryandersonh/goexcel/config"
	"github.com/jerryandersonh/goexcel/internal/database"
	"github.com/jerryandersonh/goexcel/internal/database/repositories"
	"github.com/jerryandersonh/goexcel/internal/legacy"
	"github.com/jerryandersonh/goexcel/internal/models"
	"github.com/jerryandersonh/goexcel/internal/services"
)

type ProyectoHandler struct {
	proyectoRepo     *repositories.ProyectoRepository
	partidaRepo      *repositories.PartidaRepository
	normalizationSvc *services.NormalizationService
	migrationSvc     *services.NormalizedMigrationService
	excelSvc         *services.ExcelService
	excelJerarquicoSvc *services.ExcelJerarquicoService
	hierarchySvc     *services.HierarchyService
}

func NewProyectoHandler(db *database.DB, cfg *config.Config) *ProyectoHandler {
	return &ProyectoHandler{
		proyectoRepo:     repositories.NewProyectoRepository(db),
		partidaRepo:      repositories.NewPartidaRepository(db),
		normalizationSvc: services.NewNormalizationService(),
		migrationSvc:     services.NewNormalizedMigrationService(db),
		excelSvc:         services.NewExcelService(cfg),
		excelJerarquicoSvc: services.NewExcelJerarquicoService(cfg),
		hierarchySvc:     services.NewHierarchyService(db.DB),
	}
}

// Request/Response structures
type CreateProjectRequest struct {
	Proyecto models.ProyectoRequest `json:"proyecto"`
	Partidas []models.PartidaRequest `json:"partidas"`
}

// Almacén temporal de JSON originales por proyecto ID
var originalJSONStore = make(map[string][]legacy.PartidaLegacy)

type CreateProjectResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProjectID string `json:"project_id"`
	Project   models.ProyectoResponse `json:"project"`
}

type ProjectListResponse struct {
	Success  bool                      `json:"success"`
	Projects []models.ProyectoResponse `json:"projects"`
	Total    int                       `json:"total"`
}

type ProjectDetailResponse struct {
	Success  bool                    `json:"success"`
	Project  models.ProyectoResponse `json:"project"`
	Partidas []models.PartidaResponse `json:"partidas"`
	Stats    models.ProjectStats     `json:"stats"`
}

// CreateProject creates a new project from ACU JSON data
func (h *ProyectoHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Recibiendo solicitud para crear proyecto")
	
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validar datos
	if req.Proyecto.Nombre == "" {
		http.Error(w, "Nombre del proyecto es requerido", http.StatusBadRequest)
		return
	}

	if len(req.Partidas) == 0 {
		http.Error(w, "Al menos una partida es requerida", http.StatusBadRequest)
		return
	}

	log.Printf("📊 Procesando proyecto: %s con %d partidas", req.Proyecto.Nombre, len(req.Partidas))

	// Debug: Mostrar algunas partidas del frontend
	for i, partida := range req.Partidas {
		if i < 2 { // Solo las primeras 2 para no saturar logs
			log.Printf("🔍 Partida %d: %s - %s (MO:%d, Mat:%d, Eq:%d, Sub:%d)", 
				i+1, partida.Codigo, partida.Descripcion,
				len(partida.ManoObra), len(partida.Materiales), 
				len(partida.Equipos), len(partida.Subcontratos))
		}
	}

	// Convertir a formato legacy para procesamiento
	partidasLegacy := h.convertToLegacyFormat(req.Partidas)
	log.Printf("🔄 Convertidas %d partidas a formato legacy", len(partidasLegacy))

	// Normalizar datos
	normalizedData, err := h.normalizationSvc.NormalizeFromJSONData(partidasLegacy, req.Proyecto.Nombre)
	if err != nil {
		log.Printf("❌ Error normalizando datos: %v", err)
		http.Error(w, fmt.Sprintf("Error normalizing data: %v", err), http.StatusInternalServerError)
		return
	}

	// Actualizar información del proyecto
	if req.Proyecto.Descripcion != "" {
		normalizedData.Proyecto.Descripcion = req.Proyecto.Descripcion
	}
	if req.Proyecto.Moneda != "" {
		normalizedData.Proyecto.Moneda = req.Proyecto.Moneda
	}

	// Migrar a PostgreSQL
	if err := h.migrationSvc.MigrateNormalizedData(normalizedData); err != nil {
		log.Printf("❌ Error migrando a PostgreSQL: %v", err)
		http.Error(w, fmt.Sprintf("Error saving to database: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Proyecto creado exitosamente: %s", normalizedData.Proyecto.ID)

	// Guardar JSON original para generación de Excel
	originalJSONStore[normalizedData.Proyecto.ID] = partidasLegacy
	log.Printf("💾 JSON original guardado para proyecto: %s (%d partidas)", normalizedData.Proyecto.ID, len(partidasLegacy))
	
	// Debug: Mostrar contenido de la primera partida legacy
	if len(partidasLegacy) > 0 {
		primera := partidasLegacy[0]
		log.Printf("🔍 Primera partida legacy: %s - %s (MO:%d, Mat:%d, Eq:%d, Sub:%d)", 
			primera.Codigo, primera.Descripcion,
			len(primera.ManoObra), len(primera.Materiales), 
			len(primera.Equipos), len(primera.Subcontratos))
	}

	// Respuesta
	response := CreateProjectResponse{
		Success:   true,
		Message:   "Proyecto creado exitosamente",
		ProjectID: normalizedData.Proyecto.ID,
		Project: models.ProyectoResponse{
			ID:          normalizedData.Proyecto.ID,
			Nombre:      normalizedData.Proyecto.Nombre,
			Descripcion: normalizedData.Proyecto.Descripcion,
			Moneda:      normalizedData.Proyecto.Moneda,
			CreatedAt:   "",
			UpdatedAt:   "",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProjects returns all projects
func (h *ProyectoHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	log.Printf("📋 Obteniendo lista de proyectos")

	projects, err := h.proyectoRepo.GetAll()
	if err != nil {
		log.Printf("❌ Error obteniendo proyectos: %v", err)
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

// GetProject returns a specific project with its partidas
func (h *ProyectoHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("📖 Obteniendo proyecto: %s", projectID)

	// Validar UUID
	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		log.Printf("❌ UUID inválido: %s", projectID)
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	// Obtener proyecto de la base de datos
	proyecto, err := h.proyectoRepo.GetByID(proyectoUUID)
	if err != nil {
		log.Printf("❌ Error obteniendo proyecto: %v", err)
		http.Error(w, "Proyecto no encontrado", http.StatusNotFound)
		return
	}

	// Verificar si tenemos JSON original guardado
	log.Printf("🔍 Buscando JSON original para proyecto: %s", projectID)
	log.Printf("🔍 Proyectos en memoria: %d", len(originalJSONStore))
	for id := range originalJSONStore {
		log.Printf("   - %s", id)
	}
	
	if partidasLegacy, exists := originalJSONStore[projectID]; exists && len(partidasLegacy) > 0 {
		log.Printf("📋 Usando JSON original guardado - %d partidas", len(partidasLegacy))
		
		// Convertir partidas legacy al formato de respuesta
		partidasResponse := h.convertLegacyToResponse(partidasLegacy)
		
		response := ProjectDetailResponse{
			Success: true,
			Project: models.ProyectoResponse{
				ID:          proyecto.ID.String(),
				Nombre:      proyecto.Nombre,
				Descripcion: func() string {
					if proyecto.Descripcion != nil {
						return *proyecto.Descripcion
					}
					return ""
				}(),
				Moneda:      proyecto.Moneda,
				CreatedAt:   proyecto.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:   proyecto.UpdatedAt.Format("2006-01-02 15:04:05"),
			},
			Partidas: partidasResponse,
			Stats: models.ProjectStats{
				TotalPartidas:     len(partidasResponse),
				TotalRecursos:     h.countTotalRecursos(partidasLegacy),
				CostoTotal:        h.calculateTotalCosto(partidasLegacy),
				CostoManoObra:     h.calculateCostoByType(partidasLegacy, "mano_obra"),
				CostoMateriales:   h.calculateCostoByType(partidasLegacy, "materiales"),
				CostoEquipos:      h.calculateCostoByType(partidasLegacy, "equipos"),
				CostoSubcontratos: h.calculateCostoByType(partidasLegacy, "subcontratos"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si no hay JSON original, obtener de la base de datos
	log.Printf("📊 Obteniendo partidas de la base de datos")
	partidasCompletas, err := h.getPartidasConRecursos(proyectoUUID)
	if err != nil {
		log.Printf("❌ Error obteniendo partidas: %v", err)
		http.Error(w, fmt.Sprintf("Error obteniendo partidas: %v", err), http.StatusInternalServerError)
		return
	}

	// Convertir a formato de respuesta
	partidasResponse := h.convertDBToResponse(partidasCompletas)
	
	response := ProjectDetailResponse{
		Success: true,
		Project: models.ProyectoResponse{
			ID:          proyecto.ID.String(),
			Nombre:      proyecto.Nombre,
			Descripcion: func() string {
				if proyecto.Descripcion != nil {
					return *proyecto.Descripcion
				}
				return ""
			}(),
			Moneda:      proyecto.Moneda,
			CreatedAt:   proyecto.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   proyecto.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
		Partidas: partidasResponse,
		Stats: models.ProjectStats{
			TotalPartidas:     len(partidasResponse),
			TotalRecursos:     h.countTotalRecursosFromDB(partidasCompletas),
			CostoTotal:        h.calculateTotalCostoDB(partidasCompletas),
			CostoManoObra:     h.calculateCostoByTypeDB(partidasCompletas, "mano_obra"),
			CostoMateriales:   h.calculateCostoByTypeDB(partidasCompletas, "materiales"),
			CostoEquipos:      h.calculateCostoByTypeDB(partidasCompletas, "equipos"),
			CostoSubcontratos: h.calculateCostoByTypeDB(partidasCompletas, "subcontratos"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteProject deletes a project
func (h *ProyectoHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("🗑️  Eliminando proyecto: %s", projectID)

	// Validar UUID
	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		log.Printf("❌ UUID inválido: %s", projectID)
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	// Verificar que el proyecto existe antes de eliminarlo
	_, err = h.proyectoRepo.GetByID(proyectoUUID)
	if err != nil {
		log.Printf("❌ Proyecto no encontrado: %s", projectID)
		http.Error(w, "Proyecto no encontrado", http.StatusNotFound)
		return
	}

	// Eliminar proyecto de la base de datos
	err = h.proyectoRepo.Delete(proyectoUUID)
	if err != nil {
		log.Printf("❌ Error eliminando proyecto: %v", err)
		http.Error(w, fmt.Sprintf("Error eliminando proyecto: %v", err), http.StatusInternalServerError)
		return
	}

	// Limpiar JSON original del store si existe
	delete(originalJSONStore, projectID)
	log.Printf("🧹 JSON original eliminado del store para proyecto: %s", projectID)

	response := map[string]interface{}{
		"success": true,
		"message": "Proyecto eliminado exitosamente",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProject updates an existing project
func (h *ProyectoHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("✏️  Actualizando proyecto: %s", projectID)

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// TODO: Implementar actualización
	response := CreateProjectResponse{
		Success:   true,
		Message:   "Proyecto actualizado exitosamente",
		ProjectID: projectID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportProject exports project to different formats
func (h *ProyectoHandler) ExportProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "excel"
	}

	log.Printf("📤 Exportando proyecto %s en formato: %s", projectID, format)

	switch format {
	case "excel":
		// Validar UUID del proyecto
		proyectoUUID, parseErr := uuid.Parse(projectID)
		if parseErr != nil {
			log.Printf("❌ UUID inválido: %s", projectID)
			http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
			return
		}
		
		// Obtener información del proyecto
		proyecto, err := h.proyectoRepo.GetByID(proyectoUUID)
		if err != nil {
			log.Printf("❌ Error obteniendo proyecto: %v", err)
			http.Error(w, "Proyecto no encontrado", http.StatusNotFound)
			return
		}

		log.Printf("📊 Generando Excel jerárquico profesional para proyecto: %s", proyecto.Nombre)

		// Usar el nuevo servicio jerárquico directamente
		filename, err := h.excelJerarquicoSvc.GenerarExcelJerarquico(proyecto, h.hierarchySvc)
		if err != nil {
			log.Printf("❌ Error generando Excel legacy: %v", err)
			http.Error(w, fmt.Sprintf("Error generando Excel: %v", err), http.StatusInternalServerError)
			return
		}

		// Enviar archivo
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("❌ Error abriendo archivo Excel: %v", err)
			http.Error(w, "Error abriendo archivo Excel", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		defer os.Remove(filename) // Limpiar archivo temporal

		// Usar el nombre del proyecto para el download
		downloadName := proyecto.Nombre + ".xlsx"
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadName))
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		
		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("❌ Error enviando archivo Excel: %v", err)
		}

		log.Printf("✅ Excel enviado exitosamente: %s", filename)
		
	case "acu":
		// TODO: Generar .acu
		w.Header().Set("Content-Disposition", "attachment; filename=proyecto.acu")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("@proyecto{...}"))
		
	case "json":
		// TODO: Generar JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "JSON export placeholder"})
		
	default:
		http.Error(w, "Formato no soportado", http.StatusBadRequest)
	}
}

// ValidateACU validates ACU syntax
func (h *ProyectoHandler) ValidateACU(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ACUContent string `json:"acu_content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("🔍 Validando sintaxis ACU")

	// Usar el parser ACU existente
	acuParser := services.NewACUParserService()
	_, err := acuParser.ParseString(req.ACUContent)

	response := map[string]interface{}{
		"valid":   err == nil,
		"message": "ACU válido",
	}

	if err != nil {
		response["valid"] = false
		response["message"] = fmt.Sprintf("Error de sintaxis: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProjectACU returns the ACU source code for a project
func (h *ProyectoHandler) GetProjectACU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("📄 Obteniendo código ACU para proyecto: %s", projectID)

	// Validar UUID
	proyectoUUID, err := uuid.Parse(projectID)
	if err != nil {
		log.Printf("❌ UUID inválido: %s", projectID)
		http.Error(w, "ID de proyecto inválido", http.StatusBadRequest)
		return
	}

	// Obtener información del proyecto
	proyecto, err := h.proyectoRepo.GetByID(proyectoUUID)
	if err != nil {
		log.Printf("❌ Error obteniendo proyecto: %v", err)
		http.Error(w, "Proyecto no encontrado", http.StatusNotFound)
		return
	}

	// Verificar si tenemos JSON original guardado
	if partidasLegacy, exists := originalJSONStore[projectID]; exists && len(partidasLegacy) > 0 {
		log.Printf("📋 Generando ACU desde JSON original - %d partidas", len(partidasLegacy))
		
		// Generar código ACU desde el JSON original
		acuContent := h.generateACUFromLegacy(proyecto, partidasLegacy)
		
		response := map[string]interface{}{
			"success":     true,
			"acu_content": acuContent,
			"source":      "original_json",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si no hay JSON original, obtener desde la base de datos
	log.Printf("📋 Generando ACU desde base de datos")
	
	partidasCompletas, err := h.getPartidasConRecursos(proyectoUUID)
	if err != nil {
		log.Printf("❌ Error obteniendo partidas de BD: %v", err)
		http.Error(w, "Error obteniendo datos del proyecto", http.StatusInternalServerError)
		return
	}

	if len(partidasCompletas) == 0 {
		log.Printf("❌ No se encontraron partidas para el proyecto: %s", projectID)
		response := map[string]interface{}{
			"success":     true,
			"acu_content": h.generateEmptyACU(proyecto),
			"source":      "database",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generar código ACU desde datos de la BD
	acuContent := h.generateACUFromDB(proyecto, partidasCompletas)
	
	response := map[string]interface{}{
		"success":     true,
		"acu_content": acuContent,
		"source":      "database",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateACUFromLegacy genera código ACU desde datos legacy (JSON original)
func (h *ProyectoHandler) generateACUFromLegacy(proyecto *models.Proyecto, partidasLegacy []legacy.PartidaLegacy) string {
	var acuContent strings.Builder
	
	// Agregar proyecto
	acuContent.WriteString(fmt.Sprintf("@proyecto{%s,\n", strings.ToLower(strings.ReplaceAll(proyecto.Nombre, " ", "_"))))
	acuContent.WriteString(fmt.Sprintf("  nombre = \"%s\",\n", proyecto.Nombre))
	if proyecto.Descripcion != nil && *proyecto.Descripcion != "" {
		acuContent.WriteString(fmt.Sprintf("  descripcion = \"%s\",\n", *proyecto.Descripcion))
	}
	acuContent.WriteString(fmt.Sprintf("  moneda = \"%s\"\n", proyecto.Moneda))
	acuContent.WriteString("}\n\n")
	
	// Agregar partidas
	for _, partida := range partidasLegacy {
		partidaID := strings.ToLower(strings.ReplaceAll(partida.Codigo, ".", "_"))
		partidaID = strings.ReplaceAll(partidaID, " ", "_")
		
		acuContent.WriteString(fmt.Sprintf("@partida{%s,\n", partidaID))
		acuContent.WriteString(fmt.Sprintf("  codigo = \"%s\",\n", partida.Codigo))
		acuContent.WriteString(fmt.Sprintf("  descripcion = \"%s\",\n", partida.Descripcion))
		acuContent.WriteString(fmt.Sprintf("  unidad = \"%s\",\n", partida.Unidad))
		acuContent.WriteString(fmt.Sprintf("  rendimiento = %.1f,\n", partida.Rendimiento))
		
		// Agregar secciones de recursos
		h.addRecursosSectionToACU(&acuContent, "mano_obra", partida.ManoObra)
		h.addRecursosSectionToACU(&acuContent, "materiales", partida.Materiales)
		h.addRecursosSectionToACU(&acuContent, "equipos", partida.Equipos)
		h.addRecursosSectionToACU(&acuContent, "subcontratos", partida.Subcontratos)
		
		acuContent.WriteString("}\n\n")
	}
	
	return acuContent.String()
}

// generateACUFromDB genera código ACU desde datos de la base de datos
func (h *ProyectoHandler) generateACUFromDB(proyecto *models.Proyecto, partidasCompletas []PartidaConRecursos) string {
	var acuContent strings.Builder
	
	// Agregar proyecto
	acuContent.WriteString(fmt.Sprintf("@proyecto{%s,\n", strings.ToLower(strings.ReplaceAll(proyecto.Nombre, " ", "_"))))
	acuContent.WriteString(fmt.Sprintf("  nombre = \"%s\",\n", proyecto.Nombre))
	if proyecto.Descripcion != nil && *proyecto.Descripcion != "" {
		acuContent.WriteString(fmt.Sprintf("  descripcion = \"%s\",\n", *proyecto.Descripcion))
	}
	acuContent.WriteString(fmt.Sprintf("  moneda = \"%s\"\n", proyecto.Moneda))
	acuContent.WriteString("}\n\n")
	
	// Agregar partidas
	for _, partida := range partidasCompletas {
		partidaID := strings.ToLower(strings.ReplaceAll(partida.Codigo, ".", "_"))
		partidaID = strings.ReplaceAll(partidaID, " ", "_")
		
		acuContent.WriteString(fmt.Sprintf("@partida{%s,\n", partidaID))
		acuContent.WriteString(fmt.Sprintf("  codigo = \"%s\",\n", partida.Codigo))
		acuContent.WriteString(fmt.Sprintf("  descripcion = \"%s\",\n", partida.Descripcion))
		acuContent.WriteString(fmt.Sprintf("  unidad = \"%s\",\n", partida.Unidad))
		acuContent.WriteString(fmt.Sprintf("  rendimiento = %.1f,\n", partida.Rendimiento))
		
		// Convertir recursos de BD a formato legacy para reutilizar función
		manoObraLegacy := h.convertRecursosCompletosToLegacy(partida.ManoObra)
		materialesLegacy := h.convertRecursosCompletosToLegacy(partida.Materiales)
		equiposLegacy := h.convertRecursosCompletosToLegacy(partida.Equipos)
		subcontratosLegacy := h.convertRecursosCompletosToLegacy(partida.Subcontratos)
		
		// Agregar secciones de recursos
		h.addRecursosSectionToACU(&acuContent, "mano_obra", manoObraLegacy)
		h.addRecursosSectionToACU(&acuContent, "materiales", materialesLegacy)
		h.addRecursosSectionToACU(&acuContent, "equipos", equiposLegacy)
		h.addRecursosSectionToACU(&acuContent, "subcontratos", subcontratosLegacy)
		
		acuContent.WriteString("}\n\n")
	}
	
	return acuContent.String()
}

// generateEmptyACU genera código ACU vacío para un proyecto sin partidas
func (h *ProyectoHandler) generateEmptyACU(proyecto *models.Proyecto) string {
	var acuContent strings.Builder
	
	acuContent.WriteString(fmt.Sprintf("@proyecto{%s,\n", strings.ToLower(strings.ReplaceAll(proyecto.Nombre, " ", "_"))))
	acuContent.WriteString(fmt.Sprintf("  nombre = \"%s\",\n", proyecto.Nombre))
	if proyecto.Descripcion != nil && *proyecto.Descripcion != "" {
		acuContent.WriteString(fmt.Sprintf("  descripcion = \"%s\",\n", *proyecto.Descripcion))
	}
	acuContent.WriteString(fmt.Sprintf("  moneda = \"%s\"\n", proyecto.Moneda))
	acuContent.WriteString("}\n\n")
	
	acuContent.WriteString("// Agrega tus partidas aquí\n")
	acuContent.WriteString("// Ejemplo:\n")
	acuContent.WriteString("// @partida{ejemplo,\n")
	acuContent.WriteString("//   codigo = \"01.01.01\",\n")
	acuContent.WriteString("//   descripcion = \"EXCAVACIÓN MANUAL\",\n")
	acuContent.WriteString("//   unidad = \"m3\",\n")
	acuContent.WriteString("//   rendimiento = 8.0,\n")
	acuContent.WriteString("//   \n")
	acuContent.WriteString("//   mano_obra = {\n")
	acuContent.WriteString("//     {codigo = \"470101\", desc = \"OPERARIO\", unidad = \"hh\", cantidad = 1.0, precio = 25.00, cuadrilla = 1.0}\n")
	acuContent.WriteString("//   }\n")
	acuContent.WriteString("// }\n")
	
	return acuContent.String()
}

// addRecursosSectionToACU agrega una sección de recursos al código ACU
func (h *ProyectoHandler) addRecursosSectionToACU(acuContent *strings.Builder, nombreSeccion string, recursos []legacy.RecursoLegacy) {
	if len(recursos) == 0 {
		return
	}
	
	acuContent.WriteString(fmt.Sprintf("  \n  %s = {\n", nombreSeccion))
	
	for _, recurso := range recursos {
		acuContent.WriteString(fmt.Sprintf("    {codigo = \"%s\", desc = \"%s\", unidad = \"%s\", cantidad = %.4g, precio = %.2f",
			recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.Cantidad, recurso.Precio))
		
		// Agregar cuadrilla solo para mano de obra y si es mayor que 0
		if nombreSeccion == "mano_obra" && recurso.Cuadrilla > 0 {
			acuContent.WriteString(fmt.Sprintf(", cuadrilla = %.4g", recurso.Cuadrilla))
		}
		
		acuContent.WriteString("},\n")
	}
	
	acuContent.WriteString("  },\n")
}

// Helper function to convert request format to legacy format
func (h *ProyectoHandler) convertToLegacyFormat(partidas []models.PartidaRequest) []legacy.PartidaLegacy {
	var result []legacy.PartidaLegacy
	
	for _, p := range partidas {
		partidaLegacy := legacy.PartidaLegacy{
			Codigo:       p.Codigo,
			Descripcion:  p.Descripcion,
			Unidad:       p.Unidad,
			Rendimiento:  p.Rendimiento,
			ManoObra:     h.convertRecursosToLegacy(p.ManoObra),
			Materiales:   h.convertRecursosToLegacy(p.Materiales),
			Equipos:      h.convertRecursosToLegacy(p.Equipos),
			Subcontratos: h.convertRecursosToLegacy(p.Subcontratos),
		}
		result = append(result, partidaLegacy)
	}
	
	return result
}

func (h *ProyectoHandler) convertRecursosToLegacy(recursos []models.RecursoRequest) []legacy.RecursoLegacy {
	var result []legacy.RecursoLegacy
	
	for _, r := range recursos {
		recursoLegacy := legacy.RecursoLegacy{
			Codigo:      r.Codigo,
			Descripcion: r.Descripcion,
			Unidad:      r.Unidad,
			Cantidad:    r.Cantidad,
			Precio:      r.Precio,
		}
		
		if r.Cuadrilla != nil {
			recursoLegacy.Cuadrilla = *r.Cuadrilla
		}
		
		result = append(result, recursoLegacy)
	}
	
	return result
}

// generateExcelLegacy genera Excel usando el nuevo servicio jerárquico con datos de la BD
func (h *ProyectoHandler) generateExcelLegacy(proyecto *models.Proyecto, proyectoUUID uuid.UUID) (string, error) {
	log.Printf("🔄 Generando Excel jerárquico profesional para proyecto: %s", proyecto.Nombre)

	// Usar el nuevo servicio jerárquico que obtiene datos directamente de la BD
	filename, err := h.excelJerarquicoSvc.GenerarExcelJerarquico(proyecto, h.hierarchySvc)
	if err != nil {
		return "", fmt.Errorf("error generando Excel jerárquico: %w", err)
	}

	log.Printf("✅ Excel jerárquico generado exitosamente: %s", filename)
	return filename, nil
}

// generateExcelFromOriginalJSON genera Excel usando el JSON original del frontend o BD jerárquica
func (h *ProyectoHandler) generateExcelFromOriginalJSON(proyecto *models.Proyecto, projectID string) (string, error) {
	// Buscar JSON original guardado
	partidasLegacy, exists := originalJSONStore[projectID]
	if !exists {
		log.Printf("⚠️ No se encontró JSON original para proyecto %s, usando método jerárquico desde BD", projectID)
		// Fallback al método jerárquico desde BD
		return h.excelJerarquicoSvc.GenerarExcelJerarquico(proyecto, h.hierarchySvc)
	}

	if len(partidasLegacy) == 0 {
		return "", fmt.Errorf("no hay partidas en el JSON original del proyecto")
	}

	log.Printf("📋 Usando JSON original con %d partidas - fallback a método jerárquico", len(partidasLegacy))

	// Por ahora, usar el servicio jerárquico hasta implementar soporte para JSON legacy
	log.Printf("⚠️ JSON legacy no soportado, usando método jerárquico desde BD")
	return h.excelJerarquicoSvc.GenerarExcelJerarquico(proyecto, h.hierarchySvc)
}

// getPartidasConRecursos obtiene partidas con todos sus recursos de la BD
func (h *ProyectoHandler) getPartidasConRecursos(proyectoUUID uuid.UUID) ([]PartidaConRecursos, error) {
	// Consulta para obtener partidas con sus recursos agrupados por tipo
	query := `
		SELECT 
			p.id, p.codigo, p.descripcion, p.unidad, p.rendimiento,
			COALESCE(json_agg(
				json_build_object(
					'codigo', r.codigo,
					'descripcion', r.descripcion,
					'unidad', r.unidad,
					'cantidad', pr.cantidad,
					'precio', pr.precio,
					'cuadrilla', pr.cuadrilla,
					'tipo', tr.nombre
				) ORDER BY r.codigo
			) FILTER (WHERE r.id IS NOT NULL), '[]') as recursos
		FROM partidas p
		LEFT JOIN partida_recursos pr ON p.id = pr.partida_id
		LEFT JOIN recursos r ON pr.recurso_id = r.id
		LEFT JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
		WHERE p.proyecto_id = $1
		GROUP BY p.id, p.codigo, p.descripcion, p.unidad, p.rendimiento
		ORDER BY p.codigo
	`

	rows, err := h.partidaRepo.GetDB().Query(query, proyectoUUID)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando consulta: %w", err)
	}
	defer rows.Close()

	var partidasConRecursos []PartidaConRecursos
	for rows.Next() {
		var partida PartidaConRecursos
		var recursosJSON string

		err := rows.Scan(
			&partida.ID, &partida.Codigo, &partida.Descripcion, 
			&partida.Unidad, &partida.Rendimiento, &recursosJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando partida: %w", err)
		}

		// Parsear JSON de recursos
		var recursosData []map[string]interface{}
		if err := json.Unmarshal([]byte(recursosJSON), &recursosData); err != nil {
			log.Printf("⚠️  Error parseando recursos JSON para partida %s: %v", partida.Codigo, err)
			continue
		}

		// Agrupar recursos por tipo
		partida.ManoObra = []RecursoCompleto{}
		partida.Materiales = []RecursoCompleto{}
		partida.Equipos = []RecursoCompleto{}
		partida.Subcontratos = []RecursoCompleto{}

		for _, recursoData := range recursosData {
			if recursoData["codigo"] == nil {
				continue
			}

			recurso := RecursoCompleto{
				Codigo:      recursoData["codigo"].(string),
				Descripcion: recursoData["descripcion"].(string),
				Unidad:      recursoData["unidad"].(string),
				Cantidad:    recursoData["cantidad"].(float64),
				Precio:      recursoData["precio"].(float64),
			}

			if cuadrilla, ok := recursoData["cuadrilla"].(float64); ok && cuadrilla > 0 {
				recurso.Cuadrilla = &cuadrilla
			}

			tipo := recursoData["tipo"].(string)
			switch tipo {
			case "mano_obra":
				partida.ManoObra = append(partida.ManoObra, recurso)
			case "materiales":
				partida.Materiales = append(partida.Materiales, recurso)
			case "equipos":
				partida.Equipos = append(partida.Equipos, recurso)
			case "subcontratos":
				partida.Subcontratos = append(partida.Subcontratos, recurso)
			}
		}

		partidasConRecursos = append(partidasConRecursos, partida)
	}

	return partidasConRecursos, nil
}

// convertToLegacyFormatFromDB convierte datos de BD al formato legacy
func (h *ProyectoHandler) convertToLegacyFormatFromDB(partidasCompletas []PartidaConRecursos) []legacy.PartidaLegacy {
	var partidasLegacy []legacy.PartidaLegacy

	for _, partida := range partidasCompletas {
		partidaLegacy := legacy.PartidaLegacy{
			Codigo:      partida.Codigo,
			Descripcion: partida.Descripcion,
			Unidad:      partida.Unidad,
			Rendimiento: partida.Rendimiento,
		}

		// Convertir recursos por tipo
		partidaLegacy.ManoObra = h.convertRecursosCompletosToLegacy(partida.ManoObra)
		partidaLegacy.Materiales = h.convertRecursosCompletosToLegacy(partida.Materiales)
		partidaLegacy.Equipos = h.convertRecursosCompletosToLegacy(partida.Equipos)
		partidaLegacy.Subcontratos = h.convertRecursosCompletosToLegacy(partida.Subcontratos)

		partidasLegacy = append(partidasLegacy, partidaLegacy)
	}

	return partidasLegacy
}

// convertRecursosCompletosToLegacy convierte recursos completos a formato legacy
func (h *ProyectoHandler) convertRecursosCompletosToLegacy(recursos []RecursoCompleto) []legacy.RecursoLegacy {
	var recursosLegacy []legacy.RecursoLegacy

	for _, recurso := range recursos {
		recursoLegacy := legacy.RecursoLegacy{
			Codigo:      recurso.Codigo,
			Descripcion: recurso.Descripcion,
			Unidad:      recurso.Unidad,
			Cantidad:    recurso.Cantidad,
			Precio:      recurso.Precio,
		}

		if recurso.Cuadrilla != nil {
			recursoLegacy.Cuadrilla = *recurso.Cuadrilla
		}

		recursosLegacy = append(recursosLegacy, recursoLegacy)
	}

	return recursosLegacy
}

// Estructuras auxiliares para obtener datos completos de la BD
type PartidaConRecursos struct {
	ID           uuid.UUID         `json:"id"`
	Codigo       string            `json:"codigo"`
	Descripcion  string            `json:"descripcion"`
	Unidad       string            `json:"unidad"`
	Rendimiento  float64           `json:"rendimiento"`
	ManoObra     []RecursoCompleto `json:"mano_obra"`
	Materiales   []RecursoCompleto `json:"materiales"`
	Equipos      []RecursoCompleto `json:"equipos"`
	Subcontratos []RecursoCompleto `json:"subcontratos"`
}

type RecursoCompleto struct {
	Codigo      string   `json:"codigo"`
	Descripcion string   `json:"descripcion"`
	Unidad      string   `json:"unidad"`
	Cantidad    float64  `json:"cantidad"`
	Precio      float64  `json:"precio"`
	Cuadrilla   *float64 `json:"cuadrilla,omitempty"`
}

// Helper functions for converting legacy data to response format
func (h *ProyectoHandler) convertLegacyToResponse(partidasLegacy []legacy.PartidaLegacy) []models.PartidaResponse {
	var partidasResponse []models.PartidaResponse
	
	for _, partida := range partidasLegacy {
		partidaResponse := models.PartidaResponse{
			ID:           uuid.New().String(), // Generate temporary ID for legacy data
			Codigo:       partida.Codigo,
			Descripcion:  partida.Descripcion,
			Unidad:       partida.Unidad,
			Rendimiento:  partida.Rendimiento,
			CostoTotal:   h.calculatePartidaCosto(partida),
			ManoObra:     h.convertLegacyRecursosToResponse(partida.ManoObra),
			Materiales:   h.convertLegacyRecursosToResponse(partida.Materiales),
			Equipos:      h.convertLegacyRecursosToResponse(partida.Equipos),
			Subcontratos: h.convertLegacyRecursosToResponse(partida.Subcontratos),
		}
		partidasResponse = append(partidasResponse, partidaResponse)
	}
	
	return partidasResponse
}

func (h *ProyectoHandler) convertLegacyRecursosToResponse(recursos []legacy.RecursoLegacy) []models.RecursoResponse {
	var recursosResponse []models.RecursoResponse
	
	for _, recurso := range recursos {
		cantidad := recurso.Cantidad
		if recurso.Cuadrilla > 0 {
			cantidad = recurso.Cuadrilla
		}
		
		recursoResponse := models.RecursoResponse{
			ID:          uuid.New().String(), // Generate temporary ID for legacy data
			Codigo:      recurso.Codigo,
			Descripcion: recurso.Descripcion,
			Unidad:      recurso.Unidad,
			Cantidad:    recurso.Cantidad,
			Precio:      recurso.Precio,
			Parcial:     cantidad * recurso.Precio,
		}
		
		if recurso.Cuadrilla > 0 {
			recursoResponse.Cuadrilla = &recurso.Cuadrilla
		}
		
		recursosResponse = append(recursosResponse, recursoResponse)
	}
	
	return recursosResponse
}

// Helper functions for converting DB data to response format
func (h *ProyectoHandler) convertDBToResponse(partidasCompletas []PartidaConRecursos) []models.PartidaResponse {
	var partidasResponse []models.PartidaResponse
	
	for _, partida := range partidasCompletas {
		partidaResponse := models.PartidaResponse{
			ID:           partida.ID.String(),
			Codigo:       partida.Codigo,
			Descripcion:  partida.Descripcion,
			Unidad:       partida.Unidad,
			Rendimiento:  partida.Rendimiento,
			CostoTotal:   h.calculatePartidaCostoDB(partida),
			ManoObra:     h.convertDBRecursosToResponse(partida.ManoObra),
			Materiales:   h.convertDBRecursosToResponse(partida.Materiales),
			Equipos:      h.convertDBRecursosToResponse(partida.Equipos),
			Subcontratos: h.convertDBRecursosToResponse(partida.Subcontratos),
		}
		partidasResponse = append(partidasResponse, partidaResponse)
	}
	
	return partidasResponse
}

func (h *ProyectoHandler) convertDBRecursosToResponse(recursos []RecursoCompleto) []models.RecursoResponse {
	var recursosResponse []models.RecursoResponse
	
	for _, recurso := range recursos {
		cantidad := recurso.Cantidad
		if recurso.Cuadrilla != nil && *recurso.Cuadrilla > 0 {
			cantidad = *recurso.Cuadrilla
		}
		
		recursoResponse := models.RecursoResponse{
			ID:          uuid.New().String(), // Generate temporary ID for DB data
			Codigo:      recurso.Codigo,
			Descripcion: recurso.Descripcion,
			Unidad:      recurso.Unidad,
			Cantidad:    recurso.Cantidad,
			Precio:      recurso.Precio,
			Parcial:     cantidad * recurso.Precio,
		}
		
		if recurso.Cuadrilla != nil && *recurso.Cuadrilla > 0 {
			recursoResponse.Cuadrilla = recurso.Cuadrilla
		}
		
		recursosResponse = append(recursosResponse, recursoResponse)
	}
	
	return recursosResponse
}

// Helper functions for calculating statistics from legacy data
func (h *ProyectoHandler) countTotalRecursos(partidasLegacy []legacy.PartidaLegacy) int {
	total := 0
	for _, partida := range partidasLegacy {
		total += len(partida.ManoObra) + len(partida.Materiales) + len(partida.Equipos) + len(partida.Subcontratos)
	}
	return total
}

func (h *ProyectoHandler) calculateTotalCosto(partidasLegacy []legacy.PartidaLegacy) float64 {
	total := 0.0
	for _, partida := range partidasLegacy {
		total += h.calculatePartidaCosto(partida)
	}
	return total
}

func (h *ProyectoHandler) calculatePartidaCosto(partida legacy.PartidaLegacy) float64 {
	total := 0.0
	
	// Mano de obra
	for _, recurso := range partida.ManoObra {
		cantidad := recurso.Cantidad
		if recurso.Cuadrilla > 0 {
			cantidad = recurso.Cuadrilla
		}
		total += cantidad * recurso.Precio * partida.Rendimiento
	}
	
	// Materiales
	for _, recurso := range partida.Materiales {
		total += recurso.Cantidad * recurso.Precio * partida.Rendimiento
	}
	
	// Equipos
	for _, recurso := range partida.Equipos {
		cantidad := recurso.Cantidad
		if recurso.Cuadrilla > 0 {
			cantidad = recurso.Cuadrilla
		}
		total += cantidad * recurso.Precio * partida.Rendimiento
	}
	
	// Subcontratos
	for _, recurso := range partida.Subcontratos {
		total += recurso.Cantidad * recurso.Precio * partida.Rendimiento
	}
	
	return total
}

func (h *ProyectoHandler) calculateCostoByType(partidasLegacy []legacy.PartidaLegacy, tipoRecurso string) float64 {
	total := 0.0
	
	for _, partida := range partidasLegacy {
		var recursos []legacy.RecursoLegacy
		
		switch tipoRecurso {
		case "mano_obra":
			recursos = partida.ManoObra
		case "materiales":
			recursos = partida.Materiales
		case "equipos":
			recursos = partida.Equipos
		case "subcontratos":
			recursos = partida.Subcontratos
		}
		
		for _, recurso := range recursos {
			cantidad := recurso.Cantidad
			if (tipoRecurso == "mano_obra" || tipoRecurso == "equipos") && recurso.Cuadrilla > 0 {
				cantidad = recurso.Cuadrilla
			}
			total += cantidad * recurso.Precio * partida.Rendimiento
		}
	}
	
	return total
}

// Helper functions for calculating statistics from DB data
func (h *ProyectoHandler) countTotalRecursosFromDB(partidasCompletas []PartidaConRecursos) int {
	total := 0
	for _, partida := range partidasCompletas {
		total += len(partida.ManoObra) + len(partida.Materiales) + len(partida.Equipos) + len(partida.Subcontratos)
	}
	return total
}

func (h *ProyectoHandler) calculateTotalCostoDB(partidasCompletas []PartidaConRecursos) float64 {
	total := 0.0
	for _, partida := range partidasCompletas {
		total += h.calculatePartidaCostoDB(partida)
	}
	return total
}

func (h *ProyectoHandler) calculatePartidaCostoDB(partida PartidaConRecursos) float64 {
	total := 0.0
	
	// Mano de obra
	for _, recurso := range partida.ManoObra {
		cantidad := recurso.Cantidad
		if recurso.Cuadrilla != nil && *recurso.Cuadrilla > 0 {
			cantidad = *recurso.Cuadrilla
		}
		total += cantidad * recurso.Precio * partida.Rendimiento
	}
	
	// Materiales
	for _, recurso := range partida.Materiales {
		total += recurso.Cantidad * recurso.Precio * partida.Rendimiento
	}
	
	// Equipos
	for _, recurso := range partida.Equipos {
		cantidad := recurso.Cantidad
		if recurso.Cuadrilla != nil && *recurso.Cuadrilla > 0 {
			cantidad = *recurso.Cuadrilla
		}
		total += cantidad * recurso.Precio * partida.Rendimiento
	}
	
	// Subcontratos
	for _, recurso := range partida.Subcontratos {
		total += recurso.Cantidad * recurso.Precio * partida.Rendimiento
	}
	
	return total
}

func (h *ProyectoHandler) calculateCostoByTypeDB(partidasCompletas []PartidaConRecursos, tipoRecurso string) float64 {
	total := 0.0
	
	for _, partida := range partidasCompletas {
		var recursos []RecursoCompleto
		
		switch tipoRecurso {
		case "mano_obra":
			recursos = partida.ManoObra
		case "materiales":
			recursos = partida.Materiales
		case "equipos":
			recursos = partida.Equipos
		case "subcontratos":
			recursos = partida.Subcontratos
		}
		
		for _, recurso := range recursos {
			cantidad := recurso.Cantidad
			if (tipoRecurso == "mano_obra" || tipoRecurso == "equipos") && recurso.Cuadrilla != nil && *recurso.Cuadrilla > 0 {
				cantidad = *recurso.Cuadrilla
			}
			total += cantidad * recurso.Precio * partida.Rendimiento
		}
	}
	
	return total
}
// GetProjectHierarchy returns the hierarchical structure of a project
func (h *ProyectoHandler) GetProjectHierarchy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("🌲 Obteniendo jerarquía del proyecto: %s", projectID)

	// Obtener jerarquía completa
	jerarquia, err := h.hierarchySvc.ObtenerJerarquiaCompleta(projectID)
	if err != nil {
		log.Printf("❌ Error obteniendo jerarquía: %v", err)
		http.Error(w, fmt.Sprintf("Error obteniendo jerarquía: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"jerarquia": jerarquia,
		"total":     len(jerarquia),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProjectTitles returns only organizational titles for a project
func (h *ProyectoHandler) GetProjectTitles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("📋 Obteniendo títulos del proyecto: %s", projectID)

	// Obtener solo títulos
	titulos, err := h.hierarchySvc.ObtenerTitulosJerarquicos(projectID)
	if err != nil {
		log.Printf("❌ Error obteniendo títulos: %v", err)
		http.Error(w, fmt.Sprintf("Error obteniendo títulos: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"titulos": titulos,
		"total":   len(titulos),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProjectTitles allows customization of organizational titles
func (h *ProyectoHandler) UpdateProjectTitles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	log.Printf("✏️ Actualizando títulos del proyecto: %s", projectID)

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Actualizar títulos personalizados
	if err := h.hierarchySvc.ActualizarTitulosPersonalizados(projectID, req); err != nil {
		log.Printf("❌ Error actualizando títulos: %v", err)
		http.Error(w, fmt.Sprintf("Error actualizando títulos: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Títulos actualizados exitosamente",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
