package services

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jerryandersonh/goexcel/internal/legacy"
	"github.com/jerryandersonh/goexcel/internal/models"
)

type ACUParserService struct{}

func NewACUParserService() *ACUParserService {
	return &ACUParserService{}
}

// ParseFile parsea un archivo .acu y devuelve el proyecto
func (s *ACUParserService) ParseFile(filename string) (*models.ACUProject, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}
	
	return s.ParseString(string(content))
}

// ParseString parsea el contenido de un archivo .acu como string
func (s *ACUParserService) ParseString(content string) (*models.ACUProject, error) {
	project := &models.ACUProject{
		ID:       uuid.New().String(),
		Partidas: []models.ACUPartida{},
	}
	
	// Parsear proyecto
	if err := s.parseProject(content, project); err != nil {
		return nil, fmt.Errorf("error parseando proyecto: %w", err)
	}
	
	// Parsear partidas
	partidas, err := s.parsePartidas(content)
	if err != nil {
		return nil, fmt.Errorf("error parseando partidas: %w", err)
	}
	
	project.Partidas = partidas
	
	return project, nil
}

// parseProject extrae información del proyecto
func (s *ACUParserService) parseProject(content string, project *models.ACUProject) error {
	// Regex para extraer el bloque @proyecto
	projectRegex := regexp.MustCompile(`@proyecto\s*\{\s*([^,]+),\s*([^}]+)\}`)
	matches := projectRegex.FindStringSubmatch(content)
	
	if len(matches) > 0 {
		// Parsear campos del proyecto
		fieldsContent := matches[1] + "," + matches[2]
		fields := s.parseFields(fieldsContent)
		
		if nombre, ok := fields["nombre"]; ok {
			project.Nombre = s.cleanQuotes(nombre)
		}
		if desc, ok := fields["descripcion"]; ok {
			project.Descripcion = s.cleanQuotes(desc)
		}
		if moneda, ok := fields["moneda"]; ok {
			project.Moneda = s.cleanQuotes(moneda)
		}
	}
	
	// Valores por defecto
	if project.Nombre == "" {
		project.Nombre = "Proyecto ACU"
	}
	if project.Moneda == "" {
		project.Moneda = "PEN"
	}
	
	return nil
}

// parsePartidas extrae todas las partidas del contenido
func (s *ACUParserService) parsePartidas(content string) ([]models.ACUPartida, error) {
	var partidas []models.ACUPartida
	
	// Regex mejorada para capturar bloques @partida completos
	partidaRegex := regexp.MustCompile(`@partida\s*\{\s*([^,]+),\s*((?:[^{}]*\{[^{}]*\}[^{}]*)*[^}]*)\}`)
	matches := partidaRegex.FindAllStringSubmatch(content, -1)
	
	fmt.Printf("🔍 Encontradas %d partidas en el archivo\n", len(matches))
	
	for i, match := range matches {
		if len(match) < 3 {
			fmt.Printf("⚠️  Match %d incompleto\n", i)
			continue
		}
		
		partidaID := strings.TrimSpace(match[1])
		partidaContent := match[2]
		
		fmt.Printf("📋 Procesando partida: %s\n", partidaID)
		
		partida, err := s.parsePartida(partidaID, partidaContent)
		if err != nil {
			fmt.Printf("⚠️  Error parseando partida %s: %v\n", partidaID, err)
			continue
		}
		
		partidas = append(partidas, *partida)
	}
	
	return partidas, nil
}

// parsePartida parsea una partida individual
func (s *ACUParserService) parsePartida(partidaID, content string) (*models.ACUPartida, error) {
	partida := &models.ACUPartida{
		ID: uuid.New().String(),
	}
	
	// Parsear campos básicos
	fields := s.parseFields(content)
	
	if codigo, ok := fields["codigo"]; ok {
		partida.Codigo = s.cleanQuotes(codigo)
	}
	if desc, ok := fields["descripcion"]; ok {
		partida.Descripcion = s.cleanQuotes(desc)
	}
	if unidad, ok := fields["unidad"]; ok {
		partida.Unidad = s.cleanQuotes(unidad)
	}
	if rendStr, ok := fields["rendimiento"]; ok {
		if rend, err := strconv.ParseFloat(rendStr, 64); err == nil {
			partida.Rendimiento = rend
		}
	}
	
	// Parsear recursos por tipo
	partida.ManoObra = s.parseRecursos(content, "mano_obra")
	partida.Materiales = s.parseRecursos(content, "materiales")
	partida.Equipos = s.parseRecursos(content, "equipos")
	partida.Subcontratos = s.parseRecursos(content, "subcontratos")
	
	return partida, nil
}

// parseRecursos extrae recursos de un tipo específico
func (s *ACUParserService) parseRecursos(content, tipoRecurso string) []models.ACURecurso {
	var recursos []models.ACURecurso
	
	// Regex para extraer el bloque de recursos
	pattern := fmt.Sprintf(`%s\s*=\s*\{([^}]+)\}`, tipoRecurso)
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(content)
	
	if len(matches) < 2 {
		return recursos
	}
	
	recursosContent := matches[1]
	
	// Regex para extraer cada recurso individual
	recursoRegex := regexp.MustCompile(`\{([^}]+)\}`)
	recursoMatches := recursoRegex.FindAllStringSubmatch(recursosContent, -1)
	
	for _, match := range recursoMatches {
		if len(match) < 2 {
			continue
		}
		
		recurso := s.parseRecurso(match[1])
		if recurso.Codigo != "" {
			recursos = append(recursos, recurso)
		}
	}
	
	return recursos
}

// parseRecurso parsea un recurso individual
func (s *ACUParserService) parseRecurso(content string) models.ACURecurso {
	recurso := models.ACURecurso{}
	
	fields := s.parseFields(content)
	
	if codigo, ok := fields["codigo"]; ok {
		recurso.Codigo = s.cleanQuotes(codigo)
	}
	if desc, ok := fields["desc"]; ok {
		recurso.Descripcion = s.cleanQuotes(desc)
	}
	if unidad, ok := fields["unidad"]; ok {
		recurso.Unidad = s.cleanQuotes(unidad)
	}
	if cantStr, ok := fields["cantidad"]; ok {
		if cant, err := strconv.ParseFloat(cantStr, 64); err == nil {
			recurso.Cantidad = cant
		}
	}
	if precioStr, ok := fields["precio"]; ok {
		if precio, err := strconv.ParseFloat(precioStr, 64); err == nil {
			recurso.Precio = precio
		}
	}
	if cuadrillaStr, ok := fields["cuadrilla"]; ok {
		if cuadrilla, err := strconv.ParseFloat(cuadrillaStr, 64); err == nil && cuadrilla > 0 {
			recurso.Cuadrilla = &cuadrilla
		}
	}
	
	return recurso
}

// parseFields extrae campos clave=valor del contenido
func (s *ACUParserService) parseFields(content string) map[string]string {
	fields := make(map[string]string)
	
	// Regex mejorada para extraer pares clave=valor, excluyendo bloques
	fieldRegex := regexp.MustCompile(`(\w+)\s*=\s*("([^"]*)"|([^,}{\s]+))`)
	matches := fieldRegex.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			
			// Extraer valor, preferir el grupo con comillas
			var value string
			if match[3] != "" {
				value = match[3] // Valor con comillas
			} else {
				value = match[4] // Valor sin comillas
			}
			
			value = strings.TrimSpace(value)
			if value != "" {
				fields[key] = value
				fmt.Printf("  🔧 Campo: %s = %s\n", key, value)
			}
		}
	}
	
	return fields
}

// cleanQuotes remueve comillas del valor
func (s *ACUParserService) cleanQuotes(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// ConvertToJSON convierte un proyecto ACU a formato JSON legacy
func (s *ACUParserService) ConvertToJSON(project *models.ACUProject) ([]legacy.PartidaLegacy, error) {
	var partidas []legacy.PartidaLegacy
	
	for _, partidaACU := range project.Partidas {
		partida := legacy.PartidaLegacy{
			Codigo:       partidaACU.Codigo,
			Descripcion:  partidaACU.Descripcion,
			Unidad:       partidaACU.Unidad,
			Rendimiento:  partidaACU.Rendimiento,
			ManoObra:     s.convertRecursos(partidaACU.ManoObra),
			Materiales:   s.convertRecursos(partidaACU.Materiales),
			Equipos:      s.convertRecursos(partidaACU.Equipos),
			Subcontratos: s.convertRecursos(partidaACU.Subcontratos),
		}
		
		partidas = append(partidas, partida)
	}
	
	return partidas, nil
}

// convertRecursos convierte recursos ACU a legacy
func (s *ACUParserService) convertRecursos(recursosACU []models.ACURecurso) []legacy.RecursoLegacy {
	var recursos []legacy.RecursoLegacy
	
	for _, recursoACU := range recursosACU {
		recurso := legacy.RecursoLegacy{
			Codigo:      recursoACU.Codigo,
			Descripcion: recursoACU.Descripcion,
			Unidad:      recursoACU.Unidad,
			Cantidad:    recursoACU.Cantidad,
			Precio:      recursoACU.Precio,
		}
		
		if recursoACU.Cuadrilla != nil {
			recurso.Cuadrilla = *recursoACU.Cuadrilla
		}
		
		recursos = append(recursos, recurso)
	}
	
	return recursos
}

// SaveAsJSON guarda el proyecto en formato JSON
func (s *ACUParserService) SaveAsJSON(project *models.ACUProject, filename string) error {
	partidas, err := s.ConvertToJSON(project)
	if err != nil {
		return fmt.Errorf("error convirtiendo a JSON: %w", err)
	}
	
	jsonData, err := json.MarshalIndent(partidas, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando JSON: %w", err)
	}
	
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo: %w", err)
	}
	
	fmt.Printf("💾 Proyecto guardado como JSON: %s\n", filename)
	return nil
}

// ValidateACU valida la sintaxis de un archivo .acu
func (s *ACUParserService) ValidateACU(filename string) error {
	_, err := s.ParseFile(filename)
	if err != nil {
		return fmt.Errorf("error de validación: %w", err)
	}
	
	fmt.Printf("✅ Archivo .acu válido: %s\n", filename)
	return nil
}