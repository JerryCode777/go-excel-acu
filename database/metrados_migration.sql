-- Migración para agregar tabla de metrados por proyecto
-- Esta tabla almacena los metrados específicos de cada partida para cada proyecto

-- Tabla de metrados de partidas por proyecto
CREATE TABLE IF NOT EXISTS metrados_partidas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    proyecto_id UUID REFERENCES proyectos(id) ON DELETE CASCADE,
    partida_codigo VARCHAR(50) NOT NULL,
    metrado DECIMAL(15,6) NOT NULL DEFAULT 0,
    unidad VARCHAR(20),
    observaciones TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proyecto_id, partida_codigo)
);

-- Índices para optimizar consultas
CREATE INDEX IF NOT EXISTS idx_metrados_proyecto_id ON metrados_partidas(proyecto_id);
CREATE INDEX IF NOT EXISTS idx_metrados_partida_codigo ON metrados_partidas(partida_codigo);
CREATE INDEX IF NOT EXISTS idx_metrados_proyecto_partida ON metrados_partidas(proyecto_id, partida_codigo);

-- Trigger para actualizar timestamp
CREATE TRIGGER update_metrados_partidas_updated_at BEFORE UPDATE ON metrados_partidas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Vista para obtener metrados con información de partidas
CREATE VIEW vista_metrados_completos AS
SELECT 
    mp.id,
    mp.proyecto_id,
    mp.partida_codigo,
    mp.metrado,
    mp.unidad as metrado_unidad,
    mp.observaciones,
    p.descripcion as partida_descripcion,
    p.unidad as partida_unidad,
    p.costo_total as costo_unitario,
    (mp.metrado * p.costo_total) as costo_total_partida,
    pr.nombre as proyecto_nombre,
    mp.created_at,
    mp.updated_at
FROM metrados_partidas mp
LEFT JOIN partidas p ON p.codigo = mp.partida_codigo AND p.proyecto_id = mp.proyecto_id
LEFT JOIN proyectos pr ON mp.proyecto_id = pr.id;

-- Función para calcular el costo total del proyecto con metrados
CREATE OR REPLACE FUNCTION calcular_costo_total_proyecto(proyecto_uuid UUID)
RETURNS DECIMAL(15,4) AS $$
DECLARE
    total DECIMAL(15,4) := 0;
BEGIN
    SELECT COALESCE(SUM(mp.metrado * p.costo_total), 0)
    INTO total
    FROM metrados_partidas mp
    JOIN partidas p ON p.codigo = mp.partida_codigo AND p.proyecto_id = mp.proyecto_id
    WHERE mp.proyecto_id = proyecto_uuid;
    
    RETURN total;
END;
$$ LANGUAGE plpgsql;

-- Función para obtener resumen de costos por proyecto
CREATE OR REPLACE FUNCTION obtener_resumen_proyecto(proyecto_uuid UUID)
RETURNS TABLE(
    total_partidas BIGINT,
    costo_directo DECIMAL(15,4),
    partidas_con_metrado BIGINT,
    partidas_sin_metrado BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(p.id) as total_partidas,
        COALESCE(SUM(mp.metrado * p.costo_total), 0) as costo_directo,
        COUNT(mp.id) as partidas_con_metrado,
        COUNT(p.id) - COUNT(mp.id) as partidas_sin_metrado
    FROM partidas p
    LEFT JOIN metrados_partidas mp ON p.codigo = mp.partida_codigo AND p.proyecto_id = mp.proyecto_id
    WHERE p.proyecto_id = proyecto_uuid;
END;
$$ LANGUAGE plpgsql;