-- Migración para sistema jerárquico de presupuestos
-- Soporte para presupuestos, subpresupuestos, títulos anidados hasta 10 niveles y partidas

-- Tabla de presupuestos (proyecto principal)
CREATE TABLE IF NOT EXISTS presupuestos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo VARCHAR(50) NOT NULL,
    nombre TEXT NOT NULL,
    cliente TEXT,
    lugar TEXT,
    moneda VARCHAR(10) DEFAULT 'PEN',
    fecha_creacion DATE DEFAULT CURRENT_DATE,
    usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE,
    organizacion_id UUID REFERENCES organizaciones(id) ON DELETE SET NULL,
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(codigo)
);

-- Tabla de subpresupuestos
CREATE TABLE IF NOT EXISTS subpresupuestos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    presupuesto_id UUID REFERENCES presupuestos(id) ON DELETE CASCADE,
    codigo VARCHAR(50) NOT NULL,
    nombre TEXT NOT NULL,
    orden INTEGER DEFAULT 0,
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(presupuesto_id, codigo)
);

-- Tabla de títulos jerárquicos (hasta 10 niveles)
CREATE TABLE IF NOT EXISTS titulos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    presupuesto_id UUID REFERENCES presupuestos(id) ON DELETE CASCADE,
    subpresupuesto_id UUID REFERENCES subpresupuestos(id) ON DELETE CASCADE,
    titulo_padre_id UUID REFERENCES titulos(id) ON DELETE CASCADE,
    nivel INTEGER NOT NULL CHECK (nivel >= 1 AND nivel <= 10),
    numero INTEGER NOT NULL,
    codigo_completo VARCHAR(50) NOT NULL, -- Ej: "01.01.02.01"
    nombre TEXT NOT NULL,
    orden INTEGER DEFAULT 0,
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(presupuesto_id, codigo_completo)
);

-- Actualizar tabla de partidas para soporte jerárquico
ALTER TABLE partidas ADD COLUMN IF NOT EXISTS presupuesto_id UUID REFERENCES presupuestos(id) ON DELETE CASCADE;
ALTER TABLE partidas ADD COLUMN IF NOT EXISTS subpresupuesto_id UUID REFERENCES subpresupuestos(id) ON DELETE CASCADE;
ALTER TABLE partidas ADD COLUMN IF NOT EXISTS titulo_id UUID REFERENCES titulos(id) ON DELETE CASCADE;
ALTER TABLE partidas ADD COLUMN IF NOT EXISTS numero INTEGER DEFAULT 0;
ALTER TABLE partidas ADD COLUMN IF NOT EXISTS orden INTEGER DEFAULT 0;

-- Actualizar constraint único de partidas
ALTER TABLE partidas DROP CONSTRAINT IF EXISTS partidas_proyecto_id_codigo_key;
ALTER TABLE partidas ADD CONSTRAINT partidas_presupuesto_codigo_unique UNIQUE(presupuesto_id, codigo);

-- Índices para optimizar consultas jerárquicas
CREATE INDEX IF NOT EXISTS idx_presupuestos_codigo ON presupuestos(codigo);
CREATE INDEX IF NOT EXISTS idx_subpresupuestos_presupuesto ON subpresupuestos(presupuesto_id);
CREATE INDEX IF NOT EXISTS idx_titulos_presupuesto ON titulos(presupuesto_id);
CREATE INDEX IF NOT EXISTS idx_titulos_subpresupuesto ON titulos(subpresupuesto_id);
CREATE INDEX IF NOT EXISTS idx_titulos_padre ON titulos(titulo_padre_id);
CREATE INDEX IF NOT EXISTS idx_titulos_nivel ON titulos(nivel);
CREATE INDEX IF NOT EXISTS idx_partidas_presupuesto ON partidas(presupuesto_id);
CREATE INDEX IF NOT EXISTS idx_partidas_titulo ON partidas(titulo_id);

-- Triggers para timestamps
CREATE TRIGGER update_presupuestos_updated_at BEFORE UPDATE ON presupuestos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subpresupuestos_updated_at BEFORE UPDATE ON subpresupuestos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_titulos_updated_at BEFORE UPDATE ON titulos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Función para generar código jerárquico automático
CREATE OR REPLACE FUNCTION generar_codigo_jerarquico(
    presupuesto_uuid UUID,
    subpresupuesto_uuid UUID DEFAULT NULL,
    titulo_padre_uuid UUID DEFAULT NULL,
    nivel_param INTEGER DEFAULT 1
) RETURNS VARCHAR(50) AS $$
DECLARE
    codigo_base VARCHAR(50) := '';
    siguiente_numero INTEGER;
BEGIN
    -- Si hay título padre, obtener su código como base
    IF titulo_padre_uuid IS NOT NULL THEN
        SELECT codigo_completo INTO codigo_base
        FROM titulos 
        WHERE id = titulo_padre_uuid;
        
        -- Obtener siguiente número en este nivel
        SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
        FROM titulos 
        WHERE titulo_padre_id = titulo_padre_uuid;
        
        -- Construir código completo
        RETURN codigo_base || '.' || LPAD(siguiente_numero::TEXT, 2, '0');
    ELSE
        -- Es un título de primer nivel
        IF subpresupuesto_uuid IS NOT NULL THEN
            -- Obtener siguiente número para este subpresupuesto
            SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
            FROM titulos 
            WHERE subpresupuesto_id = subpresupuesto_uuid 
            AND titulo_padre_id IS NULL;
        ELSE
            -- Obtener siguiente número para este presupuesto
            SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
            FROM titulos 
            WHERE presupuesto_id = presupuesto_uuid 
            AND titulo_padre_id IS NULL;
        END IF;
        
        RETURN LPAD(siguiente_numero::TEXT, 2, '0');
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Función para generar código de partida automático
CREATE OR REPLACE FUNCTION generar_codigo_partida(
    presupuesto_uuid UUID,
    titulo_uuid UUID DEFAULT NULL
) RETURNS VARCHAR(50) AS $$
DECLARE
    codigo_base VARCHAR(50) := '';
    siguiente_numero INTEGER;
BEGIN
    IF titulo_uuid IS NOT NULL THEN
        -- Obtener código del título padre
        SELECT codigo_completo INTO codigo_base
        FROM titulos 
        WHERE id = titulo_uuid;
        
        -- Obtener siguiente número de partida en este título
        SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
        FROM partidas 
        WHERE titulo_id = titulo_uuid;
        
        -- Construir código completo
        RETURN codigo_base || '.' || LPAD(siguiente_numero::TEXT, 2, '0');
    ELSE
        -- Partida sin título (no recomendado pero posible)
        SELECT COALESCE(MAX(numero), 0) + 1 INTO siguiente_numero
        FROM partidas 
        WHERE presupuesto_id = presupuesto_uuid 
        AND titulo_id IS NULL;
        
        RETURN LPAD(siguiente_numero::TEXT, 2, '0');
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Vista jerárquica completa
CREATE OR REPLACE VIEW vista_estructura_jerarquica AS
WITH RECURSIVE jerarquia AS (
    -- Nivel base: títulos sin padre
    SELECT 
        t.id,
        t.presupuesto_id,
        t.subpresupuesto_id,
        t.titulo_padre_id,
        t.nivel,
        t.numero,
        t.codigo_completo,
        t.nombre,
        t.orden,
        ARRAY[t.orden] as path_orden,
        1 as depth
    FROM titulos t
    WHERE t.titulo_padre_id IS NULL
    
    UNION ALL
    
    -- Niveles anidados
    SELECT 
        t.id,
        t.presupuesto_id,
        t.subpresupuesto_id,
        t.titulo_padre_id,
        t.nivel,
        t.numero,
        t.codigo_completo,
        t.nombre,
        t.orden,
        j.path_orden || t.orden,
        j.depth + 1
    FROM titulos t
    INNER JOIN jerarquia j ON t.titulo_padre_id = j.id
)
SELECT 
    p.id as presupuesto_id,
    p.codigo as presupuesto_codigo,
    p.nombre as presupuesto_nombre,
    sp.id as subpresupuesto_id,
    sp.codigo as subpresupuesto_codigo,
    sp.nombre as subpresupuesto_nombre,
    j.id as titulo_id,
    j.nivel,
    j.codigo_completo as titulo_codigo,
    j.nombre as titulo_nombre,
    j.depth,
    j.path_orden,
    COUNT(pa.id) as total_partidas,
    COALESCE(SUM(pa.costo_total), 0) as costo_total_titulos
FROM presupuestos p
LEFT JOIN subpresupuestos sp ON p.id = sp.presupuesto_id
LEFT JOIN jerarquia j ON p.id = j.presupuesto_id
LEFT JOIN partidas pa ON j.id = pa.titulo_id
GROUP BY p.id, p.codigo, p.nombre, sp.id, sp.codigo, sp.nombre, 
         j.id, j.nivel, j.codigo_completo, j.nombre, j.depth, j.path_orden
ORDER BY p.codigo, sp.orden, j.path_orden;

-- Vista de partidas con jerarquía
CREATE OR REPLACE VIEW vista_partidas_jerarquicas AS
SELECT 
    p.id as partida_id,
    p.codigo as partida_codigo,
    p.descripcion as partida_descripcion,
    p.unidad,
    p.rendimiento,
    p.costo_total,
    p.numero as partida_numero,
    p.orden as partida_orden,
    pr.id as presupuesto_id,
    pr.codigo as presupuesto_codigo,
    pr.nombre as presupuesto_nombre,
    sp.id as subpresupuesto_id,
    sp.codigo as subpresupuesto_codigo,
    sp.nombre as subpresupuesto_nombre,
    t.id as titulo_id,
    t.codigo_completo as titulo_codigo,
    t.nombre as titulo_nombre,
    t.nivel as titulo_nivel
FROM partidas p
LEFT JOIN presupuestos pr ON p.presupuesto_id = pr.id
LEFT JOIN subpresupuestos sp ON p.subpresupuesto_id = sp.id
LEFT JOIN titulos t ON p.titulo_id = t.id
ORDER BY pr.codigo, sp.orden, t.codigo_completo, p.orden;

-- Función para calcular costo total por título
CREATE OR REPLACE FUNCTION calcular_costo_titulo(titulo_uuid UUID)
RETURNS DECIMAL(15,4) AS $$
DECLARE
    total DECIMAL(15,4) := 0;
BEGIN
    -- Sumar costos de partidas directas del título
    SELECT COALESCE(SUM(p.costo_total), 0) INTO total
    FROM partidas p
    WHERE p.titulo_id = titulo_uuid;
    
    -- Sumar costos de títulos hijos (recursivo)
    total := total + (
        SELECT COALESCE(SUM(calcular_costo_titulo(t.id)), 0)
        FROM titulos t
        WHERE t.titulo_padre_id = titulo_uuid
    );
    
    RETURN total;
END;
$$ LANGUAGE plpgsql;

-- Función para obtener resumen jerárquico
CREATE OR REPLACE FUNCTION obtener_resumen_jerarquico(presupuesto_uuid UUID)
RETURNS TABLE(
    total_subpresupuestos BIGINT,
    total_titulos BIGINT,
    total_partidas BIGINT,
    costo_total DECIMAL(15,4),
    niveles_maximos INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*) FROM subpresupuestos WHERE presupuesto_id = presupuesto_uuid) as total_subpresupuestos,
        (SELECT COUNT(*) FROM titulos WHERE presupuesto_id = presupuesto_uuid) as total_titulos,
        (SELECT COUNT(*) FROM partidas WHERE presupuesto_id = presupuesto_uuid) as total_partidas,
        (SELECT COALESCE(SUM(p.costo_total), 0) FROM partidas p WHERE p.presupuesto_id = presupuesto_uuid) as costo_total,
        (SELECT COALESCE(MAX(nivel), 0) FROM titulos WHERE presupuesto_id = presupuesto_uuid) as niveles_maximos;
END;
$$ LANGUAGE plpgsql;