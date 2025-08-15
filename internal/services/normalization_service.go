package services

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"goexcel/internal/legacy"
	"goexcel/internal/models"
)

type NormalizationService struct{}

func NewNormalizationService() *NormalizationService {
	return &NormalizationService{}
}

func (s *NormalizationService) NormalizeFromJSON(archivoJSON string) (*models.NormalizedData, error) {
	// Leer archivo JSON original
	data, err := os.ReadFile(archivoJSON)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}

	var partidasJSON []legacy.PartidaLegacy
	if err := json.Unmarshal(data, &partidasJSON); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}

	fmt.Printf("📊 Normalizando %d partidas del archivo %s\n", len(partidasJSON), archivoJSON)

	// Crear datos normalizados
	normalized := &models.NormalizedData{}

	// 1. Crear proyecto
	proyectoID := uuid.New().String()
	normalized.Proyecto = models.ProyectoNormalizado{
		ID:          proyectoID,
		Nombre:      fmt.Sprintf("Proyecto ACU - %s", archivoJSON),
		Descripcion: fmt.Sprintf("Proyecto normalizado desde %s", archivoJSON),
		Moneda:      "PEN",
	}

	// 2. Mapas para evitar duplicados
	recursosMap := make(map[string]models.RecursoNormalizado)
	partidasMap := make(map[string]models.PartidaNormalizada)
	var relaciones []models.RelacionNormalizada

	// 3. Procesar cada partida del JSON original
	for _, partidaJSON := range partidasJSON {
		if partidaJSON.Codigo == "" || partidaJSON.Descripcion == "" {
			continue
		}

		// Crear partida normalizada
		partidaID := uuid.New().String()
		partida := models.PartidaNormalizada{
			ID:          partidaID,
			ProyectoID:  proyectoID,
			Codigo:      partidaJSON.Codigo,
			Descripcion: partidaJSON.Descripcion,
			Unidad:      partidaJSON.Unidad,
			Rendimiento: partidaJSON.Rendimiento,
		}

		// Verificar si la partida ya existe (por código)
		if _, exists := partidasMap[partidaJSON.Codigo]; exists {
			fmt.Printf("⚠️  Partida duplicada omitida: %s\n", partidaJSON.Codigo)
			continue
		}

		partidasMap[partidaJSON.Codigo] = partida

		// Procesar recursos de cada tipo
		s.procesarRecursos(partidaJSON.ManoObra, "mano_obra", partidaID, recursosMap, &relaciones)
		s.procesarRecursos(partidaJSON.Materiales, "materiales", partidaID, recursosMap, &relaciones)
		s.procesarRecursos(partidaJSON.Equipos, "equipos", partidaID, recursosMap, &relaciones)
		s.procesarRecursos(partidaJSON.Subcontratos, "subcontratos", partidaID, recursosMap, &relaciones)
	}

	// 4. Convertir mapas a slices
	for _, recurso := range recursosMap {
		normalized.Recursos = append(normalized.Recursos, recurso)
	}

	for _, partida := range partidasMap {
		normalized.Partidas = append(normalized.Partidas, partida)
	}

	normalized.Relaciones = relaciones

	fmt.Printf("✅ Normalización completada:\n")
	fmt.Printf("   📁 Proyecto: %s\n", normalized.Proyecto.Nombre)
	fmt.Printf("   📋 Partidas únicas: %d\n", len(normalized.Partidas))
	fmt.Printf("   🔧 Recursos únicos: %d\n", len(normalized.Recursos))
	fmt.Printf("   🔗 Relaciones: %d\n", len(normalized.Relaciones))

	return normalized, nil
}

func (s *NormalizationService) procesarRecursos(
	recursos []legacy.RecursoLegacy,
	tipoRecurso string,
	partidaID string,
	recursosMap map[string]models.RecursoNormalizado,
	relaciones *[]models.RelacionNormalizada,
) {
	for _, recursoJSON := range recursos {
		if recursoJSON.Codigo == "" || recursoJSON.Descripcion == "" {
			continue
		}

		// Crear clave única para el recurso (código + tipo)
		claveRecurso := fmt.Sprintf("%s_%s", recursoJSON.Codigo, tipoRecurso)

		// Crear recurso normalizado si no existe
		if _, exists := recursosMap[claveRecurso]; !exists {
			recursoID := uuid.New().String()
			recursosMap[claveRecurso] = models.RecursoNormalizado{
				ID:          recursoID,
				Codigo:      recursoJSON.Codigo,
				Descripcion: recursoJSON.Descripcion,
				Unidad:      recursoJSON.Unidad,
				PrecioBase:  recursoJSON.Precio,
				TipoRecurso: tipoRecurso,
			}
		}

		// Crear relación partida-recurso
		var cuadrilla *float64
		if recursoJSON.Cuadrilla > 0 {
			cuadrilla = &recursoJSON.Cuadrilla
		}

		relacion := models.RelacionNormalizada{
			ID:        uuid.New().String(),
			PartidaID: partidaID,
			RecursoID: recursosMap[claveRecurso].ID,
			Cantidad:  recursoJSON.Cantidad,
			Precio:    recursoJSON.Precio,
			Cuadrilla: cuadrilla,
		}

		*relaciones = append(*relaciones, relacion)
	}
}

func (s *NormalizationService) SaveToFile(data *models.NormalizedData, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando datos: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo: %w", err)
	}

	fmt.Printf("💾 Datos normalizados guardados en: %s\n", filename)
	return nil
}

func (s *NormalizationService) LoadFromFile(filename string) (*models.NormalizedData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo normalizado: %w", err)
	}

	var normalized models.NormalizedData
	if err := json.Unmarshal(data, &normalized); err != nil {
		return nil, fmt.Errorf("error parseando archivo normalizado: %w", err)
	}

	return &normalized, nil
}

// NormalizeFromJSONData normaliza datos que ya están en formato legacy.PartidaLegacy
func (s *NormalizationService) NormalizeFromJSONData(partidasJSON []legacy.PartidaLegacy, nombreProyecto string) (*models.NormalizedData, error) {
	fmt.Printf("📊 Normalizando %d partidas desde datos JSON en memoria\n", len(partidasJSON))

	// Crear datos normalizados
	normalized := &models.NormalizedData{}

	// 1. Crear proyecto
	proyectoID := uuid.New().String()
	normalized.Proyecto = models.ProyectoNormalizado{
		ID:          proyectoID,
		Nombre:      nombreProyecto,
		Descripcion: fmt.Sprintf("Proyecto desde datos ACU: %s", nombreProyecto),
		Moneda:      "PEN",
	}

	// 2. Mapas para evitar duplicados
	recursosMap := make(map[string]models.RecursoNormalizado)
	partidasMap := make(map[string]models.PartidaNormalizada)
	var relaciones []models.RelacionNormalizada

	// 3. Procesar cada partida
	for _, partidaJSON := range partidasJSON {
		if partidaJSON.Codigo == "" || partidaJSON.Descripcion == "" {
			continue
		}

		// Crear partida normalizada
		partidaID := uuid.New().String()
		partida := models.PartidaNormalizada{
			ID:          partidaID,
			ProyectoID:  proyectoID,
			Codigo:      partidaJSON.Codigo,
			Descripcion: partidaJSON.Descripcion,
			Unidad:      partidaJSON.Unidad,
			Rendimiento: partidaJSON.Rendimiento,
		}

		// Verificar si la partida ya existe (por código)
		if _, exists := partidasMap[partidaJSON.Codigo]; exists {
			fmt.Printf("⚠️  Partida duplicada omitida: %s\n", partidaJSON.Codigo)
			continue
		}

		partidasMap[partidaJSON.Codigo] = partida

		// Procesar recursos de cada tipo
		s.procesarRecursos(partidaJSON.ManoObra, "mano_obra", partidaID, recursosMap, &relaciones)
		s.procesarRecursos(partidaJSON.Materiales, "materiales", partidaID, recursosMap, &relaciones)
		s.procesarRecursos(partidaJSON.Equipos, "equipos", partidaID, recursosMap, &relaciones)
		s.procesarRecursos(partidaJSON.Subcontratos, "subcontratos", partidaID, recursosMap, &relaciones)
	}

	// 4. Convertir mapas a slices
	for _, recurso := range recursosMap {
		normalized.Recursos = append(normalized.Recursos, recurso)
	}

	for _, partida := range partidasMap {
		normalized.Partidas = append(normalized.Partidas, partida)
	}

	normalized.Relaciones = relaciones

	fmt.Printf("✅ Normalización desde datos completada:\n")
	fmt.Printf("   📁 Proyecto: %s\n", normalized.Proyecto.Nombre)
	fmt.Printf("   📋 Partidas únicas: %d\n", len(normalized.Partidas))
	fmt.Printf("   🔧 Recursos únicos: %d\n", len(normalized.Recursos))
	fmt.Printf("   🔗 Relaciones: %d\n", len(normalized.Relaciones))

	return normalized, nil
}