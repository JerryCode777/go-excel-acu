package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"goexcel/config"
)

type DB struct {
	*sql.DB
}

func New(cfg *config.Config) (*DB, error) {
	db, err := sql.Open("postgres", cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión a base de datos: %w", err)
	}

	// Configurar pool de conexiones
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a base de datos: %w", err)
	}

	log.Printf("Conectado exitosamente a PostgreSQL: %s:%d/%s", 
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

// Ejecutar migraciones
func (db *DB) RunMigrations() error {
	migrationQuery := `
	-- Crear extensiones necesarias
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	
	-- Crear tablas si no existen
	CREATE TABLE IF NOT EXISTS proyectos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		nombre VARCHAR(255) NOT NULL,
		descripcion TEXT,
		ubicacion VARCHAR(255),
		cliente VARCHAR(255),
		fecha_inicio DATE,
		fecha_fin DATE,
		moneda VARCHAR(10) DEFAULT 'PEN',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS tipos_recurso (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		nombre VARCHAR(50) NOT NULL UNIQUE,
		descripcion TEXT,
		orden INTEGER DEFAULT 0
	);
	
	-- Insertar tipos de recurso por defecto si no existen
	INSERT INTO tipos_recurso (nombre, descripcion, orden) VALUES
		('mano_obra', 'Mano de Obra', 1),
		('materiales', 'Materiales', 2),
		('equipos', 'Equipos', 3),
		('subcontratos', 'Subcontratos', 4)
	ON CONFLICT (nombre) DO NOTHING;
	
	CREATE TABLE IF NOT EXISTS recursos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		codigo VARCHAR(50) NOT NULL UNIQUE,
		descripcion TEXT NOT NULL,
		unidad VARCHAR(20) NOT NULL,
		precio_base DECIMAL(15,4) NOT NULL DEFAULT 0,
		tipo_recurso_id UUID REFERENCES tipos_recurso(id),
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS partidas (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		proyecto_id UUID REFERENCES proyectos(id) ON DELETE CASCADE,
		codigo VARCHAR(50) NOT NULL,
		descripcion TEXT NOT NULL,
		unidad VARCHAR(20) NOT NULL,
		rendimiento DECIMAL(15,6) NOT NULL DEFAULT 1.0,
		costo_total DECIMAL(15,4) DEFAULT 0,
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(proyecto_id, codigo)
	);
	
	CREATE TABLE IF NOT EXISTS partida_recursos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		partida_id UUID REFERENCES partidas(id) ON DELETE CASCADE,
		recurso_id UUID REFERENCES recursos(id) ON DELETE CASCADE,
		cantidad DECIMAL(15,6) NOT NULL DEFAULT 0,
		precio DECIMAL(15,4) NOT NULL DEFAULT 0,
		cuadrilla DECIMAL(15,6) DEFAULT NULL,
		parcial DECIMAL(15,4) GENERATED ALWAYS AS (cantidad * precio) STORED,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(partida_id, recurso_id)
	);
	
	CREATE TABLE IF NOT EXISTS analisis_historicos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		proyecto_id UUID REFERENCES proyectos(id) ON DELETE CASCADE,
		nombre_archivo VARCHAR(255),
		fecha_analisis TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		total_partidas INTEGER DEFAULT 0,
		costo_total_mano_obra DECIMAL(15,4) DEFAULT 0,
		costo_total_materiales DECIMAL(15,4) DEFAULT 0,
		costo_total_equipos DECIMAL(15,4) DEFAULT 0,
		costo_total_subcontratos DECIMAL(15,4) DEFAULT 0,
		costo_total_proyecto DECIMAL(15,4) DEFAULT 0,
		archivo_excel_url TEXT
	);
	
	-- Crear índices si no existen
	CREATE INDEX IF NOT EXISTS idx_partidas_proyecto_id ON partidas(proyecto_id);
	CREATE INDEX IF NOT EXISTS idx_partidas_codigo ON partidas(codigo);
	CREATE INDEX IF NOT EXISTS idx_recursos_codigo ON recursos(codigo);
	CREATE INDEX IF NOT EXISTS idx_recursos_tipo ON recursos(tipo_recurso_id);
	CREATE INDEX IF NOT EXISTS idx_partida_recursos_partida_id ON partida_recursos(partida_id);
	CREATE INDEX IF NOT EXISTS idx_partida_recursos_recurso_id ON partida_recursos(recurso_id);
	`

	// Ejecutar migraciones base
	_, err := db.Exec(migrationQuery)
	if err != nil {
		return fmt.Errorf("error ejecutando migraciones base: %w", err)
	}

	// Ejecutar migraciones jerárquicas
	if err := db.runJerarquicoMigrations(); err != nil {
		return fmt.Errorf("error ejecutando migraciones jerárquicas: %w", err)
	}

	log.Println("Migraciones ejecutadas exitosamente")
	return nil
}

// runJerarquicoMigrations ejecuta las migraciones para el sistema jerárquico
func (db *DB) runJerarquicoMigrations() error {
	jerarquicoMigration := `
	-- Tablas para sistema jerárquico de presupuestos
	CREATE TABLE IF NOT EXISTS presupuestos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		codigo VARCHAR(50) NOT NULL,
		nombre TEXT NOT NULL,
		cliente TEXT,
		lugar TEXT,
		moneda VARCHAR(10) DEFAULT 'PEN',
		fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		usuario_id UUID, -- Para futuro soporte multiusuario
		organizacion_id UUID, -- Para futuro soporte multi-tenant
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS subpresupuestos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		presupuesto_id UUID NOT NULL REFERENCES presupuestos(id) ON DELETE CASCADE,
		codigo VARCHAR(50) NOT NULL,
		nombre TEXT NOT NULL,
		orden INTEGER DEFAULT 0,
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(presupuesto_id, codigo)
	);

	CREATE TABLE IF NOT EXISTS titulos (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		presupuesto_id UUID NOT NULL REFERENCES presupuestos(id) ON DELETE CASCADE,
		subpresupuesto_id UUID REFERENCES subpresupuestos(id) ON DELETE CASCADE,
		titulo_padre_id UUID REFERENCES titulos(id) ON DELETE CASCADE,
		nivel INTEGER NOT NULL CHECK (nivel >= 1 AND nivel <= 10),
		numero INTEGER NOT NULL DEFAULT 1,
		codigo_completo VARCHAR(50) NOT NULL,
		nombre TEXT NOT NULL,
		orden INTEGER DEFAULT 0,
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(presupuesto_id, codigo_completo)
	);

	-- Actualizar tabla partidas para soporte jerárquico
	ALTER TABLE partidas ADD COLUMN IF NOT EXISTS presupuesto_id UUID REFERENCES presupuestos(id) ON DELETE CASCADE;
	ALTER TABLE partidas ADD COLUMN IF NOT EXISTS subpresupuesto_id UUID REFERENCES subpresupuestos(id) ON DELETE CASCADE;
	ALTER TABLE partidas ADD COLUMN IF NOT EXISTS titulo_id UUID REFERENCES titulos(id) ON DELETE CASCADE;
	ALTER TABLE partidas ADD COLUMN IF NOT EXISTS numero INTEGER DEFAULT 1;
	ALTER TABLE partidas ADD COLUMN IF NOT EXISTS orden INTEGER DEFAULT 0;

	-- Función para generar códigos jerárquicos automáticamente
	CREATE OR REPLACE FUNCTION generar_codigo_jerarquico(
		presupuesto_uuid UUID,
		subpresupuesto_uuid UUID DEFAULT NULL,
		titulo_padre_uuid UUID DEFAULT NULL,
		nivel_param INTEGER DEFAULT 1
	) RETURNS VARCHAR(50) AS $$
	DECLARE
		codigo_resultado VARCHAR(50);
		siguiente_numero INTEGER;
		codigo_padre VARCHAR(50);
	BEGIN
		-- Si hay título padre, obtener su código y generar subcódigo
		IF titulo_padre_uuid IS NOT NULL THEN
			SELECT codigo_completo INTO codigo_padre 
			FROM titulos WHERE id = titulo_padre_uuid;
			
			SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
			FROM titulos 
			WHERE titulo_padre_id = titulo_padre_uuid AND nivel = nivel_param;
			
			codigo_resultado := codigo_padre || '.' || LPAD(siguiente_numero::TEXT, 2, '0');
		
		-- Si hay subpresupuesto pero no título padre
		ELSIF subpresupuesto_uuid IS NOT NULL THEN
			SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
			FROM titulos 
			WHERE presupuesto_id = presupuesto_uuid 
			  AND subpresupuesto_id = subpresupuesto_uuid 
			  AND titulo_padre_id IS NULL;
			
			codigo_resultado := LPAD(siguiente_numero::TEXT, 2, '0');
		
		-- Título de primer nivel
		ELSE
			SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
			FROM titulos 
			WHERE presupuesto_id = presupuesto_uuid 
			  AND titulo_padre_id IS NULL 
			  AND subpresupuesto_id IS NULL;
			
			codigo_resultado := LPAD(siguiente_numero::TEXT, 2, '0');
		END IF;
		
		RETURN codigo_resultado;
	END;
	$$ LANGUAGE plpgsql;

	-- Vista para estructura jerárquica completa
	CREATE OR REPLACE VIEW vista_estructura_jerarquica AS
	WITH RECURSIVE jerarquia AS (
		-- Nodo raíz: presupuestos
		SELECT 
			p.id as presupuesto_id,
			p.codigo as presupuesto_codigo,
			p.nombre as presupuesto_nombre,
			NULL::UUID as subpresupuesto_id,
			NULL::VARCHAR as subpresupuesto_codigo,
			NULL::TEXT as subpresupuesto_nombre,
			NULL::UUID as titulo_id,
			NULL::INTEGER as nivel,
			NULL::VARCHAR as titulo_codigo,
			NULL::TEXT as titulo_nombre,
			0 as depth,
			'0' as path_orden,
			0::BIGINT as total_partidas,
			0::DECIMAL as costo_total_titulos
		FROM presupuestos p
		WHERE p.activo = true
		
		UNION ALL
		
		-- Títulos jerárquicos
		SELECT 
			j.presupuesto_id,
			j.presupuesto_codigo,
			j.presupuesto_nombre,
			t.subpresupuesto_id,
			s.codigo as subpresupuesto_codigo,
			s.nombre as subpresupuesto_nombre,
			t.id as titulo_id,
			t.nivel,
			t.codigo_completo as titulo_codigo,
			t.nombre as titulo_nombre,
			j.depth + 1,
			j.path_orden || '.' || LPAD(t.orden::TEXT, 3, '0'),
			(SELECT COUNT(*) FROM partidas pt WHERE pt.titulo_id = t.id AND pt.activo = true) as total_partidas,
			(SELECT COALESCE(SUM(pt.costo_total), 0) FROM partidas pt WHERE pt.titulo_id = t.id AND pt.activo = true) as costo_total_titulos
		FROM jerarquia j
		JOIN titulos t ON t.presupuesto_id = j.presupuesto_id
		LEFT JOIN subpresupuestos s ON s.id = t.subpresupuesto_id
		WHERE t.activo = true
		  AND (t.titulo_padre_id IS NULL OR t.titulo_padre_id = j.titulo_id)
	)
	SELECT * FROM jerarquia WHERE titulo_id IS NOT NULL;

	-- Vista para partidas con información jerárquica
	CREATE OR REPLACE VIEW vista_partidas_jerarquicas AS
	SELECT 
		pt.id as partida_id,
		pt.codigo as partida_codigo,
		pt.descripcion as partida_descripcion,
		pt.unidad,
		pt.rendimiento,
		pt.costo_total,
		pt.numero as partida_numero,
		pt.orden as partida_orden,
		-- Información del presupuesto
		p.id as presupuesto_id,
		p.codigo as presupuesto_codigo,
		p.nombre as presupuesto_nombre,
		-- Información del subpresupuesto
		sp.id as subpresupuesto_id,
		sp.codigo as subpresupuesto_codigo,
		sp.nombre as subpresupuesto_nombre,
		-- Información del título
		t.id as titulo_id,
		t.codigo_completo as titulo_codigo,
		t.nombre as titulo_nombre,
		t.nivel as titulo_nivel
	FROM partidas pt
	JOIN presupuestos p ON p.id = pt.presupuesto_id
	LEFT JOIN subpresupuestos sp ON sp.id = pt.subpresupuesto_id
	LEFT JOIN titulos t ON t.id = pt.titulo_id
	WHERE pt.activo = true AND p.activo = true;

	-- Función para obtener resumen jerárquico
	CREATE OR REPLACE FUNCTION obtener_resumen_jerarquico(presupuesto_uuid UUID)
	RETURNS TABLE(
		total_subpresupuestos BIGINT,
		total_titulos BIGINT,
		total_partidas BIGINT,
		costo_total DECIMAL,
		niveles_maximos INTEGER
	) AS $$
	BEGIN
		RETURN QUERY
		SELECT 
			(SELECT COUNT(*) FROM subpresupuestos WHERE presupuesto_id = presupuesto_uuid AND activo = true),
			(SELECT COUNT(*) FROM titulos WHERE presupuesto_id = presupuesto_uuid AND activo = true),
			(SELECT COUNT(*) FROM partidas WHERE presupuesto_id = presupuesto_uuid AND activo = true),
			(SELECT COALESCE(SUM(costo_total), 0) FROM partidas WHERE presupuesto_id = presupuesto_uuid AND activo = true),
			(SELECT COALESCE(MAX(nivel), 0) FROM titulos WHERE presupuesto_id = presupuesto_uuid AND activo = true);
	END;
	$$ LANGUAGE plpgsql;

	-- Índices para el sistema jerárquico
	CREATE INDEX IF NOT EXISTS idx_presupuestos_codigo ON presupuestos(codigo);
	CREATE INDEX IF NOT EXISTS idx_presupuestos_activo ON presupuestos(activo);
	CREATE INDEX IF NOT EXISTS idx_subpresupuestos_presupuesto ON subpresupuestos(presupuesto_id);
	CREATE INDEX IF NOT EXISTS idx_subpresupuestos_codigo ON subpresupuestos(codigo);
	CREATE INDEX IF NOT EXISTS idx_titulos_presupuesto ON titulos(presupuesto_id);
	CREATE INDEX IF NOT EXISTS idx_titulos_padre ON titulos(titulo_padre_id);
	CREATE INDEX IF NOT EXISTS idx_titulos_nivel ON titulos(nivel);
	CREATE INDEX IF NOT EXISTS idx_titulos_codigo ON titulos(codigo_completo);
	CREATE INDEX IF NOT EXISTS idx_partidas_titulo ON partidas(titulo_id);
	CREATE INDEX IF NOT EXISTS idx_partidas_presupuesto_jerarquico ON partidas(presupuesto_id);
	`

	_, err := db.Exec(jerarquicoMigration)
	return err
}