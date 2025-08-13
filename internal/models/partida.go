package models

import (
	"time"

	"github.com/google/uuid"
)

type Partida struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ProyectoID  uuid.UUID `json:"proyecto_id" db:"proyecto_id"`
	Codigo      string    `json:"codigo" db:"codigo"`
	Descripcion string    `json:"descripcion" db:"descripcion"`
	Unidad      string    `json:"unidad" db:"unidad"`
	Rendimiento float64   `json:"rendimiento" db:"rendimiento"`
	CostoTotal  float64   `json:"costo_total" db:"costo_total"`
	Activo      bool      `json:"activo" db:"activo"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relaciones
	Proyecto  *Proyecto        `json:"proyecto,omitempty"`
	Recursos  []PartidaRecurso `json:"recursos,omitempty"`
}

type PartidaCompleta struct {
	Partida
	CostoManoObra     float64 `json:"costo_mano_obra" db:"costo_mano_obra"`
	CostoMateriales   float64 `json:"costo_materiales" db:"costo_materiales"`
	CostoEquipos      float64 `json:"costo_equipos" db:"costo_equipos"`
	CostoSubcontratos float64 `json:"costo_subcontratos" db:"costo_subcontratos"`
	ProyectoNombre    string  `json:"proyecto_nombre" db:"proyecto_nombre"`
}

type PartidaCreateRequest struct {
	ProyectoID  uuid.UUID `json:"proyecto_id" validate:"required"`
	Codigo      string    `json:"codigo" validate:"required,min=1,max=50"`
	Descripcion string    `json:"descripcion" validate:"required"`
	Unidad      string    `json:"unidad" validate:"required,max=20"`
	Rendimiento float64   `json:"rendimiento" validate:"min=0"`
}

type PartidaUpdateRequest struct {
	Codigo      *string  `json:"codigo,omitempty"`
	Descripcion *string  `json:"descripcion,omitempty"`
	Unidad      *string  `json:"unidad,omitempty"`
	Rendimiento *float64 `json:"rendimiento,omitempty"`
	Activo      *bool    `json:"activo,omitempty"`
}

// Estructuras para compatibilidad con el JSON original
type PartidaJSON struct {
	Codigo       string        `json:"codigo"`
	Descripcion  string        `json:"descripcion"`
	Unidad       string        `json:"unidad"`
	Rendimiento  float64       `json:"rendimiento"`
	ManoObra     []RecursoJSON `json:"mano_obra"`
	Materiales   []RecursoJSON `json:"materiales"`
	Equipos      []RecursoJSON `json:"equipos"`
	Subcontratos []RecursoJSON `json:"subcontratos"`
}

type RecursoJSON struct {
	Codigo      string  `json:"codigo"`
	Descripcion string  `json:"descripcion"`
	Unidad      string  `json:"unidad"`
	Cuadrilla   float64 `json:"cuadrilla,omitempty"`
	Cantidad    float64 `json:"cantidad"`
	Precio      float64 `json:"precio"`
}