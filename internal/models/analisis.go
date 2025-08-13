package models

import (
	"time"

	"github.com/google/uuid"
)

type AnalisisHistorico struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	ProyectoID              uuid.UUID `json:"proyecto_id" db:"proyecto_id"`
	NombreArchivo           *string   `json:"nombre_archivo" db:"nombre_archivo"`
	FechaAnalisis          time.Time `json:"fecha_analisis" db:"fecha_analisis"`
	TotalPartidas          int       `json:"total_partidas" db:"total_partidas"`
	CostoTotalManoObra     float64   `json:"costo_total_mano_obra" db:"costo_total_mano_obra"`
	CostoTotalMateriales   float64   `json:"costo_total_materiales" db:"costo_total_materiales"`
	CostoTotalEquipos      float64   `json:"costo_total_equipos" db:"costo_total_equipos"`
	CostoTotalSubcontratos float64   `json:"costo_total_subcontratos" db:"costo_total_subcontratos"`
	CostoTotalProyecto     float64   `json:"costo_total_proyecto" db:"costo_total_proyecto"`
	ArchivoExcelURL        *string   `json:"archivo_excel_url" db:"archivo_excel_url"`

	// Relaciones
	Proyecto *Proyecto `json:"proyecto,omitempty"`
}

type ResumenCostos struct {
	TotalPartidas          int     `json:"total_partidas"`
	CostoTotalManoObra     float64 `json:"costo_total_mano_obra"`
	CostoTotalMateriales   float64 `json:"costo_total_materiales"`
	CostoTotalEquipos      float64 `json:"costo_total_equipos"`
	CostoTotalSubcontratos float64 `json:"costo_total_subcontratos"`
	CostoTotalProyecto     float64 `json:"costo_total_proyecto"`
}

type AnalisisRequest struct {
	ProyectoID    uuid.UUID `json:"proyecto_id" validate:"required"`
	NombreArchivo *string   `json:"nombre_archivo,omitempty"`
}