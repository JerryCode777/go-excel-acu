package repositories

import (
	"database/sql"
	"fmt"

	"goexcel/internal/database"
	"goexcel/internal/models"

	"github.com/google/uuid"
)

type OrganizacionRepository struct {
	db *database.DB
}

func NewOrganizacionRepository(db *database.DB) *OrganizacionRepository {
	return &OrganizacionRepository{db: db}
}

func (r *OrganizacionRepository) Create(organizacion *models.Organizacion) error {
	query := `
		INSERT INTO organizaciones (id, nombre, descripcion, logo_url)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`
	
	if organizacion.ID == uuid.Nil {
		organizacion.ID = uuid.New()
	}

	err := r.db.QueryRow(query,
		organizacion.ID,
		organizacion.Nombre,
		organizacion.Descripcion,
		organizacion.LogoURL,
	).Scan(&organizacion.CreatedAt, &organizacion.UpdatedAt)

	return err
}

func (r *OrganizacionRepository) GetByID(id uuid.UUID) (*models.Organizacion, error) {
	query := `
		SELECT id, nombre, descripcion, logo_url, activo, created_at, updated_at
		FROM organizaciones 
		WHERE id = $1
	`
	
	var organizacion models.Organizacion
	err := r.db.QueryRow(query, id).Scan(
		&organizacion.ID, &organizacion.Nombre, &organizacion.Descripcion,
		&organizacion.LogoURL, &organizacion.Activo, &organizacion.CreatedAt, &organizacion.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organización no encontrada")
		}
		return nil, err
	}

	return &organizacion, nil
}

func (r *OrganizacionRepository) GetAll(limit, offset int) ([]models.Organizacion, error) {
	query := `
		SELECT o.id, o.nombre, o.descripcion, o.logo_url, o.activo, 
		       o.created_at, o.updated_at,
		       COUNT(u.id) as total_usuarios,
		       COUNT(p.id) as total_proyectos
		FROM organizaciones o
		LEFT JOIN usuarios u ON o.id = u.organizacion_id AND u.activo = true
		LEFT JOIN proyectos p ON o.id = p.organizacion_id
		WHERE o.activo = true
		GROUP BY o.id, o.nombre, o.descripcion, o.logo_url, o.activo, o.created_at, o.updated_at
		ORDER BY o.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizaciones []models.Organizacion
	for rows.Next() {
		var organizacion models.Organizacion
		var totalUsuarios, totalProyectos int

		err := rows.Scan(
			&organizacion.ID, &organizacion.Nombre, &organizacion.Descripcion,
			&organizacion.LogoURL, &organizacion.Activo, &organizacion.CreatedAt, &organizacion.UpdatedAt,
			&totalUsuarios, &totalProyectos,
		)
		if err != nil {
			return nil, err
		}

		organizaciones = append(organizaciones, organizacion)
	}

	return organizaciones, rows.Err()
}

func (r *OrganizacionRepository) Update(organizacion *models.Organizacion) error {
	query := `
		UPDATE organizaciones 
		SET nombre = $2, descripcion = $3, logo_url = $4, activo = $5,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	
	err := r.db.QueryRow(query,
		organizacion.ID,
		organizacion.Nombre,
		organizacion.Descripcion,
		organizacion.LogoURL,
		organizacion.Activo,
	).Scan(&organizacion.UpdatedAt)

	return err
}

func (r *OrganizacionRepository) Delete(id uuid.UUID) error {
	query := `UPDATE organizaciones SET activo = false WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organización no encontrada")
	}

	return nil
}

func (r *OrganizacionRepository) GetUsuarios(organizacionID uuid.UUID, limit, offset int) ([]models.Usuario, error) {
	query := `
		SELECT id, email, nombre, apellido, rol, avatar_url, activo, 
		       email_verificado, ultimo_acceso, created_at, updated_at
		FROM usuarios
		WHERE organizacion_id = $1 AND activo = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, organizacionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []models.Usuario
	for rows.Next() {
		var usuario models.Usuario
		err := rows.Scan(
			&usuario.ID, &usuario.Email, &usuario.Nombre, &usuario.Apellido,
			&usuario.Rol, &usuario.AvatarURL, &usuario.Activo,
			&usuario.EmailVerificado, &usuario.UltimoAcceso,
			&usuario.CreatedAt, &usuario.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, usuario)
	}

	return usuarios, rows.Err()
}