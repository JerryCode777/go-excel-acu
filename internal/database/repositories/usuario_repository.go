package repositories

import (
	"database/sql"
	"fmt"

	"goexcel/internal/database"
	"goexcel/internal/models"

	"github.com/google/uuid"
)

type UsuarioRepository struct {
	db *database.DB
}

func NewUsuarioRepository(db *database.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

func (r *UsuarioRepository) Create(usuario *models.Usuario) error {
	query := `
		INSERT INTO usuarios (id, email, password_hash, nombre, apellido, rol, organizacion_id, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`
	
	if usuario.ID == uuid.Nil {
		usuario.ID = uuid.New()
	}

	err := r.db.QueryRow(query,
		usuario.ID,
		usuario.Email,
		usuario.PasswordHash,
		usuario.Nombre,
		usuario.Apellido,
		usuario.Rol,
		usuario.OrganizacionID,
		usuario.AvatarURL,
	).Scan(&usuario.CreatedAt, &usuario.UpdatedAt)

	return err
}

func (r *UsuarioRepository) GetByID(id uuid.UUID) (*models.Usuario, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.nombre, u.apellido, u.rol, 
		       u.organizacion_id, u.avatar_url, u.activo, u.email_verificado, 
		       u.ultimo_acceso, u.created_at, u.updated_at,
		       o.id, o.nombre, o.descripcion, o.logo_url, o.activo, o.created_at, o.updated_at
		FROM usuarios u
		LEFT JOIN organizaciones o ON u.organizacion_id = o.id
		WHERE u.id = $1
	`
	
	var usuario models.Usuario
	var org models.Organizacion
	var orgID sql.NullString
	var orgNombre sql.NullString
	var orgDescripcion sql.NullString
	var orgLogoURL sql.NullString
	var orgActivo sql.NullBool
	var orgCreatedAt sql.NullTime
	var orgUpdatedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&usuario.ID, &usuario.Email, &usuario.PasswordHash, &usuario.Nombre,
		&usuario.Apellido, &usuario.Rol, &usuario.OrganizacionID, &usuario.AvatarURL,
		&usuario.Activo, &usuario.EmailVerificado, &usuario.UltimoAcceso,
		&usuario.CreatedAt, &usuario.UpdatedAt,
		&orgID, &orgNombre, &orgDescripcion, &orgLogoURL, &orgActivo, &orgCreatedAt, &orgUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, err
	}

	// Asignar organización si existe
	if orgID.Valid {
		orgUUID, _ := uuid.Parse(orgID.String)
		org.ID = orgUUID
		org.Nombre = orgNombre.String
		if orgDescripcion.Valid {
			org.Descripcion = &orgDescripcion.String
		}
		if orgLogoURL.Valid {
			org.LogoURL = &orgLogoURL.String
		}
		org.Activo = orgActivo.Bool
		org.CreatedAt = orgCreatedAt.Time
		org.UpdatedAt = orgUpdatedAt.Time
		usuario.Organizacion = &org
	}

	return &usuario, nil
}

func (r *UsuarioRepository) GetByEmail(email string) (*models.Usuario, error) {
	query := `
		SELECT id, email, password_hash, nombre, apellido, rol, organizacion_id, 
		       avatar_url, activo, email_verificado, ultimo_acceso, created_at, updated_at
		FROM usuarios 
		WHERE email = $1 AND activo = true
	`
	
	var usuario models.Usuario
	err := r.db.QueryRow(query, email).Scan(
		&usuario.ID, &usuario.Email, &usuario.PasswordHash, &usuario.Nombre,
		&usuario.Apellido, &usuario.Rol, &usuario.OrganizacionID, &usuario.AvatarURL,
		&usuario.Activo, &usuario.EmailVerificado, &usuario.UltimoAcceso,
		&usuario.CreatedAt, &usuario.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, err
	}

	return &usuario, nil
}

func (r *UsuarioRepository) Update(usuario *models.Usuario) error {
	query := `
		UPDATE usuarios 
		SET nombre = $2, apellido = $3, avatar_url = $4, organizacion_id = $5,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	
	err := r.db.QueryRow(query,
		usuario.ID,
		usuario.Nombre,
		usuario.Apellido,
		usuario.AvatarURL,
		usuario.OrganizacionID,
	).Scan(&usuario.UpdatedAt)

	return err
}

func (r *UsuarioRepository) UpdatePassword(userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE usuarios 
		SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	result, err := r.db.Exec(query, userID, passwordHash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("usuario no encontrado")
	}

	return nil
}

func (r *UsuarioRepository) UpdateLastAccess(userID uuid.UUID) error {
	query := `
		UPDATE usuarios 
		SET ultimo_acceso = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *UsuarioRepository) GetAll(limit, offset int) ([]models.Usuario, error) {
	query := `
		SELECT u.id, u.email, u.nombre, u.apellido, u.rol, u.organizacion_id,
		       u.avatar_url, u.activo, u.email_verificado, u.ultimo_acceso,
		       u.created_at, u.updated_at,
		       o.nombre as org_nombre
		FROM usuarios u
		LEFT JOIN organizaciones o ON u.organizacion_id = o.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []models.Usuario
	for rows.Next() {
		var usuario models.Usuario
		var orgNombre sql.NullString

		err := rows.Scan(
			&usuario.ID, &usuario.Email, &usuario.Nombre, &usuario.Apellido,
			&usuario.Rol, &usuario.OrganizacionID, &usuario.AvatarURL,
			&usuario.Activo, &usuario.EmailVerificado, &usuario.UltimoAcceso,
			&usuario.CreatedAt, &usuario.UpdatedAt, &orgNombre,
		)
		if err != nil {
			return nil, err
		}

		if orgNombre.Valid {
			usuario.Organizacion = &models.Organizacion{Nombre: orgNombre.String}
		}

		usuarios = append(usuarios, usuario)
	}

	return usuarios, rows.Err()
}

func (r *UsuarioRepository) Delete(id uuid.UUID) error {
	query := `UPDATE usuarios SET activo = false WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("usuario no encontrado")
	}

	return nil
}

func (r *UsuarioRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM usuarios WHERE email = $1 AND activo = true)`
	
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}