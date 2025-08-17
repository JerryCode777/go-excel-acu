package models

import (
	"time"

	"github.com/google/uuid"
)

// MetradoPartida representa un metrado específico de una partida en un proyecto
type MetradoPartida struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ProyectoID    uuid.UUID  `json:"proyecto_id" db:"proyecto_id"`
	PartidaCodigo string     `json:"partida_codigo" db:"partida_codigo"`
	Metrado       float64    `json:"metrado" db:"metrado"`
	Unidad        *string    `json:"unidad,omitempty" db:"unidad"`
	Observaciones *string    `json:"observaciones,omitempty" db:"observaciones"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// MetradoCompleto representa un metrado con información completa de la partida
type MetradoCompleto struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	ProyectoID           uuid.UUID  `json:"proyecto_id" db:"proyecto_id"`
	PartidaCodigo        string     `json:"partida_codigo" db:"partida_codigo"`
	Metrado              float64    `json:"metrado" db:"metrado"`
	MetradoUnidad        *string    `json:"metrado_unidad,omitempty" db:"metrado_unidad"`
	Observaciones        *string    `json:"observaciones,omitempty" db:"observaciones"`
	PartidaDescripcion   *string    `json:"partida_descripcion,omitempty" db:"partida_descripcion"`
	PartidaUnidad        *string    `json:"partida_unidad,omitempty" db:"partida_unidad"`
	CostoUnitario        *float64   `json:"costo_unitario,omitempty" db:"costo_unitario"`
	CostoTotalPartida    *float64   `json:"costo_total_partida,omitempty" db:"costo_total_partida"`
	ProyectoNombre       *string    `json:"proyecto_nombre,omitempty" db:"proyecto_nombre"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// ResumenProyecto representa el resumen financiero de un proyecto
type ResumenProyecto struct {
	TotalPartidas       int64   `json:"total_partidas" db:"total_partidas"`
	CostoDirecto        float64 `json:"costo_directo" db:"costo_directo"`
	PartidasConMetrado  int64   `json:"partidas_con_metrado" db:"partidas_con_metrado"`
	PartidasSinMetrado  int64   `json:"partidas_sin_metrado" db:"partidas_sin_metrado"`
}

// MetradoRequest representa la estructura para crear/actualizar metrados
type MetradoRequest struct {
	PartidaCodigo string  `json:"partida_codigo" validate:"required"`
	Metrado       float64 `json:"metrado" validate:"min=0"`
	Unidad        *string `json:"unidad,omitempty"`
	Observaciones *string `json:"observaciones,omitempty"`
}

// MetradosLoteRequest representa una solicitud para actualizar múltiples metrados
type MetradosLoteRequest struct {
	Metrados []MetradoRequest `json:"metrados" validate:"required,dive"`
}

// MetradoResponse representa la respuesta de la API para metrados
type MetradoResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    *MetradoCompleto  `json:"data,omitempty"`
	Metrados []MetradoCompleto `json:"metrados,omitempty"`
}

// ResumenProyectoResponse representa la respuesta de la API para el resumen del proyecto
type ResumenProyectoResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message,omitempty"`
	Data    *ResumenProyecto `json:"data,omitempty"`
}