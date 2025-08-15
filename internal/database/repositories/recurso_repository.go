package repositories

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"goexcel/internal/database"
	"goexcel/internal/models"
)

type RecursoRepository struct {
	db *database.DB
}

func NewRecursoRepository(db *database.DB) *RecursoRepository {
	return &RecursoRepository{db: db}
}

func (r *RecursoRepository) GetTipoRecursoByNombre(nombre string) (*models.TipoRecurso, error) {
	query := `SELECT id, nombre, descripcion, orden FROM tipos_recurso WHERE nombre = $1`
	
	var tr models.TipoRecurso
	err := r.db.QueryRow(query, nombre).Scan(&tr.ID, &tr.Nombre, &tr.Descripcion, &tr.Orden)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tipo de recurso '%s' no encontrado", nombre)
		}
		return nil, fmt.Errorf("error obteniendo tipo de recurso: %w", err)
	}
	
	return &tr, nil
}

func (r *RecursoRepository) CreateOrGetRecurso(codigo, descripcion, unidad string, precioBase float64, tipoRecursoID uuid.UUID) (*models.Recurso, error) {
	// Usar UPSERT para manejar duplicados
	query := `
		INSERT INTO recursos (codigo, descripcion, unidad, precio_base, tipo_recurso_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			precio_base = EXCLUDED.precio_base,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, codigo, descripcion, unidad, precio_base, tipo_recurso_id, activo, created_at, updated_at
	`
	
	var recurso models.Recurso
	err := r.db.QueryRow(query, codigo, descripcion, unidad, precioBase, tipoRecursoID).Scan(
		&recurso.ID, &recurso.Codigo, &recurso.Descripcion, &recurso.Unidad,
		&recurso.PrecioBase, &recurso.TipoRecursoID, &recurso.Activo,
		&recurso.CreatedAt, &recurso.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("error creando/actualizando recurso: %w", err)
	}
	
	return &recurso, nil
}