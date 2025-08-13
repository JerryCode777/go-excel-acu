package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/jerryandersonh/goexcel/config"
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

	_, err := db.Exec(migrationQuery)
	if err != nil {
		return fmt.Errorf("error ejecutando migraciones: %w", err)
	}

	log.Println("Migraciones ejecutadas exitosamente")
	return nil
}