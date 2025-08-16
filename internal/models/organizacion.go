package models

import (
	"time"

	"github.com/google/uuid"
)

type Organizacion struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Nombre      string    `json:"nombre" db:"nombre"`
	Descripcion *string   `json:"descripcion" db:"descripcion"`
	LogoURL     *string   `json:"logo_url" db:"logo_url"`
	Activo      bool      `json:"activo" db:"activo"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relaciones
	Usuarios  []Usuario  `json:"usuarios,omitempty"`
	Proyectos []Proyecto `json:"proyectos,omitempty"`
}

type OrganizacionCreateRequest struct {
	Nombre      string  `json:"nombre" validate:"required,min=1,max=255"`
	Descripcion *string `json:"descripcion"`
	LogoURL     *string `json:"logo_url"`
}

type OrganizacionUpdateRequest struct {
	Nombre      *string `json:"nombre,omitempty"`
	Descripcion *string `json:"descripcion,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Activo      *bool   `json:"activo,omitempty"`
}