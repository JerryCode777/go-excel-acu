package services

import (
	"database/sql"
	"fmt"
	"goexcel/internal/models"
)

type HierarchyService struct {
	db *sql.DB
}

func NewHierarchyService(db *sql.DB) *HierarchyService {
	return &HierarchyService{db: db}
}

// ElementoJerarquico representa un elemento en la jerarquía
type ElementoJerarquico struct {
	ID            string  `json:"id"`
	ProyectoID    string  `json:"proyecto_id"`
	Codigo        string  `json:"codigo"`
	Descripcion   string  `json:"descripcion"`
	TipoElemento  string  `json:"tipo_elemento"` // "titulo" o "partida"
	Nivel         int     `json:"nivel"`
	CodigoPadre   *string `json:"codigo_padre"`
	Unidad        *string `json:"unidad"`
	Rendimiento   *float64 `json:"rendimiento"`
	CostoTotal    *float64 `json:"costo_total"`
	OrdenDisplay  int     `json:"orden_display"`
	Hijos         []ElementoJerarquico `json:"hijos,omitempty"`
}

// ObtenerJerarquiaCompleta devuelve la estructura jerárquica completa de un proyecto
func (h *HierarchyService) ObtenerJerarquiaCompleta(proyectoID string) ([]ElementoJerarquico, error) {
	query := `
		SELECT 
			id, proyecto_id, codigo, descripcion, tipo_elemento, 
			nivel, codigo_padre, unidad, rendimiento, costo_total, orden_display
		FROM elementos_jerarquicos 
		WHERE proyecto_id = $1 
		ORDER BY orden_display, codigo
	`

	rows, err := h.db.Query(query, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando jerarquía: %v", err)
	}
	defer rows.Close()

	var elementos []ElementoJerarquico
	for rows.Next() {
		var elem ElementoJerarquico
		err := rows.Scan(
			&elem.ID, &elem.ProyectoID, &elem.Codigo, &elem.Descripcion,
			&elem.TipoElemento, &elem.Nivel, &elem.CodigoPadre,
			&elem.Unidad, &elem.Rendimiento, &elem.CostoTotal, &elem.OrdenDisplay,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando elemento: %v", err)
		}
		elementos = append(elementos, elem)
	}


	// Construir jerarquía recursivamente
	jerarquia := h.construirJerarquia(elementos, nil)
	
	return jerarquia, nil
}

// construirJerarquia construye recursivamente la estructura jerárquica
func (h *HierarchyService) construirJerarquia(elementos []ElementoJerarquico, codigoPadre *string) []ElementoJerarquico {
	var hijos []ElementoJerarquico
	
	for _, elem := range elementos {
		// Verificar si este elemento pertenece al nivel actual
		if (codigoPadre == nil && elem.CodigoPadre == nil) || 
		   (codigoPadre != nil && elem.CodigoPadre != nil && *elem.CodigoPadre == *codigoPadre) {
			
			// Buscar recursivamente los hijos de este elemento
			elem.Hijos = h.construirJerarquia(elementos, &elem.Codigo)
			hijos = append(hijos, elem)
		}
	}
	
	return hijos
}

// ObtenerPartidasConJerarquia devuelve solo las partidas (no títulos) con información de jerarquía
func (h *HierarchyService) ObtenerPartidasConJerarquia(proyectoID string) ([]models.PartidaCompleta, error) {
	query := `
		SELECT 
			p.id, p.codigo, p.descripcion, p.unidad, p.rendimiento,
			COALESCE(mo.total, 0) as costo_mano_obra,
			COALESCE(mat.total, 0) as costo_materiales,
			COALESCE(eq.total, 0) as costo_equipos,
			COALESCE(sub.total, 0) as costo_subcontratos,
			p.costo_total,
			eh.codigo_padre,
			eh.nivel
		FROM partidas p
		LEFT JOIN elementos_jerarquicos eh ON p.elemento_jerarquico_id = eh.id
		LEFT JOIN (
			SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
			FROM partida_recursos pr
			JOIN recursos r ON pr.recurso_id = r.id
			JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
			WHERE tr.nombre = 'mano_obra'
			GROUP BY pr.partida_id
		) mo ON p.id = mo.partida_id
		LEFT JOIN (
			SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
			FROM partida_recursos pr
			JOIN recursos r ON pr.recurso_id = r.id
			JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
			WHERE tr.nombre = 'materiales'
			GROUP BY pr.partida_id
		) mat ON p.id = mat.partida_id
		LEFT JOIN (
			SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
			FROM partida_recursos pr
			JOIN recursos r ON pr.recurso_id = r.id
			JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
			WHERE tr.nombre = 'equipos'
			GROUP BY pr.partida_id
		) eq ON p.id = eq.partida_id
		LEFT JOIN (
			SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
			FROM partida_recursos pr
			JOIN recursos r ON pr.recurso_id = r.id
			JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
			WHERE tr.nombre = 'subcontratos'
			GROUP BY pr.partida_id
		) sub ON p.id = sub.partida_id
		WHERE p.proyecto_id = $1
		ORDER BY eh.orden_display, p.codigo
	`

	rows, err := h.db.Query(query, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando partidas con jerarquía: %v", err)
	}
	defer rows.Close()

	var partidas []models.PartidaCompleta
	for rows.Next() {
		var partida models.PartidaCompleta
		var codigoPadre sql.NullString
		var nivel sql.NullInt32

		err := rows.Scan(
			&partida.ID, &partida.Codigo, &partida.Descripcion,
			&partida.Unidad, &partida.Rendimiento,
			&partida.CostoManoObra, &partida.CostoMateriales,
			&partida.CostoEquipos, &partida.CostoSubcontratos,
			&partida.CostoTotal, &codigoPadre, &nivel,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando partida: %v", err)
		}

		partidas = append(partidas, partida)
	}

	return partidas, nil
}

// ObtenerTitulosJerarquicos devuelve solo los títulos organizacionales
func (h *HierarchyService) ObtenerTitulosJerarquicos(proyectoID string) ([]ElementoJerarquico, error) {
	query := `
		SELECT 
			id, proyecto_id, codigo, descripcion, tipo_elemento, 
			nivel, codigo_padre, orden_display
		FROM elementos_jerarquicos 
		WHERE proyecto_id = $1 AND tipo_elemento = 'titulo'
		ORDER BY orden_display, codigo
	`

	rows, err := h.db.Query(query, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando títulos: %v", err)
	}
	defer rows.Close()

	var titulos []ElementoJerarquico
	for rows.Next() {
		var titulo ElementoJerarquico
		err := rows.Scan(
			&titulo.ID, &titulo.ProyectoID, &titulo.Codigo, &titulo.Descripcion,
			&titulo.TipoElemento, &titulo.Nivel, &titulo.CodigoPadre, &titulo.OrdenDisplay,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando título: %v", err)
		}
		titulos = append(titulos, titulo)
	}

	return titulos, nil
}

// CrearPartidaConJerarquia crea una nueva partida y su jerarquía automáticamente
func (h *HierarchyService) CrearPartidaConJerarquia(proyectoID, codigo, descripcion, unidad string, rendimiento float64) error {
	tx, err := h.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %v", err)
	}
	defer tx.Rollback()

	// Usar la función de PostgreSQL para crear la jerarquía
	_, err = tx.Exec(`
		SELECT insertar_elemento_con_jerarquia($1, $2, $3, $4, $5)
	`, proyectoID, codigo, descripcion, unidad, rendimiento)
	
	if err != nil {
		return fmt.Errorf("error creando elemento jerárquico: %v", err)
	}

	return tx.Commit()
}

// ActualizarTitulosPersonalizados permite personalizar los títulos generados automáticamente
func (h *HierarchyService) ActualizarTitulosPersonalizados(proyectoID string, titulosPersonalizados map[string]string) error {
	tx, err := h.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %v", err)
	}
	defer tx.Rollback()

	for codigo, descripcion := range titulosPersonalizados {
		_, err = tx.Exec(`
			UPDATE elementos_jerarquicos 
			SET descripcion = $1, updated_at = CURRENT_TIMESTAMP
			WHERE proyecto_id = $2 AND codigo = $3 AND tipo_elemento = 'titulo'
		`, descripcion, proyectoID, codigo)
		
		if err != nil {
			return fmt.Errorf("error actualizando título %s: %v", codigo, err)
		}
	}

	return tx.Commit()
}