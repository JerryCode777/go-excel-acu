package models

import (
	"time"

	"github.com/google/uuid"
)

type Proyecto struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Nombre            string     `json:"nombre" db:"nombre"`
	Descripcion       *string    `json:"descripcion" db:"descripcion"`
	Ubicacion         *string    `json:"ubicacion" db:"ubicacion"`
	Cliente           *string    `json:"cliente" db:"cliente"`
	FechaInicio       *time.Time `json:"fecha_inicio" db:"fecha_inicio"`
	FechaFin          *time.Time `json:"fecha_fin" db:"fecha_fin"`
	Moneda            string     `json:"moneda" db:"moneda"`
	UsuarioID         *uuid.UUID `json:"usuario_id" db:"usuario_id"`
	OrganizacionID    *uuid.UUID `json:"organizacion_id" db:"organizacion_id"`
	Visibility        string     `json:"visibility" db:"visibility"`
	TemplateCategoria *string    `json:"template_categoria" db:"template_categoria"`
	ImagenPortada     *string    `json:"imagen_portada" db:"imagen_portada"`
	LikesCount        int        `json:"likes_count" db:"likes_count"`
	VistasCount       int        `json:"vistas_count" db:"vistas_count"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	
	// Relaciones
	Usuario      *Usuario      `json:"usuario,omitempty"`
	Organizacion *Organizacion `json:"organizacion,omitempty"`
	Partidas     []Partida     `json:"partidas,omitempty"`
	IsLiked      bool          `json:"is_liked,omitempty"` // Para indicar si el usuario actual le dio like
}

type ProyectoCreateRequest struct {
	Nombre            string     `json:"nombre" validate:"required,min=1,max=255"`
	Descripcion       *string    `json:"descripcion"`
	Ubicacion         *string    `json:"ubicacion"`
	Cliente           *string    `json:"cliente"`
	FechaInicio       *time.Time `json:"fecha_inicio"`
	FechaFin          *time.Time `json:"fecha_fin"`
	Moneda            string     `json:"moneda" validate:"required,len=3"`
	TemplateCategoria *string    `json:"template_categoria"`
	ImagenPortada     *string    `json:"imagen_portada"`
}

type ProyectoUpdateRequest struct {
	Nombre            *string    `json:"nombre,omitempty"`
	Descripcion       *string    `json:"descripcion,omitempty"`
	Ubicacion         *string    `json:"ubicacion,omitempty"`
	Cliente           *string    `json:"cliente,omitempty"`
	FechaInicio       *time.Time `json:"fecha_inicio,omitempty"`
	FechaFin          *time.Time `json:"fecha_fin,omitempty"`
	Moneda            *string    `json:"moneda,omitempty"`
	TemplateCategoria *string    `json:"template_categoria,omitempty"`
	ImagenPortada     *string    `json:"imagen_portada,omitempty"`
	Visibility        *string    `json:"visibility,omitempty"`
}

type ProyectoLike struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ProyectoID uuid.UUID `json:"proyecto_id" db:"proyecto_id"`
	UsuarioID  uuid.UUID `json:"usuario_id" db:"usuario_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}