package models

import (
	"time"

	"github.com/google/uuid"
)

type Usuario struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Nombre           string     `json:"nombre" db:"nombre"`
	Apellido         *string    `json:"apellido" db:"apellido"`
	Rol              string     `json:"rol" db:"rol"`
	OrganizacionID   *uuid.UUID `json:"organizacion_id" db:"organizacion_id"`
	AvatarURL        *string    `json:"avatar_url" db:"avatar_url"`
	Activo           bool       `json:"activo" db:"activo"`
	EmailVerificado  bool       `json:"email_verificado" db:"email_verificado"`
	UltimoAcceso     *time.Time `json:"ultimo_acceso" db:"ultimo_acceso"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// Relaciones
	Organizacion *Organizacion `json:"organizacion,omitempty"`
	Proyectos    []Proyecto    `json:"proyectos,omitempty"`
}

type UsuarioCreateRequest struct {
	Email          string     `json:"email" validate:"required,email"`
	Password       string     `json:"password" validate:"required,min=6"`
	Nombre         string     `json:"nombre" validate:"required,min=1,max=255"`
	Apellido       *string    `json:"apellido"`
	OrganizacionID *uuid.UUID `json:"organizacion_id"`
}

type UsuarioUpdateRequest struct {
	Nombre         *string    `json:"nombre,omitempty"`
	Apellido       *string    `json:"apellido,omitempty"`
	AvatarURL      *string    `json:"avatar_url,omitempty"`
	OrganizacionID *uuid.UUID `json:"organizacion_id,omitempty"`
}

type UsuarioLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UsuarioLoginResponse struct {
	Token   string  `json:"token"`
	Usuario Usuario `json:"usuario"`
}

type UsuarioChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// SesionUsuario representa una sesión activa de usuario
type SesionUsuario struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UsuarioID uuid.UUID `json:"usuario_id" db:"usuario_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	IPAddress *string   `json:"ip_address" db:"ip_address"`
	UserAgent *string   `json:"user_agent" db:"user_agent"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}