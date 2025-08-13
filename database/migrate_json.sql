-- Script para migrar datos desde JSON existente a PostgreSQL
-- Ejecutar después de crear las tablas base

-- Insertar un proyecto por defecto si no existe
INSERT INTO proyectos (id, nombre, descripcion, moneda)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    'Proyecto de Construcción - Migración JSON',
    'Proyecto creado automáticamente desde migración de datos JSON',
    'PEN'
) ON CONFLICT (id) DO NOTHING;

-- Función temporal para procesar migración JSON
CREATE OR REPLACE FUNCTION migrate_json_data(json_data JSONB, proyecto_id UUID DEFAULT '550e8400-e29b-41d4-a716-446655440001')
RETURNS INTEGER AS $$
DECLARE
    partida_item JSONB;
    recurso_item JSONB;
    partida_id UUID;
    recurso_id UUID;
    tipo_recurso_id UUID;
    recursos_insertados INTEGER := 0;
    partidas_insertadas INTEGER := 0;
BEGIN
    -- Procesar cada partida en el JSON
    FOR partida_item IN SELECT * FROM jsonb_array_elements(json_data)
    LOOP
        -- Insertar la partida
        INSERT INTO partidas (
            proyecto_id, codigo, descripcion, unidad, rendimiento
        ) VALUES (
            proyecto_id,
            partida_item->>'codigo',
            partida_item->>'descripcion',
            partida_item->>'unidad',
            COALESCE((partida_item->>'rendimiento')::DECIMAL, 1.0)
        ) ON CONFLICT (proyecto_id, codigo) DO UPDATE SET
            descripcion = EXCLUDED.descripcion,
            unidad = EXCLUDED.unidad,
            rendimiento = EXCLUDED.rendimiento
        RETURNING id INTO partida_id;
        
        partidas_insertadas := partidas_insertadas + 1;
        
        -- Procesar mano de obra
        IF partida_item ? 'mano_obra' THEN
            SELECT id INTO tipo_recurso_id FROM tipos_recurso WHERE nombre = 'mano_obra';
            
            FOR recurso_item IN SELECT * FROM jsonb_array_elements(partida_item->'mano_obra')
            LOOP
                -- Insertar o actualizar recurso
                INSERT INTO recursos (
                    codigo, descripcion, unidad, tipo_recurso_id, precio_base
                ) VALUES (
                    recurso_item->>'codigo',
                    recurso_item->>'descripcion',
                    recurso_item->>'unidad',
                    tipo_recurso_id,
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0)
                ) ON CONFLICT (codigo) DO UPDATE SET
                    descripcion = EXCLUDED.descripcion,
                    unidad = EXCLUDED.unidad,
                    precio_base = EXCLUDED.precio_base
                RETURNING id INTO recurso_id;
                
                -- Insertar relación partida-recurso
                INSERT INTO partida_recursos (
                    partida_id, recurso_id, cantidad, precio, cuadrilla
                ) VALUES (
                    partida_id,
                    recurso_id,
                    COALESCE((recurso_item->>'cantidad')::DECIMAL, 0),
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0),
                    CASE 
                        WHEN recurso_item ? 'cuadrilla' AND (recurso_item->>'cuadrilla')::DECIMAL > 0 
                        THEN (recurso_item->>'cuadrilla')::DECIMAL
                        ELSE NULL 
                    END
                ) ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
                    cantidad = EXCLUDED.cantidad,
                    precio = EXCLUDED.precio,
                    cuadrilla = EXCLUDED.cuadrilla;
                
                recursos_insertados := recursos_insertados + 1;
            END LOOP;
        END IF;
        
        -- Procesar materiales
        IF partida_item ? 'materiales' THEN
            SELECT id INTO tipo_recurso_id FROM tipos_recurso WHERE nombre = 'materiales';
            
            FOR recurso_item IN SELECT * FROM jsonb_array_elements(partida_item->'materiales')
            LOOP
                INSERT INTO recursos (
                    codigo, descripcion, unidad, tipo_recurso_id, precio_base
                ) VALUES (
                    recurso_item->>'codigo',
                    recurso_item->>'descripcion',
                    recurso_item->>'unidad',
                    tipo_recurso_id,
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0)
                ) ON CONFLICT (codigo) DO UPDATE SET
                    descripcion = EXCLUDED.descripcion,
                    unidad = EXCLUDED.unidad,
                    precio_base = EXCLUDED.precio_base
                RETURNING id INTO recurso_id;
                
                INSERT INTO partida_recursos (
                    partida_id, recurso_id, cantidad, precio, cuadrilla
                ) VALUES (
                    partida_id,
                    recurso_id,
                    COALESCE((recurso_item->>'cantidad')::DECIMAL, 0),
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0),
                    CASE 
                        WHEN recurso_item ? 'cuadrilla' AND (recurso_item->>'cuadrilla')::DECIMAL > 0 
                        THEN (recurso_item->>'cuadrilla')::DECIMAL
                        ELSE NULL 
                    END
                ) ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
                    cantidad = EXCLUDED.cantidad,
                    precio = EXCLUDED.precio,
                    cuadrilla = EXCLUDED.cuadrilla;
                
                recursos_insertados := recursos_insertados + 1;
            END LOOP;
        END IF;
        
        -- Procesar equipos
        IF partida_item ? 'equipos' THEN
            SELECT id INTO tipo_recurso_id FROM tipos_recurso WHERE nombre = 'equipos';
            
            FOR recurso_item IN SELECT * FROM jsonb_array_elements(partida_item->'equipos')
            LOOP
                INSERT INTO recursos (
                    codigo, descripcion, unidad, tipo_recurso_id, precio_base
                ) VALUES (
                    recurso_item->>'codigo',
                    recurso_item->>'descripcion',
                    recurso_item->>'unidad',
                    tipo_recurso_id,
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0)
                ) ON CONFLICT (codigo) DO UPDATE SET
                    descripcion = EXCLUDED.descripcion,
                    unidad = EXCLUDED.unidad,
                    precio_base = EXCLUDED.precio_base
                RETURNING id INTO recurso_id;
                
                INSERT INTO partida_recursos (
                    partida_id, recurso_id, cantidad, precio, cuadrilla
                ) VALUES (
                    partida_id,
                    recurso_id,
                    COALESCE((recurso_item->>'cantidad')::DECIMAL, 0),
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0),
                    CASE 
                        WHEN recurso_item ? 'cuadrilla' AND (recurso_item->>'cuadrilla')::DECIMAL > 0 
                        THEN (recurso_item->>'cuadrilla')::DECIMAL
                        ELSE NULL 
                    END
                ) ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
                    cantidad = EXCLUDED.cantidad,
                    precio = EXCLUDED.precio,
                    cuadrilla = EXCLUDED.cuadrilla;
                
                recursos_insertados := recursos_insertados + 1;
            END LOOP;
        END IF;
        
        -- Procesar subcontratos
        IF partida_item ? 'subcontratos' THEN
            SELECT id INTO tipo_recurso_id FROM tipos_recurso WHERE nombre = 'subcontratos';
            
            FOR recurso_item IN SELECT * FROM jsonb_array_elements(partida_item->'subcontratos')
            LOOP
                INSERT INTO recursos (
                    codigo, descripcion, unidad, tipo_recurso_id, precio_base
                ) VALUES (
                    recurso_item->>'codigo',
                    recurso_item->>'descripcion',
                    recurso_item->>'unidad',
                    tipo_recurso_id,
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0)
                ) ON CONFLICT (codigo) DO UPDATE SET
                    descripcion = EXCLUDED.descripcion,
                    unidad = EXCLUDED.unidad,
                    precio_base = EXCLUDED.precio_base
                RETURNING id INTO recurso_id;
                
                INSERT INTO partida_recursos (
                    partida_id, recurso_id, cantidad, precio, cuadrilla
                ) VALUES (
                    partida_id,
                    recurso_id,
                    COALESCE((recurso_item->>'cantidad')::DECIMAL, 0),
                    COALESCE((recurso_item->>'precio')::DECIMAL, 0),
                    CASE 
                        WHEN recurso_item ? 'cuadrilla' AND (recurso_item->>'cuadrilla')::DECIMAL > 0 
                        THEN (recurso_item->>'cuadrilla')::DECIMAL
                        ELSE NULL 
                    END
                ) ON CONFLICT (partida_id, recurso_id) DO UPDATE SET
                    cantidad = EXCLUDED.cantidad,
                    precio = EXCLUDED.precio,
                    cuadrilla = EXCLUDED.cuadrilla;
                
                recursos_insertados := recursos_insertados + 1;
            END LOOP;
        END IF;
    END LOOP;
    
    RETURN partidas_insertadas;
END;
$$ LANGUAGE plpgsql;

-- Instrucciones para usar la función de migración:
-- 1. Cargar el contenido JSON en una variable o tabla temporal
-- 2. Ejecutar: SELECT migrate_json_data('contenido_json_aqui'::JSONB);

-- Ejemplo de uso (comentado, descomentar y ajustar según sea necesario):
-- SELECT migrate_json_data('[
--   {
--     "codigo": "01.01.01.01", 
--     "descripcion": "CARTEL DE IDENTIFICACIÓN DE OBRA",
--     "unidad": "und",
--     "rendimiento": 1.0000,
--     "mano_obra": [...],
--     "materiales": [...],
--     "equipos": [...],
--     "subcontratos": [...]
--   }
-- ]'::JSONB);