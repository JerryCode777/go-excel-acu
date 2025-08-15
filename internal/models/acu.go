package models

// Estructuras para formato .acu
type ACUProject struct {
	ID          string      `json:"id"`
	Nombre      string      `json:"nombre"`
	Descripcion string      `json:"descripcion"`
	Moneda      string      `json:"moneda"`
	Partidas    []ACUPartida `json:"partidas"`
}

type ACUPartida struct {
	ID           string        `json:"id"`
	Codigo       string        `json:"codigo"`
	Descripcion  string        `json:"descripcion"`
	Unidad       string        `json:"unidad"`
	Rendimiento  float64       `json:"rendimiento"`
	ManoObra     []ACURecurso  `json:"mano_obra,omitempty"`
	Materiales   []ACURecurso  `json:"materiales,omitempty"`
	Equipos      []ACURecurso  `json:"equipos,omitempty"`
	Subcontratos []ACURecurso  `json:"subcontratos,omitempty"`
}

type ACURecurso struct {
	Codigo      string   `json:"codigo"`
	Descripcion string   `json:"descripcion"`
	Unidad      string   `json:"unidad"`
	Cantidad    float64  `json:"cantidad"`
	Precio      float64  `json:"precio"`
	Cuadrilla   *float64 `json:"cuadrilla,omitempty"`
}

// Token types para el parser
type TokenType int

const (
	// Literals
	TOKEN_AT TokenType = iota
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_EQUALS
	TOKEN_COMMA
	TOKEN_SEMICOLON
	TOKEN_STRING
	TOKEN_NUMBER
	TOKEN_IDENTIFIER
	TOKEN_EOF
	TOKEN_ILLEGAL
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Estructura del parser
type ACUParser struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
	tokens       []Token
	current      int
}