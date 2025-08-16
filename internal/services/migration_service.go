package services

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"goexcel/internal/database/repositories"
	"goexcel/internal/legacy"
	"goexcel/internal/models"
)

type MigrationService struct {
	proyectoRepo *repositories.ProyectoRepository
	partidaRepo  *repositories.PartidaRepository
	recursoRepo  *repositories.RecursoRepository
}

func NewMigrationService(
	proyectoRepo *repositories.ProyectoRepository,
	partidaRepo *repositories.PartidaRepository,
	recursoRepo *repositories.RecursoRepository,
) *MigrationService {
	return &MigrationService{
		proyectoRepo: proyectoRepo,
		partidaRepo:  partidaRepo,
		recursoRepo:  recursoRepo,
	}
}

func (s *MigrationService) MigrarDesdeJSON(partidasJSON []legacy.PartidaLegacy, nombreArchivo string) (*models.Proyecto, error) {
	// 1. Crear proyecto
	proyectoReq := &models.ProyectoCreateRequest{
		Nombre: fmt.Sprintf("Proyecto Migrado - %s", nombreArchivo),
		Moneda: "PEN",
	}
	
	// Para migración, usar usuario admin por defecto
	adminUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	proyecto, err := s.proyectoRepo.Create(proyectoReq, adminUserID)
	if err != nil {
		return nil, fmt.Errorf("error creando proyecto: %w", err)
	}

	log.Printf("📁 Proyecto creado: %s (ID: %s)", proyecto.Nombre, proyecto.ID.String()[:8])

	// 2. Obtener tipos de recurso
	tipoMO, err := s.recursoRepo.GetTipoRecursoByNombre("mano_obra")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tipo mano_obra: %w", err)
	}
	
	tipoMat, err := s.recursoRepo.GetTipoRecursoByNombre("materiales")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tipo materiales: %w", err)
	}
	
	tipoEq, err := s.recursoRepo.GetTipoRecursoByNombre("equipos")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tipo equipos: %w", err)
	}
	
	tipoSub, err := s.recursoRepo.GetTipoRecursoByNombre("subcontratos")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tipo subcontratos: %w", err)
	}

	// 3. Procesar cada partida
	for i, partidaJSON := range partidasJSON {
		log.Printf("Procesando partida %d/%d: %s", i+1, len(partidasJSON), partidaJSON.Codigo)
		
		// Crear partida
		partidaReq := &models.PartidaCreateRequest{
			ProyectoID:  proyecto.ID,
			Codigo:      partidaJSON.Codigo,
			Descripcion: partidaJSON.Descripcion,
			Unidad:      partidaJSON.Unidad,
			Rendimiento: partidaJSON.Rendimiento,
		}
		
		partida, err := s.partidaRepo.Create(partidaReq)
		if err != nil {
			log.Printf("Error creando partida %s: %v", partidaJSON.Codigo, err)
			continue
		}

		// Procesar mano de obra
		if err := s.procesarRecursos(partida.ID, partidaJSON.ManoObra, tipoMO.ID); err != nil {
			log.Printf("Error procesando mano de obra para %s: %v", partidaJSON.Codigo, err)
		}

		// Procesar materiales
		if err := s.procesarRecursos(partida.ID, partidaJSON.Materiales, tipoMat.ID); err != nil {
			log.Printf("Error procesando materiales para %s: %v", partidaJSON.Codigo, err)
		}

		// Procesar equipos
		if err := s.procesarRecursos(partida.ID, partidaJSON.Equipos, tipoEq.ID); err != nil {
			log.Printf("Error procesando equipos para %s: %v", partidaJSON.Codigo, err)
		}

		// Procesar subcontratos
		if err := s.procesarRecursos(partida.ID, partidaJSON.Subcontratos, tipoSub.ID); err != nil {
			log.Printf("Error procesando subcontratos para %s: %v", partidaJSON.Codigo, err)
		}
	}

	log.Printf("✅ %d partidas procesadas para el proyecto %s", len(partidasJSON), proyecto.ID.String()[:8])
	
	return proyecto, nil
}

func (s *MigrationService) procesarRecursos(partidaID uuid.UUID, recursos []legacy.RecursoLegacy, tipoRecursoID uuid.UUID) error {
	for _, recursoJSON := range recursos {
		if recursoJSON.Codigo == "" || recursoJSON.Descripcion == "" {
			continue
		}

		// Crear o obtener recurso
		recurso, err := s.recursoRepo.CreateOrGetRecurso(
			recursoJSON.Codigo,
			recursoJSON.Descripcion,
			recursoJSON.Unidad,
			recursoJSON.Precio, // Usar precio del JSON como precio base
			tipoRecursoID,
		)
		if err != nil {
			return fmt.Errorf("error creando recurso %s: %w", recursoJSON.Codigo, err)
		}

		// Crear relación partida-recurso
		var cuadrilla *float64
		if recursoJSON.Cuadrilla > 0 {
			cuadrilla = &recursoJSON.Cuadrilla
		}

		partidaRecursoReq := &models.PartidaRecursoCreateRequest{
			PartidaID: partidaID,
			RecursoID: recurso.ID,
			Cantidad:  recursoJSON.Cantidad,
			Precio:    recursoJSON.Precio,
			Cuadrilla: cuadrilla,
		}

		_, err = s.partidaRepo.AddRecurso(partidaRecursoReq)
		if err != nil {
			return fmt.Errorf("error agregando recurso %s a partida: %w", recursoJSON.Codigo, err)
		}
	}

	return nil
}