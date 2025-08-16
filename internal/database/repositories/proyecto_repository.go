package repositories

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"goexcel/internal/database"
	"goexcel/internal/models"
)

type ProyectoRepository struct {
	db *database.DB
}

func NewProyectoRepository(db *database.DB) *ProyectoRepository {
	return &ProyectoRepository{db: db}
}

func (r *ProyectoRepository) Create(proyecto *models.ProyectoCreateRequest, usuarioID uuid.UUID) (*models.Proyecto, error) {
	query := `
		INSERT INTO proyectos (nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda, 
		                      usuario_id, template_categoria, imagen_portada)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, nombre, descripcion, ubicacion, cliente, fecha_inicio, fecha_fin, moneda,
		          usuario_id, organizacion_id, visibility, template_categoria, imagen_portada,
		          likes_count, vistas_count, created_at, updated_at
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
		usuarioID,
		proyecto.TemplateCategoria,
		proyecto.ImagenPortada,
	).Scan(
		&p.ID,
		&p.Nombre,
		&p.Descripcion,
		&p.Ubicacion,
		&p.Cliente,
		&p.FechaInicio,
		&p.FechaFin,
		&p.Moneda,
		&p.UsuarioID,
		&p.OrganizacionID,
		&p.Visibility,
		&p.TemplateCategoria,
		&p.ImagenPortada,
		&p.LikesCount,
		&p.VistasCount,
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
		SELECT p.id, p.nombre, p.descripcion, p.ubicacion, p.cliente, p.fecha_inicio, p.fecha_fin, 
		       p.moneda, p.usuario_id, p.organizacion_id, p.visibility, p.template_categoria, 
		       p.imagen_portada, p.likes_count, p.vistas_count, p.created_at, p.updated_at,
		       u.nombre as usuario_nombre, u.email as usuario_email,
		       o.nombre as organizacion_nombre
		FROM proyectos p
		LEFT JOIN usuarios u ON p.usuario_id = u.id
		LEFT JOIN organizaciones o ON p.organizacion_id = o.id
		WHERE p.id = $1
	`

	var p models.Proyecto
	var usuarioNombre, usuarioEmail, orgNombre sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Nombre, &p.Descripcion, &p.Ubicacion, &p.Cliente,
		&p.FechaInicio, &p.FechaFin, &p.Moneda, &p.UsuarioID, &p.OrganizacionID,
		&p.Visibility, &p.TemplateCategoria, &p.ImagenPortada,
		&p.LikesCount, &p.VistasCount, &p.CreatedAt, &p.UpdatedAt,
		&usuarioNombre, &usuarioEmail, &orgNombre,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("proyecto no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo proyecto: %w", err)
	}

	// Agregar información del usuario si existe
	if usuarioNombre.Valid && p.UsuarioID != nil {
		p.Usuario = &models.Usuario{
			ID:     *p.UsuarioID,
			Nombre: usuarioNombre.String,
			Email:  usuarioEmail.String,
		}
	}

	// Agregar información de la organización si existe
	if orgNombre.Valid && p.OrganizacionID != nil {
		p.Organizacion = &models.Organizacion{
			ID:     *p.OrganizacionID,
			Nombre: orgNombre.String,
		}
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

// Métodos específicos para multi-tenancy y proyectos públicos

func (r *ProyectoRepository) GetByUsuario(usuarioID uuid.UUID, limit, offset int) ([]models.Proyecto, error) {
	query := `
		SELECT p.id, p.nombre, p.descripcion, p.ubicacion, p.cliente, p.fecha_inicio, p.fecha_fin,
		       p.moneda, p.usuario_id, p.organizacion_id, p.visibility, p.template_categoria,
		       p.imagen_portada, p.likes_count, p.vistas_count, p.created_at, p.updated_at,
		       u.nombre as usuario_nombre, o.nombre as organizacion_nombre
		FROM proyectos p
		LEFT JOIN usuarios u ON p.usuario_id = u.id
		LEFT JOIN organizaciones o ON p.organizacion_id = o.id
		WHERE p.usuario_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	return r.queryProyectos(query, usuarioID, limit, offset)
}

func (r *ProyectoRepository) GetPublicos(limit, offset int) ([]models.Proyecto, error) {
	query := `
		SELECT p.id, p.nombre, p.descripcion, p.ubicacion, p.cliente, p.fecha_inicio, p.fecha_fin,
		       p.moneda, p.usuario_id, p.organizacion_id, p.visibility, p.template_categoria,
		       p.imagen_portada, p.likes_count, p.vistas_count, p.created_at, p.updated_at,
		       u.nombre as usuario_nombre, o.nombre as organizacion_nombre
		FROM proyectos p
		LEFT JOIN usuarios u ON p.usuario_id = u.id
		LEFT JOIN organizaciones o ON p.organizacion_id = o.id
		WHERE p.visibility IN ('public', 'featured')
		ORDER BY p.visibility DESC, p.likes_count DESC, p.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	return r.queryProyectos(query, limit, offset)
}

func (r *ProyectoRepository) GetFeatured(limit int) ([]models.Proyecto, error) {
	query := `
		SELECT p.id, p.nombre, p.descripcion, p.ubicacion, p.cliente, p.fecha_inicio, p.fecha_fin,
		       p.moneda, p.usuario_id, p.organizacion_id, p.visibility, p.template_categoria,
		       p.imagen_portada, p.likes_count, p.vistas_count, p.created_at, p.updated_at,
		       u.nombre as usuario_nombre, o.nombre as organizacion_nombre
		FROM proyectos p
		LEFT JOIN usuarios u ON p.usuario_id = u.id
		LEFT JOIN organizaciones o ON p.organizacion_id = o.id
		WHERE p.visibility = 'featured'
		ORDER BY p.likes_count DESC, p.created_at DESC
		LIMIT $1
	`
	
	return r.queryProyectos(query, limit)
}

func (r *ProyectoRepository) UpdateVisibility(id uuid.UUID, visibility string, usuarioID *uuid.UUID) error {
	var query string
	var args []interface{}

	if usuarioID != nil {
		// Usuario normal solo puede cambiar sus propios proyectos
		query = `UPDATE proyectos SET visibility = $1 WHERE id = $2 AND usuario_id = $3`
		args = []interface{}{visibility, id, *usuarioID}
	} else {
		// Admin puede cambiar cualquier proyecto
		query = `UPDATE proyectos SET visibility = $1 WHERE id = $2`
		args = []interface{}{visibility, id}
	}

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error actualizando visibilidad: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("proyecto no encontrado o sin permisos")
	}

	return nil
}

func (r *ProyectoRepository) ToggleLike(proyectoID, usuarioID uuid.UUID) (bool, error) {
	// Verificar si ya tiene like
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM proyecto_likes WHERE proyecto_id = $1 AND usuario_id = $2)`
	err := r.db.QueryRow(checkQuery, proyectoID, usuarioID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error verificando like: %w", err)
	}

	if exists {
		// Remover like
		_, err = r.db.Exec(`DELETE FROM proyecto_likes WHERE proyecto_id = $1 AND usuario_id = $2`, proyectoID, usuarioID)
		if err != nil {
			return false, fmt.Errorf("error removiendo like: %w", err)
		}
		return false, nil
	} else {
		// Agregar like
		_, err = r.db.Exec(`INSERT INTO proyecto_likes (proyecto_id, usuario_id) VALUES ($1, $2)`, proyectoID, usuarioID)
		if err != nil {
			return false, fmt.Errorf("error agregando like: %w", err)
		}
		return true, nil
	}
}

func (r *ProyectoRepository) IncrementViews(id uuid.UUID) error {
	query := `UPDATE proyectos SET vistas_count = vistas_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ProyectoRepository) GetByIDWithLikeStatus(id uuid.UUID, usuarioID *uuid.UUID) (*models.Proyecto, error) {
	proyecto, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Si hay usuario, verificar si le dio like
	if usuarioID != nil {
		var isLiked bool
		query := `SELECT EXISTS(SELECT 1 FROM proyecto_likes WHERE proyecto_id = $1 AND usuario_id = $2)`
		err = r.db.QueryRow(query, id, *usuarioID).Scan(&isLiked)
		if err != nil {
			return nil, fmt.Errorf("error verificando like status: %w", err)
		}
		proyecto.IsLiked = isLiked
	}

	return proyecto, nil
}

// Helper method para queries comunes
func (r *ProyectoRepository) queryProyectos(query string, args ...interface{}) ([]models.Proyecto, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando query: %w", err)
	}
	defer rows.Close()

	var proyectos []models.Proyecto
	for rows.Next() {
		var p models.Proyecto
		var usuarioNombre, orgNombre sql.NullString

		err := rows.Scan(
			&p.ID, &p.Nombre, &p.Descripcion, &p.Ubicacion, &p.Cliente,
			&p.FechaInicio, &p.FechaFin, &p.Moneda, &p.UsuarioID, &p.OrganizacionID,
			&p.Visibility, &p.TemplateCategoria, &p.ImagenPortada,
			&p.LikesCount, &p.VistasCount, &p.CreatedAt, &p.UpdatedAt,
			&usuarioNombre, &orgNombre,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando proyecto: %w", err)
		}

		// Agregar información del usuario y organización
		if usuarioNombre.Valid && p.UsuarioID != nil {
			p.Usuario = &models.Usuario{
				ID:     *p.UsuarioID,
				Nombre: usuarioNombre.String,
			}
		}
		if orgNombre.Valid && p.OrganizacionID != nil {
			p.Organizacion = &models.Organizacion{
				ID:     *p.OrganizacionID,
				Nombre: orgNombre.String,
			}
		}

		proyectos = append(proyectos, p)
	}

	return proyectos, rows.Err()
}