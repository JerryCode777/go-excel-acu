package repositories

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"goexcel/internal/models"
)

// PresupuestoRepository maneja las operaciones de base de datos para presupuestos jerárquicos
type PresupuestoRepository struct {
	db *sql.DB
}

// NewPresupuestoRepository crea una nueva instancia del repositorio de presupuestos
func NewPresupuestoRepository(db *sql.DB) *PresupuestoRepository {
	return &PresupuestoRepository{db: db}
}

// CrearPresupuesto crea un nuevo presupuesto
func (r *PresupuestoRepository) CrearPresupuesto(req models.PresupuestoRequest, usuarioID *uuid.UUID) (*models.Presupuesto, error) {
	query := `
		INSERT INTO presupuestos (codigo, nombre, cliente, lugar, moneda, usuario_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, codigo, nombre, cliente, lugar, moneda, fecha_creacion, usuario_id, organizacion_id, activo, created_at, updated_at`

	var presupuesto models.Presupuesto
	err := r.db.QueryRow(query, req.Codigo, req.Nombre, req.Cliente, req.Lugar, req.Moneda, usuarioID).Scan(
		&presupuesto.ID,
		&presupuesto.Codigo,
		&presupuesto.Nombre,
		&presupuesto.Cliente,
		&presupuesto.Lugar,
		&presupuesto.Moneda,
		&presupuesto.FechaCreacion,
		&presupuesto.UsuarioID,
		&presupuesto.OrganizacionID,
		&presupuesto.Activo,
		&presupuesto.CreatedAt,
		&presupuesto.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando presupuesto: %v", err)
	}

	return &presupuesto, nil
}

// ObtenerPresupuesto obtiene un presupuesto por ID
func (r *PresupuestoRepository) ObtenerPresupuesto(id uuid.UUID) (*models.Presupuesto, error) {
	query := `
		SELECT id, codigo, nombre, cliente, lugar, moneda, fecha_creacion, usuario_id, organizacion_id, activo, created_at, updated_at
		FROM presupuestos 
		WHERE id = $1 AND activo = true`

	var presupuesto models.Presupuesto
	err := r.db.QueryRow(query, id).Scan(
		&presupuesto.ID,
		&presupuesto.Codigo,
		&presupuesto.Nombre,
		&presupuesto.Cliente,
		&presupuesto.Lugar,
		&presupuesto.Moneda,
		&presupuesto.FechaCreacion,
		&presupuesto.UsuarioID,
		&presupuesto.OrganizacionID,
		&presupuesto.Activo,
		&presupuesto.CreatedAt,
		&presupuesto.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("presupuesto no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo presupuesto: %v", err)
	}

	return &presupuesto, nil
}

// CrearSubpresupuesto crea un nuevo subpresupuesto
func (r *PresupuestoRepository) CrearSubpresupuesto(presupuestoID uuid.UUID, req models.SubpresupuestoRequest) (*models.Subpresupuesto, error) {
	query := `
		INSERT INTO subpresupuestos (presupuesto_id, codigo, nombre, orden)
		VALUES ($1, $2, $3, $4)
		RETURNING id, presupuesto_id, codigo, nombre, orden, activo, created_at, updated_at`

	var subpresupuesto models.Subpresupuesto
	err := r.db.QueryRow(query, presupuestoID, req.Codigo, req.Nombre, req.Orden).Scan(
		&subpresupuesto.ID,
		&subpresupuesto.PresupuestoID,
		&subpresupuesto.Codigo,
		&subpresupuesto.Nombre,
		&subpresupuesto.Orden,
		&subpresupuesto.Activo,
		&subpresupuesto.CreatedAt,
		&subpresupuesto.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando subpresupuesto: %v", err)
	}

	return &subpresupuesto, nil
}

// CrearTitulo crea un nuevo título jerárquico
func (r *PresupuestoRepository) CrearTitulo(presupuestoID uuid.UUID, req models.TituloRequest) (*models.Titulo, error) {
	// Calcular nivel y código automáticamente
	nivel := 1
	var codigoCompleto string
	var err error

	if req.TituloPadreID != nil {
		// Obtener nivel del padre + 1
		var nivelPadre int
		err = r.db.QueryRow("SELECT nivel FROM titulos WHERE id = $1", req.TituloPadreID).Scan(&nivelPadre)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo título padre: %v", err)
		}
		nivel = nivelPadre + 1
	}

	// Generar código jerárquico
	if req.TituloPadreID != nil {
		codigoCompleto, err = r.generarCodigoConPadre(*req.TituloPadreID)
	} else if req.SubpresupuestoID != nil {
		codigoCompleto, err = r.generarCodigoEnSubpresupuesto(presupuestoID, *req.SubpresupuestoID)
	} else {
		codigoCompleto, err = r.generarCodigoEnPresupuesto(presupuestoID)
	}

	if err != nil {
		return nil, fmt.Errorf("error generando código: %v", err)
	}

	// Obtener siguiente número para este nivel
	numero, err := r.obtenerSiguienteNumero(presupuestoID, req.SubpresupuestoID, req.TituloPadreID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo siguiente número: %v", err)
	}

	query := `
		INSERT INTO titulos (presupuesto_id, subpresupuesto_id, titulo_padre_id, nivel, numero, codigo_completo, nombre, orden)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, presupuesto_id, subpresupuesto_id, titulo_padre_id, nivel, numero, codigo_completo, nombre, orden, activo, created_at, updated_at`

	var titulo models.Titulo
	err = r.db.QueryRow(query, presupuestoID, req.SubpresupuestoID, req.TituloPadreID, nivel, numero, codigoCompleto, req.Nombre, req.Orden).Scan(
		&titulo.ID,
		&titulo.PresupuestoID,
		&titulo.SubpresupuestoID,
		&titulo.TituloPadreID,
		&titulo.Nivel,
		&titulo.Numero,
		&titulo.CodigoCompleto,
		&titulo.Nombre,
		&titulo.Orden,
		&titulo.Activo,
		&titulo.CreatedAt,
		&titulo.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando título: %v", err)
	}

	return &titulo, nil
}

// generarCodigoConPadre genera código para un título con padre
func (r *PresupuestoRepository) generarCodigoConPadre(padreID uuid.UUID) (string, error) {
	var codigoPadre string
	var siguienteNumero int

	// Obtener código del padre
	err := r.db.QueryRow("SELECT codigo_completo FROM titulos WHERE id = $1", padreID).Scan(&codigoPadre)
	if err != nil {
		return "", fmt.Errorf("error obteniendo código padre: %v", err)
	}

	// Obtener siguiente número en este nivel
	err = r.db.QueryRow("SELECT COALESCE(MAX(numero), 0) + 1 FROM titulos WHERE titulo_padre_id = $1", padreID).Scan(&siguienteNumero)
	if err != nil {
		return "", fmt.Errorf("error obteniendo siguiente número: %v", err)
	}

	return fmt.Sprintf("%s.%02d", codigoPadre, siguienteNumero), nil
}

// generarCodigoEnSubpresupuesto genera código para título en subpresupuesto
func (r *PresupuestoRepository) generarCodigoEnSubpresupuesto(presupuestoID, subpresupuestoID uuid.UUID) (string, error) {
	var siguienteNumero int

	err := r.db.QueryRow(`
		SELECT COALESCE(MAX(numero), 0) + 1 
		FROM titulos 
		WHERE presupuesto_id = $1 AND subpresupuesto_id = $2 AND titulo_padre_id IS NULL`,
		presupuestoID, subpresupuestoID).Scan(&siguienteNumero)
	
	if err != nil {
		return "", fmt.Errorf("error obteniendo siguiente número: %v", err)
	}

	return fmt.Sprintf("%02d", siguienteNumero), nil
}

// generarCodigoEnPresupuesto genera código para título en presupuesto
func (r *PresupuestoRepository) generarCodigoEnPresupuesto(presupuestoID uuid.UUID) (string, error) {
	var siguienteNumero int

	err := r.db.QueryRow(`
		SELECT COALESCE(MAX(numero), 0) + 1 
		FROM titulos 
		WHERE presupuesto_id = $1 AND titulo_padre_id IS NULL AND subpresupuesto_id IS NULL`,
		presupuestoID).Scan(&siguienteNumero)
	
	if err != nil {
		return "", fmt.Errorf("error obteniendo siguiente número: %v", err)
	}

	return fmt.Sprintf("%02d", siguienteNumero), nil
}

// obtenerSiguienteNumero obtiene el siguiente número para un título
func (r *PresupuestoRepository) obtenerSiguienteNumero(presupuestoID uuid.UUID, subpresupuestoID, tituloPadreID *uuid.UUID) (int, error) {
	var query string
	var args []interface{}

	if tituloPadreID != nil {
		query = "SELECT COALESCE(MAX(numero), 0) + 1 FROM titulos WHERE titulo_padre_id = $1"
		args = []interface{}{*tituloPadreID}
	} else if subpresupuestoID != nil {
		query = "SELECT COALESCE(MAX(numero), 0) + 1 FROM titulos WHERE presupuesto_id = $1 AND subpresupuesto_id = $2 AND titulo_padre_id IS NULL"
		args = []interface{}{presupuestoID, *subpresupuestoID}
	} else {
		query = "SELECT COALESCE(MAX(numero), 0) + 1 FROM titulos WHERE presupuesto_id = $1 AND titulo_padre_id IS NULL AND subpresupuesto_id IS NULL"
		args = []interface{}{presupuestoID}
	}

	var numero int
	err := r.db.QueryRow(query, args...).Scan(&numero)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo siguiente número: %v", err)
	}

	return numero, nil
}

// ObtenerEstructuraJerarquica obtiene la estructura completa de un presupuesto
func (r *PresupuestoRepository) ObtenerEstructuraJerarquica(presupuestoID uuid.UUID) ([]models.EstructuraJerarquica, error) {
	query := `SELECT * FROM vista_estructura_jerarquica WHERE presupuesto_id = $1 ORDER BY path_orden`

	rows, err := r.db.Query(query, presupuestoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando estructura jerárquica: %v", err)
	}
	defer rows.Close()

	var estructura []models.EstructuraJerarquica
	for rows.Next() {
		var item models.EstructuraJerarquica
		err := rows.Scan(
			&item.PresupuestoID,
			&item.PresupuestoCodigo,
			&item.PresupuestoNombre,
			&item.SubpresupuestoID,
			&item.SubpresupuestoCodigo,
			&item.SubpresupuestoNombre,
			&item.TituloID,
			&item.Nivel,
			&item.TituloCodigo,
			&item.TituloNombre,
			&item.Depth,
			&item.PathOrden,
			&item.TotalPartidas,
			&item.CostoTotalTitulos,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando estructura jerárquica: %v", err)
		}
		estructura = append(estructura, item)
	}

	return estructura, nil
}

// ObtenerPartidasJerarquicas obtiene todas las partidas con su información jerárquica
func (r *PresupuestoRepository) ObtenerPartidasJerarquicas(presupuestoID uuid.UUID) ([]models.PartidaJerarquica, error) {
	query := `SELECT * FROM vista_partidas_jerarquicas WHERE presupuesto_id = $1 ORDER BY titulo_codigo, partida_orden`

	rows, err := r.db.Query(query, presupuestoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando partidas jerárquicas: %v", err)
	}
	defer rows.Close()

	var partidas []models.PartidaJerarquica
	for rows.Next() {
		var partida models.PartidaJerarquica
		err := rows.Scan(
			&partida.PartidaID,
			&partida.PartidaCodigo,
			&partida.PartidaDescripcion,
			&partida.Unidad,
			&partida.Rendimiento,
			&partida.CostoTotal,
			&partida.PartidaNumero,
			&partida.PartidaOrden,
			&partida.PresupuestoID,
			&partida.PresupuestoCodigo,
			&partida.PresupuestoNombre,
			&partida.SubpresupuestoID,
			&partida.SubpresupuestoCodigo,
			&partida.SubpresupuestoNombre,
			&partida.TituloID,
			&partida.TituloCodigo,
			&partida.TituloNombre,
			&partida.TituloNivel,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando partida jerárquica: %v", err)
		}
		partidas = append(partidas, partida)
	}

	return partidas, nil
}

// ObtenerResumenJerarquico obtiene estadísticas del presupuesto
func (r *PresupuestoRepository) ObtenerResumenJerarquico(presupuestoID uuid.UUID) (*models.ResumenJerarquico, error) {
	query := `SELECT * FROM obtener_resumen_jerarquico($1)`

	var resumen models.ResumenJerarquico
	err := r.db.QueryRow(query, presupuestoID).Scan(
		&resumen.TotalSubpresupuestos,
		&resumen.TotalTitulos,
		&resumen.TotalPartidas,
		&resumen.CostoTotal,
		&resumen.NivelesMaximos,
	)

	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen jerárquico: %v", err)
	}

	return &resumen, nil
}

// CrearPartidaJerarquica crea una nueva partida en la estructura jerárquica
func (r *PresupuestoRepository) CrearPartidaJerarquica(presupuestoID uuid.UUID, req models.PartidaJerarquicaRequest) (*models.PartidaJerarquica, error) {
	// Generar código automático
	codigo, err := r.generarCodigoPartida(presupuestoID, req.TituloID)
	if err != nil {
		return nil, fmt.Errorf("error generando código de partida: %v", err)
	}

	// Obtener siguiente número
	numero, err := r.obtenerSiguienteNumeroPartida(presupuestoID, req.TituloID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo siguiente número de partida: %v", err)
	}

	query := `
		INSERT INTO partidas (presupuesto_id, subpresupuesto_id, titulo_id, codigo, descripcion, unidad, rendimiento, numero, orden)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var partidaID uuid.UUID
	err = r.db.QueryRow(query, presupuestoID, req.SubpresupuestoID, req.TituloID, codigo, req.Descripcion, req.Unidad, req.Rendimiento, numero, req.Orden).Scan(&partidaID)
	if err != nil {
		return nil, fmt.Errorf("error creando partida jerárquica: %v", err)
	}

	// Obtener la partida creada con información jerárquica
	partidas, err := r.ObtenerPartidasJerarquicas(presupuestoID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo partida creada: %v", err)
	}

	for _, partida := range partidas {
		if partida.PartidaID == partidaID {
			return &partida, nil
		}
	}

	return nil, fmt.Errorf("partida creada no encontrada")
}

// generarCodigoPartida genera código automático para una partida
func (r *PresupuestoRepository) generarCodigoPartida(presupuestoID uuid.UUID, tituloID *uuid.UUID) (string, error) {
	if tituloID != nil {
		var codigoTitulo string
		var siguienteNumero int

		// Obtener código del título
		err := r.db.QueryRow("SELECT codigo_completo FROM titulos WHERE id = $1", tituloID).Scan(&codigoTitulo)
		if err != nil {
			return "", fmt.Errorf("error obteniendo código del título: %v", err)
		}

		// Obtener siguiente número de partida en este título
		err = r.db.QueryRow("SELECT COALESCE(MAX(numero), 0) + 1 FROM partidas WHERE titulo_id = $1", tituloID).Scan(&siguienteNumero)
		if err != nil {
			return "", fmt.Errorf("error obteniendo siguiente número de partida: %v", err)
		}

		return fmt.Sprintf("%s.%02d", codigoTitulo, siguienteNumero), nil
	}

	// Partida sin título (no recomendado)
	var siguienteNumero int
	err := r.db.QueryRow("SELECT COALESCE(MAX(numero), 0) + 1 FROM partidas WHERE presupuesto_id = $1 AND titulo_id IS NULL", presupuestoID).Scan(&siguienteNumero)
	if err != nil {
		return "", fmt.Errorf("error obteniendo siguiente número de partida: %v", err)
	}

	return fmt.Sprintf("%02d", siguienteNumero), nil
}

// obtenerSiguienteNumeroPartida obtiene el siguiente número para una partida
func (r *PresupuestoRepository) obtenerSiguienteNumeroPartida(presupuestoID uuid.UUID, tituloID *uuid.UUID) (int, error) {
	var numero int
	var err error

	if tituloID != nil {
		err = r.db.QueryRow("SELECT COALESCE(MAX(numero), 0) + 1 FROM partidas WHERE titulo_id = $1", tituloID).Scan(&numero)
	} else {
		err = r.db.QueryRow("SELECT COALESCE(MAX(numero), 0) + 1 FROM partidas WHERE presupuesto_id = $1 AND titulo_id IS NULL", presupuestoID).Scan(&numero)
	}

	if err != nil {
		return 0, fmt.Errorf("error obteniendo siguiente número de partida: %v", err)
	}

	return numero, nil
}