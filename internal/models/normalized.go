package models

// Estructuras para datos normalizados
type NormalizedData struct {
	Proyecto  ProyectoNormalizado  `json:"proyecto"`
	Recursos  []RecursoNormalizado `json:"recursos"`
	Partidas  []PartidaNormalizada `json:"partidas"`
	Relaciones []RelacionNormalizada `json:"relaciones"`
}

type ProyectoNormalizado struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion,omitempty"`
	Moneda      string `json:"moneda"`
}

type RecursoNormalizado struct {
	ID            string  `json:"id"`
	Codigo        string  `json:"codigo"`
	Descripcion   string  `json:"descripcion"`
	Unidad        string  `json:"unidad"`
	PrecioBase    float64 `json:"precio_base"`
	TipoRecurso   string  `json:"tipo_recurso"` // mano_obra, materiales, equipos, subcontratos
}

type PartidaNormalizada struct {
	ID          string  `json:"id"`
	ProyectoID  string  `json:"proyecto_id"`
	Codigo      string  `json:"codigo"`
	Descripcion string  `json:"descripcion"`
	Unidad      string  `json:"unidad"`
	Rendimiento float64 `json:"rendimiento"`
}

type RelacionNormalizada struct {
	ID        string   `json:"id"`
	PartidaID string   `json:"partida_id"`
	RecursoID string   `json:"recurso_id"`
	Cantidad  float64  `json:"cantidad"`
	Precio    float64  `json:"precio"`
	Cuadrilla *float64 `json:"cuadrilla,omitempty"`
}