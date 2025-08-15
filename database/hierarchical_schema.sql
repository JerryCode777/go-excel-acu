-- Nuevo esquema jerárquico para partidas con títulos automáticos
-- PostgreSQL Database Schema for Hierarchical Partidas

-- Tabla de elementos jerárquicos (incluye tanto partidas como títulos)
CREATE TABLE elementos_jerarquicos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    proyecto_id UUID REFERENCES proyectos(id) ON DELETE CASCADE,
    codigo VARCHAR(50) NOT NULL,
    descripcion TEXT NOT NULL,
    tipo_elemento VARCHAR(20) NOT NULL CHECK (tipo_elemento IN ('titulo', 'partida')),
    nivel INTEGER NOT NULL, -- Número de niveles en el código (ej: 01.01.02 = 3)
    codigo_padre VARCHAR(50), -- Código del elemento padre (ej: 01.01 para 01.01.02)
    unidad VARCHAR(20), -- Solo para partidas
    rendimiento DECIMAL(15,6) DEFAULT 1.0, -- Solo para partidas
    costo_total DECIMAL(15,4) DEFAULT 0, -- Solo para partidas
    orden_display INTEGER, -- Para ordenamiento en vista
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proyecto_id, codigo)
);

-- Índices para el nuevo esquema
CREATE INDEX idx_elementos_proyecto_id ON elementos_jerarquicos(proyecto_id);
CREATE INDEX idx_elementos_codigo ON elementos_jerarquicos(codigo);
CREATE INDEX idx_elementos_codigo_padre ON elementos_jerarquicos(codigo_padre);
CREATE INDEX idx_elementos_tipo ON elementos_jerarquicos(tipo_elemento);
CREATE INDEX idx_elementos_nivel ON elementos_jerarquicos(nivel);
CREATE INDEX idx_elementos_orden ON elementos_jerarquicos(orden_display);

-- Función para generar título automático basado en código
CREATE OR REPLACE FUNCTION generar_titulo_automatico(codigo_input VARCHAR(50))
RETURNS TEXT AS $$
DECLARE
    partes TEXT[];
    nivel INTEGER;
    titulo TEXT;
BEGIN
    partes := string_to_array(codigo_input, '.');
    nivel := array_length(partes, 1);
    
    -- Generar títulos jerárquicos estándar de construcción
    CASE nivel
        WHEN 1 THEN 
            CASE partes[1]
                WHEN '01' THEN titulo := 'OBRAS PROVISIONALES Y TRABAJOS PRELIMINARES';
                WHEN '02' THEN titulo := 'ESTRUCTURAS';
                WHEN '03' THEN titulo := 'ARQUITECTURA';
                WHEN '04' THEN titulo := 'INSTALACIONES SANITARIAS';
                WHEN '05' THEN titulo := 'INSTALACIONES ELÉCTRICAS';
                WHEN '06' THEN titulo := 'INSTALACIONES MECÁNICAS';
                WHEN '07' THEN titulo := 'INSTALACIONES ESPECIALES';
                WHEN '08' THEN titulo := 'EQUIPAMIENTO';
                WHEN '09' THEN titulo := 'VARIOS Y OTROS';
                ELSE titulo := 'TITULO NIVEL ' || partes[1];
            END CASE;
        WHEN 2 THEN
            titulo := 'SUBTITULO ' || codigo_input;
            -- Aquí se pueden agregar más reglas específicas según necesidad
        WHEN 3 THEN
            titulo := 'GRUPO ' || codigo_input;
        WHEN 4 THEN
            titulo := 'SUBGRUPO ' || codigo_input;
        ELSE
            titulo := 'NIVEL ' || nivel || ' - ' || codigo_input;
    END CASE;
    
    RETURN titulo;
END;
$$ LANGUAGE plpgsql;

-- Función para extraer código padre
CREATE OR REPLACE FUNCTION obtener_codigo_padre(codigo_input VARCHAR(50))
RETURNS VARCHAR(50) AS $$
DECLARE
    partes TEXT[];
    nivel INTEGER;
    codigo_padre VARCHAR(50);
BEGIN
    partes := string_to_array(codigo_input, '.');
    nivel := array_length(partes, 1);
    
    IF nivel <= 1 THEN
        RETURN NULL; -- No tiene padre
    END IF;
    
    -- Remover el último nivel para obtener el padre
    codigo_padre := array_to_string(partes[1:nivel-1], '.');
    
    RETURN codigo_padre;
END;
$$ LANGUAGE plpgsql;

-- Función para insertar/actualizar elementos jerárquicos automáticamente
CREATE OR REPLACE FUNCTION insertar_elemento_con_jerarquia(
    p_proyecto_id UUID,
    p_codigo VARCHAR(50),
    p_descripcion TEXT,
    p_unidad VARCHAR(20) DEFAULT NULL,
    p_rendimiento DECIMAL(15,6) DEFAULT 1.0
)
RETURNS UUID AS $$
DECLARE
    elemento_id UUID;
    codigo_padre VARCHAR(50);
    nivel INTEGER;
    partes TEXT[];
    i INTEGER;
    codigo_temp VARCHAR(50);
    padre_existe BOOLEAN;
BEGIN
    partes := string_to_array(p_codigo, '.');
    nivel := array_length(partes, 1);
    
    -- Crear todos los elementos padre si no existen
    FOR i IN 1..(nivel-1) LOOP
        codigo_temp := array_to_string(partes[1:i], '.');
        
        SELECT EXISTS(
            SELECT 1 FROM elementos_jerarquicos 
            WHERE proyecto_id = p_proyecto_id AND codigo = codigo_temp
        ) INTO padre_existe;
        
        IF NOT padre_existe THEN
            INSERT INTO elementos_jerarquicos (
                proyecto_id, 
                codigo, 
                descripcion, 
                tipo_elemento, 
                nivel, 
                codigo_padre,
                orden_display
            ) VALUES (
                p_proyecto_id,
                codigo_temp,
                generar_titulo_automatico(codigo_temp),
                'titulo',
                i,
                CASE WHEN i > 1 THEN array_to_string(partes[1:i-1], '.') ELSE NULL END,
                i * 1000 + COALESCE(partes[i]::int, 0) -- Para ordenamiento natural
            );
        END IF;
    END LOOP;
    
    -- Insertar/actualizar el elemento principal (partida)
    codigo_padre := obtener_codigo_padre(p_codigo);
    
    INSERT INTO elementos_jerarquicos (
        proyecto_id,
        codigo,
        descripcion,
        tipo_elemento,
        nivel,
        codigo_padre,
        unidad,
        rendimiento,
        orden_display
    ) VALUES (
        p_proyecto_id,
        p_codigo,
        p_descripcion,
        'partida',
        nivel,
        codigo_padre,
        p_unidad,
        p_rendimiento,
        nivel * 1000 + COALESCE(partes[nivel]::int, 0) -- Para ordenamiento natural
    ) 
    ON CONFLICT (proyecto_id, codigo) 
    DO UPDATE SET
        descripcion = EXCLUDED.descripcion,
        unidad = EXCLUDED.unidad,
        rendimiento = EXCLUDED.rendimiento,
        updated_at = CURRENT_TIMESTAMP
    RETURNING id INTO elemento_id;
    
    RETURN elemento_id;
END;
$$ LANGUAGE plpgsql;

-- Vista para obtener estructura jerárquica completa
CREATE VIEW vista_jerarquia_completa AS
WITH RECURSIVE jerarquia AS (
    -- Elementos raíz (nivel 1)
    SELECT 
        e.id,
        e.proyecto_id,
        e.codigo,
        e.descripcion,
        e.tipo_elemento,
        e.nivel,
        e.codigo_padre,
        e.unidad,
        e.rendimiento,
        e.costo_total,
        e.orden_display,
        ARRAY[e.codigo]::VARCHAR[] as ruta,
        e.codigo as codigo_completo,
        0 as profundidad
    FROM elementos_jerarquicos e
    WHERE e.codigo_padre IS NULL
    
    UNION ALL
    
    -- Elementos hijos
    SELECT 
        e.id,
        e.proyecto_id,
        e.codigo,
        e.descripcion,
        e.tipo_elemento,
        e.nivel,
        e.codigo_padre,
        e.unidad,
        e.rendimiento,
        e.costo_total,
        e.orden_display,
        j.ruta || e.codigo,
        e.codigo as codigo_completo,
        j.profundidad + 1
    FROM elementos_jerarquicos e
    INNER JOIN jerarquia j ON e.codigo_padre = j.codigo AND e.proyecto_id = j.proyecto_id
)
SELECT * FROM jerarquia
ORDER BY proyecto_id, ruta;

-- Actualizar tabla partidas existente para compatibility
ALTER TABLE partidas ADD COLUMN elemento_jerarquico_id UUID REFERENCES elementos_jerarquicos(id);
CREATE INDEX idx_partidas_elemento_id ON partidas(elemento_jerarquico_id);

-- Función para migrar partidas existentes al nuevo esquema
CREATE OR REPLACE FUNCTION migrar_a_esquema_jerarquico(p_proyecto_id UUID)
RETURNS INTEGER AS $$
DECLARE
    partida_record RECORD;
    elemento_id UUID;
    contador INTEGER := 0;
BEGIN
    FOR partida_record IN 
        SELECT * FROM partidas WHERE proyecto_id = p_proyecto_id ORDER BY codigo
    LOOP
        -- Insertar con jerarquía automática
        elemento_id := insertar_elemento_con_jerarquia(
            p_proyecto_id,
            partida_record.codigo,
            partida_record.descripcion,
            partida_record.unidad,
            partida_record.rendimiento
        );
        
        -- Vincular partida con elemento jerárquico
        UPDATE partidas 
        SET elemento_jerarquico_id = elemento_id
        WHERE id = partida_record.id;
        
        contador := contador + 1;
    END LOOP;
    
    RETURN contador;
END;
$$ LANGUAGE plpgsql;

-- Trigger para mantener sincronización
CREATE OR REPLACE FUNCTION sync_partida_elemento()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Auto crear elemento jerárquico para nuevas partidas
        NEW.elemento_jerarquico_id := insertar_elemento_con_jerarquia(
            NEW.proyecto_id,
            NEW.codigo,
            NEW.descripcion,
            NEW.unidad,
            NEW.rendimiento
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Actualizar elemento jerárquico cuando se actualiza partida
        UPDATE elementos_jerarquicos 
        SET 
            descripcion = NEW.descripcion,
            unidad = NEW.unidad,
            rendimiento = NEW.rendimiento,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.elemento_jerarquico_id;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_partida_elemento_trigger
    BEFORE INSERT OR UPDATE ON partidas
    FOR EACH ROW EXECUTE FUNCTION sync_partida_elemento();