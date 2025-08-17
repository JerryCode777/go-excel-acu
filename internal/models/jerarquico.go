package models

import (
	"time"

	"github.com/google/uuid"
)

// Presupuesto representa un presupuesto principal
type Presupuesto struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Codigo         string     `json:"codigo" db:"codigo"`
	Nombre         string     `json:"nombre" db:"nombre"`
	Cliente        *string    `json:"cliente,omitempty" db:"cliente"`
	Lugar          *string    `json:"lugar,omitempty" db:"lugar"`
	Moneda         string     `json:"moneda" db:"moneda"`
	FechaCreacion  *time.Time `json:"fecha_creacion,omitempty" db:"fecha_creacion"`
	UsuarioID      *uuid.UUID `json:"usuario_id,omitempty" db:"usuario_id"`
	OrganizacionID *uuid.UUID `json:"organizacion_id,omitempty" db:"organizacion_id"`
	Activo         bool       `json:"activo" db:"activo"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Subpresupuesto representa un subpresupuesto dentro de un presupuesto
type Subpresupuesto struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PresupuestoID uuid.UUID `json:"presupuesto_id" db:"presupuesto_id"`
	Codigo        string    `json:"codigo" db:"codigo"`
	Nombre        string    `json:"nombre" db:"nombre"`
	Orden         int       `json:"orden" db:"orden"`
	Activo        bool      `json:"activo" db:"activo"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Titulo representa un título jerárquico (hasta 10 niveles)
type Titulo struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PresupuestoID    uuid.UUID  `json:"presupuesto_id" db:"presupuesto_id"`
	SubpresupuestoID *uuid.UUID `json:"subpresupuesto_id,omitempty" db:"subpresupuesto_id"`
	TituloPadreID    *uuid.UUID `json:"titulo_padre_id,omitempty" db:"titulo_padre_id"`
	Nivel            int        `json:"nivel" db:"nivel"`
	Numero           int        `json:"numero" db:"numero"`
	CodigoCompleto   string     `json:"codigo_completo" db:"codigo_completo"`
	Nombre           string     `json:"nombre" db:"nombre"`
	Orden            int        `json:"orden" db:"orden"`
	Activo           bool       `json:"activo" db:"activo"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// PartidaJerarquica representa una partida con información jerárquica
type PartidaJerarquica struct {
	// Campos de la partida
	PartidaID          uuid.UUID  `json:"partida_id" db:"partida_id"`
	PartidaCodigo      string     `json:"partida_codigo" db:"partida_codigo"`
	PartidaDescripcion string     `json:"partida_descripcion" db:"partida_descripcion"`
	Unidad             string     `json:"unidad" db:"unidad"`
	Rendimiento        float64    `json:"rendimiento" db:"rendimiento"`
	CostoTotal         float64    `json:"costo_total" db:"costo_total"`
	PartidaNumero      int        `json:"partida_numero" db:"partida_numero"`
	PartidaOrden       int        `json:"partida_orden" db:"partida_orden"`
	
	// Información jerárquica
	PresupuestoID         uuid.UUID  `json:"presupuesto_id" db:"presupuesto_id"`
	PresupuestoCodigo     string     `json:"presupuesto_codigo" db:"presupuesto_codigo"`
	PresupuestoNombre     string     `json:"presupuesto_nombre" db:"presupuesto_nombre"`
	SubpresupuestoID      *uuid.UUID `json:"subpresupuesto_id,omitempty" db:"subpresupuesto_id"`
	SubpresupuestoCodigo  *string    `json:"subpresupuesto_codigo,omitempty" db:"subpresupuesto_codigo"`
	SubpresupuestoNombre  *string    `json:"subpresupuesto_nombre,omitempty" db:"subpresupuesto_nombre"`
	TituloID              *uuid.UUID `json:"titulo_id,omitempty" db:"titulo_id"`
	TituloCodigo          *string    `json:"titulo_codigo,omitempty" db:"titulo_codigo"`
	TituloNombre          *string    `json:"titulo_nombre,omitempty" db:"titulo_nombre"`
	TituloNivel           *int       `json:"titulo_nivel,omitempty" db:"titulo_nivel"`
}

// EstructuraJerarquica representa la vista completa de la jerarquía
type EstructuraJerarquica struct {
	PresupuestoID         uuid.UUID  `json:"presupuesto_id" db:"presupuesto_id"`
	PresupuestoCodigo     string     `json:"presupuesto_codigo" db:"presupuesto_codigo"`
	PresupuestoNombre     string     `json:"presupuesto_nombre" db:"presupuesto_nombre"`
	SubpresupuestoID      *uuid.UUID `json:"subpresupuesto_id,omitempty" db:"subpresupuesto_id"`
	SubpresupuestoCodigo  *string    `json:"subpresupuesto_codigo,omitempty" db:"subpresupuesto_codigo"`
	SubpresupuestoNombre  *string    `json:"subpresupuesto_nombre,omitempty" db:"subpresupuesto_nombre"`
	TituloID              *uuid.UUID `json:"titulo_id,omitempty" db:"titulo_id"`
	Nivel                 *int       `json:"nivel,omitempty" db:"nivel"`
	TituloCodigo          *string    `json:"titulo_codigo,omitempty" db:"titulo_codigo"`
	TituloNombre          *string    `json:"titulo_nombre,omitempty" db:"titulo_nombre"`
	Depth                 *int       `json:"depth,omitempty" db:"depth"`
	PathOrden             *string    `json:"path_orden,omitempty" db:"path_orden"`
	TotalPartidas         int64      `json:"total_partidas" db:"total_partidas"`
	CostoTotalTitulos     float64    `json:"costo_total_titulos" db:"costo_total_titulos"`
}

// ResumenJerarquico representa las estadísticas de un presupuesto
type ResumenJerarquico struct {
	TotalSubpresupuestos int64   `json:"total_subpresupuestos" db:"total_subpresupuestos"`
	TotalTitulos         int64   `json:"total_titulos" db:"total_titulos"`
	TotalPartidas        int64   `json:"total_partidas" db:"total_partidas"`
	CostoTotal           float64 `json:"costo_total" db:"costo_total"`
	NivelesMaximos       int     `json:"niveles_maximos" db:"niveles_maximos"`
}

// Requests para API
type PresupuestoRequest struct {
	Codigo  string  `json:"codigo" validate:"required"`
	Nombre  string  `json:"nombre" validate:"required"`
	Cliente *string `json:"cliente,omitempty"`
	Lugar   *string `json:"lugar,omitempty"`
	Moneda  string  `json:"moneda" validate:"required"`
}

type SubpresupuestoRequest struct {
	Codigo string `json:"codigo" validate:"required"`
	Nombre string `json:"nombre" validate:"required"`
	Orden  int    `json:"orden"`
}

type TituloRequest struct {
	Nombre            string     `json:"nombre" validate:"required"`
	SubpresupuestoID  *uuid.UUID `json:"subpresupuesto_id,omitempty"`
	TituloPadreID     *uuid.UUID `json:"titulo_padre_id,omitempty"`
	Orden             int        `json:"orden"`
}

type PartidaJerarquicaRequest struct {
	Descripcion      string     `json:"descripcion" validate:"required"`
	Unidad           string     `json:"unidad" validate:"required"`
	Rendimiento      float64    `json:"rendimiento" validate:"min=0"`
	TituloID         *uuid.UUID `json:"titulo_id,omitempty"`
	SubpresupuestoID *uuid.UUID `json:"subpresupuesto_id,omitempty"`
	Orden            int        `json:"orden"`
}

// Estructura para el parser ACU jerárquico
type ACUJerarquico struct {
	Presupuesto     PresupuestoData     `json:"presupuesto"`
	Subpresupuestos []SubpresupuestoData `json:"subpresupuestos,omitempty"`
	Titulos         []TituloData        `json:"titulos,omitempty"`
	Partidas        []PartidaData       `json:"partidas"`
}

type PresupuestoData struct {
	Codigo  string  `json:"codigo"`
	Nombre  string  `json:"nombre"`
	Cliente *string `json:"cliente,omitempty"`
	Lugar   *string `json:"lugar,omitempty"`
	Moneda  string  `json:"moneda"`
}

type SubpresupuestoData struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
}

type TituloData struct {
	Nivel             int     `json:"nivel"`
	Numero            int     `json:"numero"`
	CodigoCompleto    string  `json:"codigo_completo"`
	Nombre            string  `json:"nombre"`
	TituloPadreCodigo *string `json:"titulo_padre_codigo,omitempty"`
}

type PartidaData struct {
	Codigo       string        `json:"codigo"`
	Descripcion  string        `json:"descripcion"`
	Unidad       string        `json:"unidad"`
	Rendimiento  float64       `json:"rendimiento"`
	ManoObra     []RecursoData `json:"mano_obra,omitempty"`
	Materiales   []RecursoData `json:"materiales,omitempty"`
	Equipos      []RecursoData `json:"equipos,omitempty"`
	Subcontratos []RecursoData `json:"subcontratos,omitempty"`
}

type RecursoData struct {
	Codigo      string   `json:"codigo"`
	Descripcion string   `json:"descripcion"`
	Unidad      string   `json:"unidad"`
	Cantidad    float64  `json:"cantidad"`
	Precio      float64  `json:"precio"`
	Cuadrilla   *float64 `json:"cuadrilla,omitempty"`
}

// Responses para API
type PresupuestoResponse struct {
	Success     bool                    `json:"success"`
	Message     string                  `json:"message,omitempty"`
	Data        *Presupuesto            `json:"data,omitempty"`
	Presupuesto *Presupuesto            `json:"presupuesto,omitempty"`
	Estructura  []EstructuraJerarquica  `json:"estructura,omitempty"`
	Resumen     *ResumenJerarquico      `json:"resumen,omitempty"`
}

type TituloResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Data    *Titulo   `json:"data,omitempty"`
	Titulos []Titulo  `json:"titulos,omitempty"`
}

type PartidaJerarquicaResponse struct {
	Success  bool                 `json:"success"`
	Message  string               `json:"message,omitempty"`
	Data     *PartidaJerarquica   `json:"data,omitempty"`
	Partidas []PartidaJerarquica  `json:"partidas,omitempty"`
}