package models

import (
	"time"

	"github.com/google/uuid"
)

type Proyecto struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Nombre      string     `json:"nombre" db:"nombre"`
	Descripcion *string    `json:"descripcion" db:"descripcion"`
	Ubicacion   *string    `json:"ubicacion" db:"ubicacion"`
	Cliente     *string    `json:"cliente" db:"cliente"`
	FechaInicio *time.Time `json:"fecha_inicio" db:"fecha_inicio"`
	FechaFin    *time.Time `json:"fecha_fin" db:"fecha_fin"`
	Moneda      string     `json:"moneda" db:"moneda"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	
	// Relaciones
	Partidas []Partida `json:"partidas,omitempty"`
}

type ProyectoCreateRequest struct {
	Nombre      string     `json:"nombre" validate:"required,min=1,max=255"`
	Descripcion *string    `json:"descripcion"`
	Ubicacion   *string    `json:"ubicacion"`
	Cliente     *string    `json:"cliente"`
	FechaInicio *time.Time `json:"fecha_inicio"`
	FechaFin    *time.Time `json:"fecha_fin"`
	Moneda      string     `json:"moneda" validate:"required,len=3"`
}

type ProyectoUpdateRequest struct {
	Nombre      *string    `json:"nombre,omitempty"`
	Descripcion *string    `json:"descripcion,omitempty"`
	Ubicacion   *string    `json:"ubicacion,omitempty"`
	Cliente     *string    `json:"cliente,omitempty"`
	FechaInicio *time.Time `json:"fecha_inicio,omitempty"`
	FechaFin    *time.Time `json:"fecha_fin,omitempty"`
	Moneda      *string    `json:"moneda,omitempty"`
}