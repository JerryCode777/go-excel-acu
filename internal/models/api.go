package models

// Request structures for API
type ProyectoRequest struct {
	Nombre      string `json:"nombre" validate:"required"`
	Descripcion string `json:"descripcion"`
	Moneda      string `json:"moneda"`
}

type PartidaRequest struct {
	Codigo       string           `json:"codigo" validate:"required"`
	Descripcion  string           `json:"descripcion" validate:"required"`
	Unidad       string           `json:"unidad" validate:"required"`
	Rendimiento  float64          `json:"rendimiento" validate:"min=0"`
	ManoObra     []RecursoRequest `json:"mano_obra,omitempty"`
	Materiales   []RecursoRequest `json:"materiales,omitempty"`
	Equipos      []RecursoRequest `json:"equipos,omitempty"`
	Subcontratos []RecursoRequest `json:"subcontratos,omitempty"`
}

type RecursoRequest struct {
	Codigo      string   `json:"codigo" validate:"required"`
	Descripcion string   `json:"descripcion" validate:"required"`
	Unidad      string   `json:"unidad" validate:"required"`
	Cantidad    float64  `json:"cantidad" validate:"min=0"`
	Precio      float64  `json:"precio" validate:"min=0"`
	Cuadrilla   *float64 `json:"cuadrilla,omitempty"`
}

// Response structures for API
type ProyectoResponse struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Moneda      string `json:"moneda"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type PartidaResponse struct {
	ID           string            `json:"id"`
	Codigo       string            `json:"codigo"`
	Descripcion  string            `json:"descripcion"`
	Unidad       string            `json:"unidad"`
	Rendimiento  float64           `json:"rendimiento"`
	CostoTotal   float64           `json:"costo_total"`
	ManoObra     []RecursoResponse `json:"mano_obra"`
	Materiales   []RecursoResponse `json:"materiales"`
	Equipos      []RecursoResponse `json:"equipos"`
	Subcontratos []RecursoResponse `json:"subcontratos"`
}

type RecursoResponse struct {
	ID          string   `json:"id"`
	Codigo      string   `json:"codigo"`
	Descripcion string   `json:"descripcion"`
	Unidad      string   `json:"unidad"`
	Cantidad    float64  `json:"cantidad"`
	Precio      float64  `json:"precio"`
	Cuadrilla   *float64 `json:"cuadrilla,omitempty"`
	Parcial     float64  `json:"parcial"`
}

type ProjectStats struct {
	TotalPartidas     int     `json:"total_partidas"`
	TotalRecursos     int     `json:"total_recursos"`
	CostoTotal        float64 `json:"costo_total"`
	CostoManoObra     float64 `json:"costo_mano_obra"`
	CostoMateriales   float64 `json:"costo_materiales"`
	CostoEquipos      float64 `json:"costo_equipos"`
	CostoSubcontratos float64 `json:"costo_subcontratos"`
}

// Error response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// Validation structures
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationResponse struct {
	Success bool              `json:"success"`
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// Health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Database  string `json:"database"`
}