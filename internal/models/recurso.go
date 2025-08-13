package models

import (
	"time"

	"github.com/google/uuid"
)

type TipoRecurso struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Nombre      string    `json:"nombre" db:"nombre"`
	Descripcion *string   `json:"descripcion" db:"descripcion"`
	Orden       int       `json:"orden" db:"orden"`
}

type Recurso struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	Codigo        string      `json:"codigo" db:"codigo"`
	Descripcion   string      `json:"descripcion" db:"descripcion"`
	Unidad        string      `json:"unidad" db:"unidad"`
	PrecioBase    float64     `json:"precio_base" db:"precio_base"`
	TipoRecursoID uuid.UUID   `json:"tipo_recurso_id" db:"tipo_recurso_id"`
	Activo        bool        `json:"activo" db:"activo"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`

	// Relaciones
	TipoRecurso *TipoRecurso `json:"tipo_recurso,omitempty"`
}

type PartidaRecurso struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PartidaID  uuid.UUID `json:"partida_id" db:"partida_id"`
	RecursoID  uuid.UUID `json:"recurso_id" db:"recurso_id"`
	Cantidad   float64   `json:"cantidad" db:"cantidad"`
	Precio     float64   `json:"precio" db:"precio"`
	Cuadrilla  *float64  `json:"cuadrilla" db:"cuadrilla"`
	Parcial    float64   `json:"parcial" db:"parcial"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Relaciones
	Partida *Partida `json:"partida,omitempty"`
	Recurso *Recurso `json:"recurso,omitempty"`
}

type PartidaRecursoDetalle struct {
	PartidaRecurso
	RecursoCodigo      string `json:"recurso_codigo" db:"recurso_codigo"`
	RecursoDescripcion string `json:"recurso_descripcion" db:"recurso_descripcion"`
	RecursoUnidad      string `json:"recurso_unidad" db:"recurso_unidad"`
	TipoRecursoNombre  string `json:"tipo_recurso_nombre" db:"tipo_recurso_nombre"`
}

type RecursoCreateRequest struct {
	Codigo        string    `json:"codigo" validate:"required,min=1,max=50"`
	Descripcion   string    `json:"descripcion" validate:"required"`
	Unidad        string    `json:"unidad" validate:"required,max=20"`
	PrecioBase    float64   `json:"precio_base" validate:"min=0"`
	TipoRecursoID uuid.UUID `json:"tipo_recurso_id" validate:"required"`
}

type RecursoUpdateRequest struct {
	Codigo      *string  `json:"codigo,omitempty"`
	Descripcion *string  `json:"descripcion,omitempty"`
	Unidad      *string  `json:"unidad,omitempty"`
	PrecioBase  *float64 `json:"precio_base,omitempty"`
	Activo      *bool    `json:"activo,omitempty"`
}

type PartidaRecursoCreateRequest struct {
	PartidaID uuid.UUID `json:"partida_id" validate:"required"`
	RecursoID uuid.UUID `json:"recurso_id" validate:"required"`
	Cantidad  float64   `json:"cantidad" validate:"min=0"`
	Precio    float64   `json:"precio" validate:"min=0"`
	Cuadrilla *float64  `json:"cuadrilla,omitempty"`
}

type PartidaRecursoUpdateRequest struct {
	Cantidad  *float64 `json:"cantidad,omitempty"`
	Precio    *float64 `json:"precio,omitempty"`
	Cuadrilla *float64 `json:"cuadrilla,omitempty"`
}