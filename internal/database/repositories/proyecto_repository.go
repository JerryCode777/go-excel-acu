package repositories

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jerryandersonh/goexcel/internal/database"
	"github.com/jerryandersonh/goexcel/internal/models"
)

type ProyectoRepository struct {
	db *database.DB
}

func NewProyectoRepository(db *database.DB) *ProyectoRepository {
	return &ProyectoRepository{db: db}
}

func (r *ProyectoRepository) Create(proyecto *models.ProyectoCreateRequest) (*models.Proyecto, error) {
	query := `
		INSERT INTO proyectos (nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda, created_at, updated_at
	`

	var p models.Proyecto
	err := r.db.QueryRow(
		query,
		proyecto.Nombre,
		proyecto.Descripcion,
		proyecto.Ubicacion,
		proyecto.Cliente,
		proyecto.FechaInicio,
		proyecto.FechaFin,
		proyecto.Moneda,
	).Scan(
		&p.ID,
		&p.Nombre,
		&p.Descripcion,
		&p.Ubicacion,
		&p.Cliente,
		&p.FechaInicio,
		&p.FechaFin,
		&p.Moneda,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando proyecto: %w", err)
	}

	return &p, nil
}

func (r *ProyectoRepository) GetByID(id uuid.UUID) (*models.Proyecto, error) {
	query := `
		SELECT id, nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda, created_at, updated_at
		FROM proyectos
		WHERE id = $1
	`

	var p models.Proyecto
	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Nombre,
		&p.Descripcion,
		&p.Ubicacion,
		&p.Cliente,
		&p.FechaInicio,
		&p.FechaFin,
		&p.Moneda,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("proyecto no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo proyecto: %w", err)
	}

	return &p, nil
}

func (r *ProyectoRepository) GetAll() ([]models.Proyecto, error) {
	query := `
		SELECT id, nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda, created_at, updated_at
		FROM proyectos
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo proyectos: %w", err)
	}
	defer rows.Close()

	var proyectos []models.Proyecto
	for rows.Next() {
		var p models.Proyecto
		err := rows.Scan(
			&p.ID,
			&p.Nombre,
			&p.Descripcion,
			&p.Ubicacion,
			&p.Cliente,
			&p.FechaInicio,
			&p.FechaFin,
			&p.Moneda,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando proyecto: %w", err)
		}
		proyectos = append(proyectos, p)
	}

	return proyectos, nil
}

func (r *ProyectoRepository) Update(id uuid.UUID, proyecto *models.ProyectoUpdateRequest) (*models.Proyecto, error) {
	// Construir query dinámicamente basado en los campos no nulos
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if proyecto.Nombre != nil {
		setParts = append(setParts, fmt.Sprintf("nombre = $%d", argCount))
		args = append(args, *proyecto.Nombre)
		argCount++
	}
	if proyecto.Descripcion != nil {
		setParts = append(setParts, fmt.Sprintf("descripcion = $%d", argCount))
		args = append(args, *proyecto.Descripcion)
		argCount++
	}
	if proyecto.Ubicacion != nil {
		setParts = append(setParts, fmt.Sprintf("ubicacion = $%d", argCount))
		args = append(args, *proyecto.Ubicacion)
		argCount++
	}
	if proyecto.Cliente != nil {
		setParts = append(setParts, fmt.Sprintf("cliente = $%d", argCount))
		args = append(args, *proyecto.Cliente)
		argCount++
	}
	if proyecto.FechaInicio != nil {
		setParts = append(setParts, fmt.Sprintf("fecha_inicio = $%d", argCount))
		args = append(args, *proyecto.FechaInicio)
		argCount++
	}
	if proyecto.FechaFin != nil {
		setParts = append(setParts, fmt.Sprintf("fecha_fin = $%d", argCount))
		args = append(args, *proyecto.FechaFin)
		argCount++
	}
	if proyecto.Moneda != nil {
		setParts = append(setParts, fmt.Sprintf("moneda = $%d", argCount))
		args = append(args, *proyecto.Moneda)
		argCount++
	}

	if len(setParts) == 0 {
		return r.GetByID(id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = CURRENT_TIMESTAMP"))
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE proyectos 
		SET %s
		WHERE id = $%d
		RETURNING id, nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda, created_at, updated_at
	`, fmt.Sprintf("%s", setParts), argCount)

	var p models.Proyecto
	err := r.db.QueryRow(query, args...).Scan(
		&p.ID,
		&p.Nombre,
		&p.Descripcion,
		&p.Ubicacion,
		&p.Cliente,
		&p.FechaInicio,
		&p.FechaFin,
		&p.Moneda,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error actualizando proyecto: %w", err)
	}

	return &p, nil
}

func (r *ProyectoRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM proyectos WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error eliminando proyecto: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("proyecto no encontrado")
	}

	return nil
}