# Database Schema

## 🗄️ Esquema de Base de Datos PostgreSQL

Esta documentación describe el esquema completo de la base de datos de GoExcel.

## 📊 Diagrama de relaciones

```
proyectos (1) ←→ (N) partidas
    ↑                  ↓
    └─────────────── (N) partida_recursos (N) ←→ (1) recursos
                           ↓
                        tipos_recurso (1)
```

## 📋 Tablas principales

### 1. proyectos
Almacena información general de los proyectos.

```sql
CREATE TABLE proyectos (
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
```

**Campos:**
- `id`: Identificador único del proyecto
- `nombre`: Nombre del proyecto (requerido)
- `descripcion`: Descripción detallada
- `ubicacion`: Ubicación geográfica
- `cliente`: Cliente del proyecto
- `fecha_inicio`: Fecha de inicio planificada
- `fecha_fin`: Fecha de fin planificada
- `moneda`: Código de moneda (PEN, USD, etc.)
- `created_at`: Fecha de creación
- `updated_at`: Fecha de última actualización

### 2. tipos_recurso
Catálogo de tipos de recursos disponibles.

```sql
CREATE TABLE tipos_recurso (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(50) NOT NULL UNIQUE,
    descripcion TEXT,
    orden INTEGER DEFAULT 0
);

-- Datos iniciales
INSERT INTO tipos_recurso (nombre, descripcion, orden) VALUES
    ('mano_obra', 'Mano de Obra', 1),
    ('materiales', 'Materiales', 2),
    ('equipos', 'Equipos', 3),
    ('subcontratos', 'Subcontratos', 4);
```

**Campos:**
- `id`: Identificador único del tipo
- `nombre`: Nombre del tipo (único)
- `descripcion`: Descripción del tipo
- `orden`: Orden de visualización

### 3. recursos
Catálogo maestro de recursos (mano de obra, materiales, equipos, subcontratos).

```sql
CREATE TABLE recursos (
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
```

**Campos:**
- `id`: Identificador único del recurso
- `codigo`: Código único del recurso (ej: "470101")
- `descripcion`: Descripción del recurso
- `unidad`: Unidad de medida (hh, kg, m3, etc.)
- `precio_base`: Precio base de referencia
- `tipo_recurso_id`: Referencia al tipo de recurso
- `activo`: Si el recurso está activo
- `created_at`: Fecha de creación
- `updated_at`: Fecha de última actualización

### 4. partidas
Partidas individuales de cada proyecto.

```sql
CREATE TABLE partidas (
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
```

**Campos:**
- `id`: Identificador único de la partida
- `proyecto_id`: Referencia al proyecto
- `codigo`: Código de la partida (ej: "01.01.01")
- `descripcion`: Descripción de la partida
- `unidad`: Unidad de medida
- `rendimiento`: Rendimiento diario
- `costo_total`: Costo total calculado
- `activo`: Si la partida está activa
- **Constraint**: Combinación proyecto_id + codigo debe ser única

### 5. partida_recursos
Relación many-to-many entre partidas y recursos con cantidades específicas.

```sql
CREATE TABLE partida_recursos (
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
```

**Campos:**
- `id`: Identificador único de la relación
- `partida_id`: Referencia a la partida
- `recurso_id`: Referencia al recurso
- `cantidad`: Cantidad del recurso utilizada
- `precio`: Precio específico para esta partida
- `cuadrilla`: Factor de cuadrilla (solo para mano de obra)
- `parcial`: Costo parcial calculado automáticamente
- **Constraint**: Combinación partida_id + recurso_id debe ser única

### 6. analisis_historicos
Tabla para almacenar históricos de análisis y reportes.

```sql
CREATE TABLE analisis_historicos (
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
```

## 📈 Índices

Para optimizar las consultas más comunes:

```sql
-- Índices en partidas
CREATE INDEX IF NOT EXISTS idx_partidas_proyecto_id ON partidas(proyecto_id);
CREATE INDEX IF NOT EXISTS idx_partidas_codigo ON partidas(codigo);

-- Índices en recursos
CREATE INDEX IF NOT EXISTS idx_recursos_codigo ON recursos(codigo);
CREATE INDEX IF NOT EXISTS idx_recursos_tipo ON recursos(tipo_recurso_id);

-- Índices en relaciones
CREATE INDEX IF NOT EXISTS idx_partida_recursos_partida_id ON partida_recursos(partida_id);
CREATE INDEX IF NOT EXISTS idx_partida_recursos_recurso_id ON partida_recursos(recurso_id);
```

## 🔍 Consultas comunes

### Obtener proyecto completo con estadísticas
```sql
SELECT 
    p.id,
    p.nombre,
    p.descripcion,
    COUNT(DISTINCT pa.id) as total_partidas,
    COUNT(DISTINCT pr.recurso_id) as total_recursos,
    SUM(pr.parcial) as costo_total
FROM proyectos p
LEFT JOIN partidas pa ON p.id = pa.proyecto_id
LEFT JOIN partida_recursos pr ON pa.id = pr.partida_id
WHERE p.id = $1
GROUP BY p.id, p.nombre, p.descripcion;
```

### Obtener costos por tipo de recurso
```sql
SELECT 
    tr.nombre as tipo_recurso,
    SUM(pr.parcial) as costo_total
FROM partida_recursos pr
JOIN recursos r ON pr.recurso_id = r.id
JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
JOIN partidas pa ON pr.partida_id = pa.id
WHERE pa.proyecto_id = $1
GROUP BY tr.nombre, tr.orden
ORDER BY tr.orden;
```

### Listar partidas con sus costos
```sql
SELECT 
    pa.codigo,
    pa.descripcion,
    pa.unidad,
    pa.rendimiento,
    COALESCE(SUM(pr.parcial), 0) as costo_total
FROM partidas pa
LEFT JOIN partida_recursos pr ON pa.id = pr.partida_id
WHERE pa.proyecto_id = $1
GROUP BY pa.id, pa.codigo, pa.descripcion, pa.unidad, pa.rendimiento
ORDER BY pa.codigo;
```

## 🔧 Triggers y funciones

### Trigger para actualizar costo total de partidas
```sql
CREATE OR REPLACE FUNCTION update_partida_costo_total()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE partidas 
    SET costo_total = (
        SELECT COALESCE(SUM(cantidad * precio), 0)
        FROM partida_recursos 
        WHERE partida_id = NEW.partida_id
    )
    WHERE id = NEW.partida_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_partida_costo
    AFTER INSERT OR UPDATE OR DELETE ON partida_recursos
    FOR EACH ROW
    EXECUTE FUNCTION update_partida_costo_total();
```

### Trigger para updated_at automático
```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar a todas las tablas principales
CREATE TRIGGER update_proyectos_updated_at 
    BEFORE UPDATE ON proyectos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_partidas_updated_at 
    BEFORE UPDATE ON partidas 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recursos_updated_at 
    BEFORE UPDATE ON recursos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

## 🏗️ Migraciones

### Script de inicialización completa
```sql
-- Crear extensiones
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Ejecutar en orden:
-- 1. tipos_recurso
-- 2. proyectos
-- 3. recursos
-- 4. partidas
-- 5. partida_recursos
-- 6. analisis_historicos
-- 7. Índices
-- 8. Triggers
```

### Backup y restore
```bash
# Backup
pg_dump -h localhost -U postgres goexcel_db > backup.sql

# Restore
psql -h localhost -U postgres goexcel_db < backup.sql
```

## 📊 Estadísticas útiles

### Tamaño de tablas
```sql
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size(tablename::regclass)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(tablename::regclass) DESC;
```

### Conteos por tabla
```sql
SELECT 
    'proyectos' as tabla, COUNT(*) as registros FROM proyectos
UNION ALL
SELECT 'partidas', COUNT(*) FROM partidas
UNION ALL
SELECT 'recursos', COUNT(*) FROM recursos
UNION ALL
SELECT 'partida_recursos', COUNT(*) FROM partida_recursos;
```