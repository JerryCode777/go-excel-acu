package repositories

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"goexcel/internal/models"
)

// MetradoRepository maneja las operaciones de base de datos para metrados
type MetradoRepository struct {
	db *sql.DB
}

// NewMetradoRepository crea una nueva instancia del repositorio de metrados
func NewMetradoRepository(db *sql.DB) *MetradoRepository {
	return &MetradoRepository{db: db}
}

// CrearMetrado crea un nuevo metrado para una partida en un proyecto
func (r *MetradoRepository) CrearMetrado(proyectoID uuid.UUID, metrado models.MetradoRequest) (*models.MetradoCompleto, error) {
	query := `
		INSERT INTO metrados_partidas (proyecto_id, partida_codigo, metrado, unidad, observaciones)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (proyecto_id, partida_codigo)
		DO UPDATE SET 
			metrado = EXCLUDED.metrado,
			unidad = EXCLUDED.unidad,
			observaciones = EXCLUDED.observaciones,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id`

	var metradoID uuid.UUID
	err := r.db.QueryRow(query, proyectoID, metrado.PartidaCodigo, metrado.Metrado, metrado.Unidad, metrado.Observaciones).Scan(&metradoID)
	if err != nil {
		return nil, fmt.Errorf("error creando/actualizando metrado: %v", err)
	}

	// Obtener el metrado completo
	return r.ObtenerMetradoPorID(metradoID)
}

// ActualizarMetrados actualiza múltiples metrados en una sola transacción
func (r *MetradoRepository) ActualizarMetrados(proyectoID uuid.UUID, metrados []models.MetradoRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %v", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO metrados_partidas (proyecto_id, partida_codigo, metrado, unidad, observaciones)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (proyecto_id, partida_codigo)
		DO UPDATE SET 
			metrado = EXCLUDED.metrado,
			unidad = EXCLUDED.unidad,
			observaciones = EXCLUDED.observaciones,
			updated_at = CURRENT_TIMESTAMP`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparando statement: %v", err)
	}
	defer stmt.Close()

	for _, metrado := range metrados {
		_, err = stmt.Exec(proyectoID, metrado.PartidaCodigo, metrado.Metrado, metrado.Unidad, metrado.Observaciones)
		if err != nil {
			return fmt.Errorf("error actualizando metrado %s: %v", metrado.PartidaCodigo, err)
		}
	}

	return tx.Commit()
}

// ObtenerMetradoPorID obtiene un metrado específico por su ID
func (r *MetradoRepository) ObtenerMetradoPorID(id uuid.UUID) (*models.MetradoCompleto, error) {
	query := `SELECT * FROM vista_metrados_completos WHERE id = $1`

	var metrado models.MetradoCompleto
	err := r.db.QueryRow(query, id).Scan(
		&metrado.ID,
		&metrado.ProyectoID,
		&metrado.PartidaCodigo,
		&metrado.Metrado,
		&metrado.MetradoUnidad,
		&metrado.Observaciones,
		&metrado.PartidaDescripcion,
		&metrado.PartidaUnidad,
		&metrado.CostoUnitario,
		&metrado.CostoTotalPartida,
		&metrado.ProyectoNombre,
		&metrado.CreatedAt,
		&metrado.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metrado no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo metrado: %v", err)
	}

	return &metrado, nil
}

// ObtenerMetradosPorProyecto obtiene todos los metrados de un proyecto
func (r *MetradoRepository) ObtenerMetradosPorProyecto(proyectoID uuid.UUID) ([]models.MetradoCompleto, error) {
	query := `SELECT * FROM vista_metrados_completos WHERE proyecto_id = $1 ORDER BY partida_codigo`

	rows, err := r.db.Query(query, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando metrados: %v", err)
	}
	defer rows.Close()

	var metrados []models.MetradoCompleto
	for rows.Next() {
		var metrado models.MetradoCompleto
		err := rows.Scan(
			&metrado.ID,
			&metrado.ProyectoID,
			&metrado.PartidaCodigo,
			&metrado.Metrado,
			&metrado.MetradoUnidad,
			&metrado.Observaciones,
			&metrado.PartidaDescripcion,
			&metrado.PartidaUnidad,
			&metrado.CostoUnitario,
			&metrado.CostoTotalPartida,
			&metrado.ProyectoNombre,
			&metrado.CreatedAt,
			&metrado.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando metrado: %v", err)
		}
		metrados = append(metrados, metrado)
	}

	return metrados, nil
}

// ObtenerMetradoPorPartida obtiene el metrado de una partida específica en un proyecto
func (r *MetradoRepository) ObtenerMetradoPorPartida(proyectoID uuid.UUID, partidaCodigo string) (*models.MetradoCompleto, error) {
	query := `SELECT * FROM vista_metrados_completos WHERE proyecto_id = $1 AND partida_codigo = $2`

	var metrado models.MetradoCompleto
	err := r.db.QueryRow(query, proyectoID, partidaCodigo).Scan(
		&metrado.ID,
		&metrado.ProyectoID,
		&metrado.PartidaCodigo,
		&metrado.Metrado,
		&metrado.MetradoUnidad,
		&metrado.Observaciones,
		&metrado.PartidaDescripcion,
		&metrado.PartidaUnidad,
		&metrado.CostoUnitario,
		&metrado.CostoTotalPartida,
		&metrado.ProyectoNombre,
		&metrado.CreatedAt,
		&metrado.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metrado no encontrado para la partida %s", partidaCodigo)
		}
		return nil, fmt.Errorf("error obteniendo metrado: %v", err)
	}

	return &metrado, nil
}

// EliminarMetrado elimina un metrado específico
func (r *MetradoRepository) EliminarMetrado(proyectoID uuid.UUID, partidaCodigo string) error {
	query := `DELETE FROM metrados_partidas WHERE proyecto_id = $1 AND partida_codigo = $2`

	result, err := r.db.Exec(query, proyectoID, partidaCodigo)
	if err != nil {
		return fmt.Errorf("error eliminando metrado: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("metrado no encontrado para eliminar")
	}

	return nil
}

// ObtenerResumenProyecto obtiene el resumen financiero de un proyecto
func (r *MetradoRepository) ObtenerResumenProyecto(proyectoID uuid.UUID) (*models.ResumenProyecto, error) {
	query := `SELECT * FROM obtener_resumen_proyecto($1)`

	var resumen models.ResumenProyecto
	err := r.db.QueryRow(query, proyectoID).Scan(
		&resumen.TotalPartidas,
		&resumen.CostoDirecto,
		&resumen.PartidasConMetrado,
		&resumen.PartidasSinMetrado,
	)

	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen del proyecto: %v", err)
	}

	return &resumen, nil
}

// CalcularCostoTotalProyecto calcula el costo total del proyecto con metrados
func (r *MetradoRepository) CalcularCostoTotalProyecto(proyectoID uuid.UUID) (float64, error) {
	query := `SELECT calcular_costo_total_proyecto($1)`

	var costoTotal float64
	err := r.db.QueryRow(query, proyectoID).Scan(&costoTotal)
	if err != nil {
		return 0, fmt.Errorf("error calculando costo total del proyecto: %v", err)
	}

	return costoTotal, nil
}

// ObtenerMetradosSimples obtiene metrados en formato simple (mapa clave-valor)
func (r *MetradoRepository) ObtenerMetradosSimples(proyectoID uuid.UUID) (map[string]float64, error) {
	query := `SELECT partida_codigo, metrado FROM metrados_partidas WHERE proyecto_id = $1`

	rows, err := r.db.Query(query, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando metrados simples: %v", err)
	}
	defer rows.Close()

	metrados := make(map[string]float64)
	for rows.Next() {
		var codigo string
		var metrado float64
		err := rows.Scan(&codigo, &metrado)
		if err != nil {
			return nil, fmt.Errorf("error escaneando metrado simple: %v", err)
		}
		metrados[codigo] = metrado
	}

	return metrados, nil
}