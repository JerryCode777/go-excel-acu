package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"goexcel/internal/database"
	"goexcel/internal/models"
)

type NormalizedMigrationService struct {
	db *database.DB
}

// Interface común para DB y Tx
type DBExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func NewNormalizedMigrationService(db *database.DB) *NormalizedMigrationService {
	return &NormalizedMigrationService{db: db}
}

func (s *NormalizedMigrationService) MigrateNormalizedData(data *models.NormalizedData) error {
	log.Printf("🚀 Iniciando migración de datos normalizados")
	
	// Iniciar transacción explícita
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	
	// Variable para capturar errores durante la operación
	var migrationErr error
	
	// Asegurar que la transacción se complete
	defer func() {
		if migrationErr != nil {
			log.Printf("❌ Error durante migración, haciendo rollback: %v", migrationErr)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("❌ Error adicional haciendo rollback: %v", rollbackErr)
			}
		} else {
			log.Printf("✅ Migración exitosa, haciendo commit")
			if commitErr := tx.Commit(); commitErr != nil {
				log.Printf("❌ Error haciendo commit: %v", commitErr)
				migrationErr = fmt.Errorf("error haciendo commit: %w", commitErr)
			} else {
				log.Printf("🎯 Commit completado exitosamente")
			}
		}
	}()

	// 1. Crear proyecto
	proyectoUUID, migrationErr := uuid.Parse(data.Proyecto.ID)
	if migrationErr != nil {
		migrationErr = fmt.Errorf("error parseando UUID del proyecto: %w", migrationErr)
		return migrationErr
	}

	migrationErr = s.insertProyectoTx(tx, proyectoUUID, data.Proyecto)
	if migrationErr != nil {
		migrationErr = fmt.Errorf("error insertando proyecto: %w", migrationErr)
		return migrationErr
	}
	log.Printf("✅ Proyecto insertado: %s", data.Proyecto.Nombre)
	
	// Verificar que el proyecto se insertó correctamente en la transacción
	var count int
	verifyErr := tx.QueryRow("SELECT COUNT(*) FROM proyectos WHERE id = $1", proyectoUUID).Scan(&count)
	if verifyErr == nil {
		log.Printf("🔍 Proyecto verificado en TX: %d registros encontrados", count)
	} else {
		log.Printf("⚠️  Error verificando proyecto en TX: %v", verifyErr)
	}

	// 2. Obtener mapeo de tipos de recurso
	tiposRecurso, migrationErr := s.getTiposRecursoTx(tx)
	if migrationErr != nil {
		migrationErr = fmt.Errorf("error obteniendo tipos de recurso: %w", migrationErr)
		return migrationErr
	}

	// 3. Insertar recursos y crear mapeo de UUIDs reales
	recursosInsertados := 0
	recursosRealMap := make(map[string]uuid.UUID) // mapeo código -> UUID real en BD
	
	for _, recurso := range data.Recursos {
		tipoID, exists := tiposRecurso[recurso.TipoRecurso]
		if !exists {
			log.Printf("⚠️  Tipo de recurso desconocido: %s", recurso.TipoRecurso)
			continue
		}

		recursoUUID, err := uuid.Parse(recurso.ID)
		if err != nil {
			log.Printf("⚠️  UUID inválido para recurso %s: %v", recurso.Codigo, err)
			continue
		}

		// Insertar o actualizar recurso
		realUUID, err := s.insertRecursoAndGetIDTx(tx, recursoUUID, recurso, tipoID)
		if err != nil {
			log.Printf("⚠️  Error insertando recurso %s: %v", recurso.Codigo, err)
			continue
		}
		
		// Guardar mapeo del código al UUID real en la BD
		recursosRealMap[recurso.Codigo] = realUUID
		recursosInsertados++
	}
	log.Printf("✅ Recursos insertados: %d/%d", recursosInsertados, len(data.Recursos))

	// 4. Insertar partidas y crear mapeo de UUIDs reales
	partidasInsertadas := 0
	partidasRealMap := make(map[string]uuid.UUID) // mapeo código -> UUID real en BD
	
	for _, partida := range data.Partidas {
		partidaUUID, err := uuid.Parse(partida.ID)
		if err != nil {
			log.Printf("⚠️  UUID inválido para partida %s: %v", partida.Codigo, err)
			continue
		}

		// Insertar o actualizar partida
		realUUID, err := s.insertPartidaAndGetIDTx(tx, partidaUUID, partida, proyectoUUID)
		if err != nil {
			log.Printf("⚠️  Error insertando partida %s: %v", partida.Codigo, err)
			continue
		}
		
		// Guardar mapeo del código al UUID real en la BD
		partidasRealMap[partida.Codigo] = realUUID
		partidasInsertadas++
	}
	log.Printf("✅ Partidas insertadas: %d/%d", partidasInsertadas, len(data.Partidas))

	// 5. Insertar relaciones partida-recurso usando UUIDs reales
	relacionesInsertadas := 0
	
	// Crear mapeo inverso para buscar códigos por UUID normalizado
	recursoNormToCode := make(map[string]string)
	partidaNormToCode := make(map[string]string)
	
	for _, recurso := range data.Recursos {
		recursoNormToCode[recurso.ID] = recurso.Codigo
	}
	
	for _, partida := range data.Partidas {
		partidaNormToCode[partida.ID] = partida.Codigo
	}
	
	for _, relacion := range data.Relaciones {
		relacionUUID, err := uuid.Parse(relacion.ID)
		if err != nil {
			log.Printf("⚠️  UUID inválido para relación: %v", err)
			continue
		}

		// Encontrar códigos usando los UUIDs normalizados
		partidaCodigo, partidaExists := partidaNormToCode[relacion.PartidaID]
		recursoCodigo, recursoExists := recursoNormToCode[relacion.RecursoID]
		
		if !partidaExists || !recursoExists {
			log.Printf("⚠️  No se encontraron códigos para la relación")
			continue
		}
		
		// Obtener UUIDs reales usando los códigos
		partidaRealUUID, partidaRealExists := partidasRealMap[partidaCodigo]
		recursoRealUUID, recursoRealExists := recursosRealMap[recursoCodigo]
		
		if !partidaRealExists || !recursoRealExists {
			log.Printf("⚠️  No se encontraron UUIDs reales para partida %s o recurso %s", partidaCodigo, recursoCodigo)
			continue
		}

		err = s.insertRelacionTx(tx, relacionUUID, relacion, partidaRealUUID, recursoRealUUID)
		if err != nil {
			log.Printf("⚠️  Error insertando relación: %v", err)
			continue
		}
		relacionesInsertadas++
	}
	log.Printf("✅ Relaciones insertadas: %d/%d", relacionesInsertadas, len(data.Relaciones))

	log.Printf("🎉 Migración completada exitosamente")
	migrationErr = nil // Asegurar que no hay error al final
	return migrationErr
}

func (s *NormalizedMigrationService) insertProyecto(id uuid.UUID, proyecto models.ProyectoNormalizado) error {
	query := `
		INSERT INTO proyectos (id, nombre, descripcion, moneda)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			nombre = EXCLUDED.nombre,
			descripcion = EXCLUDED.descripcion,
			moneda = EXCLUDED.moneda,
			updated_at = CURRENT_TIMESTAMP
	`

	log.Printf("🔍 Ejecutando inserción de proyecto - ID: %s, Nombre: %s", id.String(), proyecto.Nombre)
	result, err := s.db.Exec(query, id, proyecto.Nombre, proyecto.Descripcion, proyecto.Moneda)
	if err != nil {
		log.Printf("❌ Error ejecutando inserción de proyecto: %v", err)
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Proyecto insertado/actualizado - Filas afectadas: %d", rowsAffected)
	return nil
}

func (s *NormalizedMigrationService) getTiposRecurso() (map[string]uuid.UUID, error) {
	query := `SELECT nombre, id FROM tipos_recurso`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tipos := make(map[string]uuid.UUID)
	for rows.Next() {
		var nombre string
		var id uuid.UUID
		if err := rows.Scan(&nombre, &id); err != nil {
			return nil, err
		}
		tipos[nombre] = id
	}

	return tipos, nil
}

func (s *NormalizedMigrationService) insertRecurso(id uuid.UUID, recurso models.RecursoNormalizado, tipoRecursoID uuid.UUID) error {
	query := `
		INSERT INTO recursos (id, codigo, descripcion, unidad, precio_base, tipo_recurso_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			precio_base = EXCLUDED.precio_base,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, id, recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.PrecioBase, tipoRecursoID)
	return err
}

func (s *NormalizedMigrationService) insertRecursoAndGetID(id uuid.UUID, recurso models.RecursoNormalizado, tipoRecursoID uuid.UUID) (uuid.UUID, error) {
	// Primero intentar insertar/actualizar
	query := `
		INSERT INTO recursos (id, codigo, descripcion, unidad, precio_base, tipo_recurso_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			precio_base = EXCLUDED.precio_base,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, id, recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.PrecioBase, tipoRecursoID)
	if err != nil {
		return uuid.Nil, err
	}

	// Luego obtener el UUID real que está en la base de datos
	var realID uuid.UUID
	selectQuery := `SELECT id FROM recursos WHERE codigo = $1`
	err = s.db.QueryRow(selectQuery, recurso.Codigo).Scan(&realID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error obteniendo UUID real del recurso: %w", err)
	}

	return realID, nil
}

func (s *NormalizedMigrationService) insertPartida(id uuid.UUID, partida models.PartidaNormalizada, proyectoID uuid.UUID) error {
	query := `
		INSERT INTO partidas (id, proyecto_id, codigo, descripcion, unidad, rendimiento)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (proyecto_id, codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			rendimiento = EXCLUDED.rendimiento,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, id, proyectoID, partida.Codigo, partida.Descripcion, partida.Unidad, partida.Rendimiento)
	return err
}

func (s *NormalizedMigrationService) insertPartidaAndGetID(id uuid.UUID, partida models.PartidaNormalizada, proyectoID uuid.UUID) (uuid.UUID, error) {
	// Primero intentar insertar/actualizar
	query := `
		INSERT INTO partidas (id, proyecto_id, codigo, descripcion, unidad, rendimiento)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (proyecto_id, codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			rendimiento = EXCLUDED.rendimiento,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, id, proyectoID, partida.Codigo, partida.Descripcion, partida.Unidad, partida.Rendimiento)
	if err != nil {
		return uuid.Nil, err
	}

	// Luego obtener el UUID real que está en la base de datos
	var realID uuid.UUID
	selectQuery := `SELECT id FROM partidas WHERE proyecto_id = $1 AND codigo = $2`
	err = s.db.QueryRow(selectQuery, proyectoID, partida.Codigo).Scan(&realID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error obteniendo UUID real de la partida: %w", err)
	}

	return realID, nil
}

func (s *NormalizedMigrationService) insertRelacion(id uuid.UUID, relacion models.RelacionNormalizada, partidaID, recursoID uuid.UUID) error {
	query := `
		INSERT INTO partida_recursos (id, partida_id, recurso_id, cantidad, precio, cuadrilla)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
			cantidad = EXCLUDED.cantidad,
			precio = EXCLUDED.precio,
			cuadrilla = EXCLUDED.cuadrilla,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, id, partidaID, recursoID, relacion.Cantidad, relacion.Precio, relacion.Cuadrilla)
	return err
}

// Versiones transaccionales de los métodos

func (s *NormalizedMigrationService) insertProyectoTx(tx *sql.Tx, id uuid.UUID, proyecto models.ProyectoNormalizado) error {
	query := `
		INSERT INTO proyectos (id, nombre, descripcion, moneda)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			nombre = EXCLUDED.nombre,
			descripcion = EXCLUDED.descripcion,
			moneda = EXCLUDED.moneda,
			updated_at = CURRENT_TIMESTAMP
	`

	log.Printf("🔍 Ejecutando inserción de proyecto - ID: %s, Nombre: %s", id.String(), proyecto.Nombre)
	result, err := tx.Exec(query, id, proyecto.Nombre, proyecto.Descripcion, proyecto.Moneda)
	if err != nil {
		log.Printf("❌ Error ejecutando inserción de proyecto: %v", err)
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Proyecto insertado/actualizado - Filas afectadas: %d", rowsAffected)
	return nil
}

func (s *NormalizedMigrationService) getTiposRecursoTx(tx *sql.Tx) (map[string]uuid.UUID, error) {
	query := `SELECT nombre, id FROM tipos_recurso`
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tipos := make(map[string]uuid.UUID)
	for rows.Next() {
		var nombre string
		var id uuid.UUID
		if err := rows.Scan(&nombre, &id); err != nil {
			return nil, err
		}
		tipos[nombre] = id
	}

	return tipos, nil
}

func (s *NormalizedMigrationService) insertRecursoAndGetIDTx(tx *sql.Tx, id uuid.UUID, recurso models.RecursoNormalizado, tipoRecursoID uuid.UUID) (uuid.UUID, error) {
	// Primero intentar insertar/actualizar
	query := `
		INSERT INTO recursos (id, codigo, descripcion, unidad, precio_base, tipo_recurso_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			precio_base = EXCLUDED.precio_base,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := tx.Exec(query, id, recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.PrecioBase, tipoRecursoID)
	if err != nil {
		return uuid.Nil, err
	}

	// Luego obtener el UUID real que está en la base de datos
	var realID uuid.UUID
	selectQuery := `SELECT id FROM recursos WHERE codigo = $1`
	err = tx.QueryRow(selectQuery, recurso.Codigo).Scan(&realID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error obteniendo UUID real del recurso: %w", err)
	}

	return realID, nil
}

func (s *NormalizedMigrationService) insertPartidaAndGetIDTx(tx *sql.Tx, id uuid.UUID, partida models.PartidaNormalizada, proyectoID uuid.UUID) (uuid.UUID, error) {
	// Primero intentar insertar/actualizar
	query := `
		INSERT INTO partidas (id, proyecto_id, codigo, descripcion, unidad, rendimiento)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (proyecto_id, codigo) DO UPDATE SET
			descripcion = EXCLUDED.descripcion,
			unidad = EXCLUDED.unidad,
			rendimiento = EXCLUDED.rendimiento,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := tx.Exec(query, id, proyectoID, partida.Codigo, partida.Descripcion, partida.Unidad, partida.Rendimiento)
	if err != nil {
		return uuid.Nil, err
	}

	// Luego obtener el UUID real que está en la base de datos
	var realID uuid.UUID
	selectQuery := `SELECT id FROM partidas WHERE proyecto_id = $1 AND codigo = $2`
	err = tx.QueryRow(selectQuery, proyectoID, partida.Codigo).Scan(&realID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error obteniendo UUID real de la partida: %w", err)
	}

	return realID, nil
}

func (s *NormalizedMigrationService) insertRelacionTx(tx *sql.Tx, id uuid.UUID, relacion models.RelacionNormalizada, partidaID, recursoID uuid.UUID) error {
	query := `
		INSERT INTO partida_recursos (id, partida_id, recurso_id, cantidad, precio, cuadrilla)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
			cantidad = EXCLUDED.cantidad,
			precio = EXCLUDED.precio,
			cuadrilla = EXCLUDED.cuadrilla,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := tx.Exec(query, id, partidaID, recursoID, relacion.Cantidad, relacion.Precio, relacion.Cuadrilla)
	return err
}

func (s *NormalizedMigrationService) MigrateNormalizedDataWithUser(data *models.NormalizedData, usuarioID uuid.UUID) error {
	log.Printf("🚀 Iniciando migración de datos normalizados con usuario: %s", usuarioID.String())
	
	// Iniciar transacción explícita
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	
	// Variable para capturar errores durante la operación
	var migrationErr error
	
	// Asegurar que la transacción se complete
	defer func() {
		if migrationErr != nil {
			log.Printf("❌ Error durante migración, haciendo rollback: %v", migrationErr)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("❌ Error adicional haciendo rollback: %v", rollbackErr)
			}
		} else {
			log.Printf("✅ Migración exitosa, haciendo commit")
			if commitErr := tx.Commit(); commitErr != nil {
				log.Printf("❌ Error haciendo commit: %v", commitErr)
				migrationErr = fmt.Errorf("error haciendo commit: %w", commitErr)
			} else {
				log.Printf("🎯 Commit completado exitosamente")
			}
		}
	}()

	// 1. Crear proyecto con usuario_id
	proyectoUUID, migrationErr := uuid.Parse(data.Proyecto.ID)
	if migrationErr != nil {
		migrationErr = fmt.Errorf("error parseando UUID del proyecto: %w", migrationErr)
		return migrationErr
	}

	migrationErr = s.insertProyectoWithUserTx(tx, proyectoUUID, data.Proyecto, usuarioID)
	if migrationErr != nil {
		migrationErr = fmt.Errorf("error insertando proyecto: %w", migrationErr)
		return migrationErr
	}
	log.Printf("✅ Proyecto insertado: %s", data.Proyecto.Nombre)
	
	// Verificar que el proyecto se insertó correctamente en la transacción
	var count int
	verifyErr := tx.QueryRow("SELECT COUNT(*) FROM proyectos WHERE id = $1", proyectoUUID).Scan(&count)
	if verifyErr == nil {
		log.Printf("🔍 Proyecto verificado en TX: %d registros encontrados", count)
	} else {
		log.Printf("⚠️  Error verificando proyecto en TX: %v", verifyErr)
	}

	// 2. Obtener mapeo de tipos de recurso
	tiposRecurso, migrationErr := s.getTiposRecursoTx(tx)
	if migrationErr != nil {
		migrationErr = fmt.Errorf("error obteniendo tipos de recurso: %w", migrationErr)
		return migrationErr
	}

	// 3. Insertar recursos y crear mapeo de UUIDs reales
	recursosInsertados := 0
	recursosRealMap := make(map[string]uuid.UUID) // mapeo código -> UUID real en BD
	
	for _, recurso := range data.Recursos {
		tipoID, exists := tiposRecurso[recurso.TipoRecurso]
		if !exists {
			log.Printf("⚠️  Tipo de recurso desconocido: %s", recurso.TipoRecurso)
			continue
		}

		recursoUUID, err := uuid.Parse(recurso.ID)
		if err != nil {
			log.Printf("⚠️  Error parseando UUID del recurso %s: %v", recurso.Codigo, err)
			continue
		}

		// Insertar recurso y obtener el UUID real en la BD
		realRecursoID, err := s.insertRecursoAndGetIDTx(tx, recursoUUID, recurso, tipoID)
		if err != nil {
			log.Printf("⚠️  Error insertando recurso %s: %v", recurso.Codigo, err)
			continue
		}

		recursosRealMap[recurso.Codigo] = realRecursoID
		recursosInsertados++
	}
	log.Printf("✅ Recursos insertados: %d/%d", recursosInsertados, len(data.Recursos))

	// 4. Insertar partidas
	partidasInsertadas := 0
	partidasRealMap := make(map[string]uuid.UUID) // mapeo código partida -> UUID real en BD
	
	for _, partida := range data.Partidas {
		partidaUUID, err := uuid.Parse(partida.ID)
		if err != nil {
			log.Printf("⚠️  Error parseando UUID de la partida %s: %v", partida.Codigo, err)
			continue
		}

		// Insertar partida y obtener el UUID real en la BD
		realPartidaID, err := s.insertPartidaAndGetIDTx(tx, partidaUUID, partida, proyectoUUID)
		if err != nil {
			log.Printf("⚠️  Error insertando partida %s: %v", partida.Codigo, err)
			continue
		}

		partidasRealMap[partida.Codigo] = realPartidaID
		partidasInsertadas++
	}
	log.Printf("✅ Partidas insertadas: %d/%d", partidasInsertadas, len(data.Partidas))

	// 5. Insertar relaciones partida-recurso usando UUIDs reales
	relacionesInsertadas := 0
	
	// Crear mapeo inverso para buscar códigos por UUID normalizado
	recursoNormToCode := make(map[string]string)
	partidaNormToCode := make(map[string]string)
	
	for _, recurso := range data.Recursos {
		recursoNormToCode[recurso.ID] = recurso.Codigo
	}
	
	for _, partida := range data.Partidas {
		partidaNormToCode[partida.ID] = partida.Codigo
	}
	
	for _, relacion := range data.Relaciones {
		relacionUUID, err := uuid.Parse(relacion.ID)
		if err != nil {
			log.Printf("⚠️  UUID inválido para relación: %v", err)
			continue
		}

		// Encontrar códigos usando los UUIDs normalizados
		partidaCodigo, partidaExists := partidaNormToCode[relacion.PartidaID]
		recursoCodigo, recursoExists := recursoNormToCode[relacion.RecursoID]
		
		if !partidaExists || !recursoExists {
			log.Printf("⚠️  No se encontraron códigos para la relación")
			continue
		}
		
		// Obtener UUIDs reales usando los códigos
		partidaRealUUID, partidaRealExists := partidasRealMap[partidaCodigo]
		recursoRealUUID, recursoRealExists := recursosRealMap[recursoCodigo]
		
		if !partidaRealExists || !recursoRealExists {
			log.Printf("⚠️  No se encontraron UUIDs reales para partida %s o recurso %s", partidaCodigo, recursoCodigo)
			continue
		}

		err = s.insertRelacionTx(tx, relacionUUID, relacion, partidaRealUUID, recursoRealUUID)
		if err != nil {
			log.Printf("⚠️  Error insertando relación: %v", err)
			continue
		}
		relacionesInsertadas++
	}
	log.Printf("✅ Relaciones insertadas: %d/%d", relacionesInsertadas, len(data.Relaciones))

	log.Printf("🎉 Migración completada exitosamente")
	return nil
}

func (s *NormalizedMigrationService) insertProyectoWithUserTx(tx *sql.Tx, id uuid.UUID, proyecto models.ProyectoNormalizado, usuarioID uuid.UUID) error {
	query := `
		INSERT INTO proyectos (id, nombre, descripcion, moneda, usuario_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			nombre = EXCLUDED.nombre,
			descripcion = EXCLUDED.descripcion,
			moneda = EXCLUDED.moneda,
			usuario_id = EXCLUDED.usuario_id,
			updated_at = CURRENT_TIMESTAMP
	`

	log.Printf("🔍 Ejecutando inserción de proyecto con usuario - ID: %s, Nombre: %s, Usuario: %s", id.String(), proyecto.Nombre, usuarioID.String())
	result, err := tx.Exec(query, id, proyecto.Nombre, proyecto.Descripcion, proyecto.Moneda, usuarioID)
	if err != nil {
		log.Printf("❌ Error ejecutando inserción de proyecto: %v", err)
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Proyecto insertado/actualizado con usuario - Filas afectadas: %d", rowsAffected)
	return nil
}