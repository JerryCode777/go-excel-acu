package services

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"goexcel/internal/database/repositories"
	"goexcel/internal/models"
)

// MigrationJerarquicoService maneja la migración de datos ACU jerárquicos a PostgreSQL
type MigrationJerarquicoService struct {
	presupuestoRepo *repositories.PresupuestoRepository
}

// NewMigrationJerarquicoService crea una nueva instancia del servicio de migración jerárquica
func NewMigrationJerarquicoService(db *sql.DB) *MigrationJerarquicoService {
	return &MigrationJerarquicoService{
		presupuestoRepo: repositories.NewPresupuestoRepository(db),
	}
}

// MigrarACUJerarquico migra un ACU jerárquico parseado a la base de datos
func (s *MigrationJerarquicoService) MigrarACUJerarquico(acuData *models.ACUJerarquico) (*models.Presupuesto, error) {
	// 1. Crear presupuesto principal
	presupuestoReq := models.PresupuestoRequest{
		Codigo:  acuData.Presupuesto.Codigo,
		Nombre:  acuData.Presupuesto.Nombre,
		Cliente: acuData.Presupuesto.Cliente,
		Lugar:   acuData.Presupuesto.Lugar,
		Moneda:  acuData.Presupuesto.Moneda,
	}

	presupuesto, err := s.presupuestoRepo.CrearPresupuesto(presupuestoReq, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando presupuesto: %v", err)
	}

	fmt.Printf("✅ Presupuesto creado: %s (ID: %s)\n", presupuesto.Nombre, presupuesto.ID.String()[:8])

	// 2. Crear subpresupuestos
	subpresupuestoMap := make(map[string]uuid.UUID)
	for _, subData := range acuData.Subpresupuestos {
		subReq := models.SubpresupuestoRequest{
			Codigo: subData.Codigo,
			Nombre: subData.Nombre,
			Orden:  len(subpresupuestoMap) + 1,
		}

		subpresupuesto, err := s.presupuestoRepo.CrearSubpresupuesto(presupuesto.ID, subReq)
		if err != nil {
			return nil, fmt.Errorf("error creando subpresupuesto %s: %v", subData.Codigo, err)
		}

		subpresupuestoMap[subData.Codigo] = subpresupuesto.ID
		fmt.Printf("  📁 Subpresupuesto: %s\n", subData.Nombre)
	}

	// 3. Crear títulos jerárquicos
	tituloMap := make(map[string]uuid.UUID)
	for _, tituloData := range acuData.Titulos {
		// Determinar título padre basado en el código
		var tituloPadreID *uuid.UUID
		if tituloData.Nivel > 1 {
			codigoPadre := s.obtenerCodigoPadre(tituloData.CodigoCompleto)
			if parentID, exists := tituloMap[codigoPadre]; exists {
				tituloPadreID = &parentID
			}
		}

		tituloReq := models.TituloRequest{
			Nombre:        tituloData.Nombre,
			TituloPadreID: tituloPadreID,
			Orden:         tituloData.Numero,
		}

		titulo, err := s.presupuestoRepo.CrearTitulo(presupuesto.ID, tituloReq)
		if err != nil {
			return nil, fmt.Errorf("error creando título %s: %v", tituloData.CodigoCompleto, err)
		}

		tituloMap[tituloData.CodigoCompleto] = titulo.ID
		fmt.Printf("  📋 Título %s: %s\n", titulo.CodigoCompleto, titulo.Nombre)
	}

	// 4. Crear partidas jerárquicas
	for _, partidaData := range acuData.Partidas {
		// Determinar título padre basado en el código de la partida
		tituloID := s.obtenerTituloIDParaPartida(partidaData.Codigo, tituloMap)

		partidaReq := models.PartidaJerarquicaRequest{
			Descripcion:  partidaData.Descripcion,
			Unidad:       partidaData.Unidad,
			Rendimiento:  partidaData.Rendimiento,
			TituloID:     tituloID,
			Orden:        s.obtenerNumeroPartida(partidaData.Codigo),
		}

		partida, err := s.presupuestoRepo.CrearPartidaJerarquica(presupuesto.ID, partidaReq)
		if err != nil {
			return nil, fmt.Errorf("error creando partida %s: %v", partidaData.Codigo, err)
		}

		fmt.Printf("    ⚙️  Partida %s: %s\n", partida.PartidaCodigo, partida.PartidaDescripcion)

		// TODO: Crear recursos de la partida (mano_obra, materiales, equipos, subcontratos)
		// Por ahora solo creamos la estructura jerárquica
	}

	fmt.Printf("📊 Migración completada: %d títulos, %d partidas\n", 
		len(acuData.Titulos), len(acuData.Partidas))

	return presupuesto, nil
}

// obtenerCodigoPadre extrae el código padre de un código jerárquico
func (s *MigrationJerarquicoService) obtenerCodigoPadre(codigo string) string {
	// Ejemplo: "01.02.03" -> "01.02"
	lastDot := -1
	for i := len(codigo) - 1; i >= 0; i-- {
		if codigo[i] == '.' {
			lastDot = i
			break
		}
	}
	
	if lastDot > 0 {
		return codigo[:lastDot]
	}
	return ""
}

// obtenerTituloIDParaPartida encuentra el ID del título al que pertenece una partida
func (s *MigrationJerarquicoService) obtenerTituloIDParaPartida(codigoPartida string, tituloMap map[string]uuid.UUID) *uuid.UUID {
	// Ejemplo: partida "01.02.03.01" pertenece al título "01.02.03"
	tituloCode := s.obtenerCodigoPadre(codigoPartida)
	
	if tituloID, exists := tituloMap[tituloCode]; exists {
		return &tituloID
	}
	return nil
}

// obtenerNumeroPartida extrae el número de la partida de su código
func (s *MigrationJerarquicoService) obtenerNumeroPartida(codigoPartida string) int {
	// Ejemplo: "01.02.03.05" -> 5
	lastDot := -1
	for i := len(codigoPartida) - 1; i >= 0; i-- {
		if codigoPartida[i] == '.' {
			lastDot = i
			break
		}
	}
	
	if lastDot >= 0 && lastDot < len(codigoPartida)-1 {
		numeroStr := codigoPartida[lastDot+1:]
		if num := parseInt(numeroStr); num > 0 {
			return num
		}
	}
	return 1
}

// parseInt convierte string a int de manera segura
func parseInt(s string) int {
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}