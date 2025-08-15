package repositories

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jerryandersonh/goexcel/internal/database"
	"github.com/jerryandersonh/goexcel/internal/models"
)

type PartidaRepository struct {
	db *database.DB
}

func NewPartidaRepository(db *database.DB) *PartidaRepository {
	return &PartidaRepository{db: db}
}

func (r *PartidaRepository) Create(partida *models.PartidaCreateRequest) (*models.Partida, error) {
	query := `
		INSERT INTO partidas (proyecto_id, codigo, descripcion, unidad, rendimiento)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, proyecto_id, codigo, descripcion, unidad, rendimiento, costo_total, activo, created_at, updated_at
	`

	var p models.Partida
	err := r.db.QueryRow(
		query,
		partida.ProyectoID,
		partida.Codigo,
		partida.Descripcion,
		partida.Unidad,
		partida.Rendimiento,
	).Scan(
		&p.ID,
		&p.ProyectoID,
		&p.Codigo,
		&p.Descripcion,
		&p.Unidad,
		&p.Rendimiento,
		&p.CostoTotal,
		&p.Activo,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando partida: %w", err)
	}

	return &p, nil
}

func (r *PartidaRepository) AddRecurso(req *models.PartidaRecursoCreateRequest) (*models.PartidaRecurso, error) {
	query := `
		INSERT INTO partida_recursos (partida_id, recurso_id, cantidad, precio, cuadrilla)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
			cantidad = EXCLUDED.cantidad,
			precio = EXCLUDED.precio,
			cuadrilla = EXCLUDED.cuadrilla,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, partida_id, recurso_id, cantidad, precio, cuadrilla, parcial, created_at, updated_at
	`

	var pr models.PartidaRecurso
	err := r.db.QueryRow(
		query,
		req.PartidaID,
		req.RecursoID,
		req.Cantidad,
		req.Precio,
		req.Cuadrilla,
	).Scan(
		&pr.ID,
		&pr.PartidaID,
		&pr.RecursoID,
		&pr.Cantidad,
		&pr.Precio,
		&pr.Cuadrilla,
		&pr.Parcial,
		&pr.CreatedAt,
		&pr.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error agregando recurso a partida: %w", err)
	}

	return &pr, nil
}

func (r *PartidaRepository) GetByProyectoID(proyectoID uuid.UUID) ([]models.PartidaCompleta, error) {
	query := `
		SELECT 
			id, codigo, descripcion, unidad, rendimiento, costo_total,
			proyecto_nombre, costo_mano_obra, costo_materiales, costo_equipos, costo_subcontratos
		FROM vista_partidas_completas v
		WHERE EXISTS (SELECT 1 FROM partidas p WHERE p.id = v.id AND p.proyecto_id = $1)
		ORDER BY codigo
	`

	rows, err := r.db.Query(query, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo partidas: %w", err)
	}
	defer rows.Close()

	var partidas []models.PartidaCompleta
	for rows.Next() {
		var p models.PartidaCompleta
		err := rows.Scan(
			&p.ID, &p.Codigo, &p.Descripcion, &p.Unidad, &p.Rendimiento, &p.CostoTotal,
			&p.ProyectoNombre, &p.CostoManoObra, &p.CostoMateriales, &p.CostoEquipos, &p.CostoSubcontratos,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando partida: %w", err)
		}
		partidas = append(partidas, p)
	}

	return partidas, nil
}

// GetDB retorna la conexión a la base de datos
func (r *PartidaRepository) GetDB() *database.DB {
	return r.db
}