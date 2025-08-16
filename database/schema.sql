-- Esquema de base de datos para ACUs (Análisis de Costos Unitarios)
-- PostgreSQL Database Schema

-- Extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabla de organizaciones
CREATE TABLE organizaciones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT,
    logo_url TEXT,
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de usuarios
CREATE TABLE usuarios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    nombre VARCHAR(255) NOT NULL,
    apellido VARCHAR(255),
    rol VARCHAR(50) DEFAULT 'user' CHECK (rol IN ('admin', 'user', 'moderator')),
    organizacion_id UUID REFERENCES organizaciones(id) ON DELETE SET NULL,
    avatar_url TEXT,
    activo BOOLEAN DEFAULT true,
    email_verificado BOOLEAN DEFAULT false,
    ultimo_acceso TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de proyectos
CREATE TABLE proyectos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT,
    ubicacion VARCHAR(255),
    cliente VARCHAR(255),
    fecha_inicio DATE,
    fecha_fin DATE,
    moneda VARCHAR(10) DEFAULT 'PEN',
    usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE,
    organizacion_id UUID REFERENCES organizaciones(id) ON DELETE SET NULL,
    visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('private', 'public', 'featured')),
    template_categoria VARCHAR(50),
    imagen_portada TEXT,
    likes_count INTEGER DEFAULT 0,
    vistas_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de partidas
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

-- Tabla de tipos de recurso
CREATE TABLE tipos_recurso (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(50) NOT NULL UNIQUE,
    descripcion TEXT,
    orden INTEGER DEFAULT 0
);

-- Insertar tipos de recurso por defecto
INSERT INTO tipos_recurso (nombre, descripcion, orden) VALUES
('mano_obra', 'Mano de Obra', 1),
('materiales', 'Materiales', 2),
('equipos', 'Equipos', 3),
('subcontratos', 'Subcontratos', 4);

-- Tabla de recursos
CREATE TABLE recursos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo VARCHAR(50) NOT NULL,
    descripcion TEXT NOT NULL,
    unidad VARCHAR(20) NOT NULL,
    precio_base DECIMAL(15,4) NOT NULL DEFAULT 0,
    tipo_recurso_id UUID REFERENCES tipos_recurso(id),
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(codigo)
);

-- Tabla de recursos por partida (relación many-to-many)
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

-- Tabla de análisis históricos
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

-- Índices para optimizar consultas
CREATE INDEX idx_usuarios_email ON usuarios(email);
CREATE INDEX idx_usuarios_organizacion ON usuarios(organizacion_id);
CREATE INDEX idx_proyectos_usuario_id ON proyectos(usuario_id);
CREATE INDEX idx_proyectos_organizacion_id ON proyectos(organizacion_id);
CREATE INDEX idx_proyectos_visibility ON proyectos(visibility);
CREATE INDEX idx_partidas_proyecto_id ON partidas(proyecto_id);
CREATE INDEX idx_partidas_codigo ON partidas(codigo);
CREATE INDEX idx_recursos_codigo ON recursos(codigo);
CREATE INDEX idx_recursos_tipo ON recursos(tipo_recurso_id);
CREATE INDEX idx_partida_recursos_partida_id ON partida_recursos(partida_id);
CREATE INDEX idx_partida_recursos_recurso_id ON partida_recursos(recurso_id);

-- Función para actualizar timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para actualizar updated_at automáticamente
CREATE TRIGGER update_organizaciones_updated_at BEFORE UPDATE ON organizaciones
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usuarios_updated_at BEFORE UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_proyectos_updated_at BEFORE UPDATE ON proyectos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_partidas_updated_at BEFORE UPDATE ON partidas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recursos_updated_at BEFORE UPDATE ON recursos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_partida_recursos_updated_at BEFORE UPDATE ON partida_recursos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Función para calcular costo total de partida
CREATE OR REPLACE FUNCTION calcular_costo_partida(partida_uuid UUID)
RETURNS DECIMAL(15,4) AS $$
DECLARE
    total DECIMAL(15,4) := 0;
BEGIN
    SELECT COALESCE(SUM(cantidad * precio), 0)
    INTO total
    FROM partida_recursos
    WHERE partida_id = partida_uuid;
    
    RETURN total;
END;
$$ LANGUAGE plpgsql;

-- Trigger para actualizar costo_total cuando se modifica partida_recursos
CREATE OR REPLACE FUNCTION update_partida_costo_total()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE partidas 
        SET costo_total = calcular_costo_partida(OLD.partida_id)
        WHERE id = OLD.partida_id;
        RETURN OLD;
    ELSE
        UPDATE partidas 
        SET costo_total = calcular_costo_partida(NEW.partida_id)
        WHERE id = NEW.partida_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_partida_costo_total_trigger
    AFTER INSERT OR UPDATE OR DELETE ON partida_recursos
    FOR EACH ROW EXECUTE FUNCTION update_partida_costo_total();

-- Vistas útiles
CREATE VIEW vista_partidas_completas AS
SELECT 
    p.id,
    p.codigo,
    p.descripcion,
    p.unidad,
    p.rendimiento,
    p.costo_total,
    pr.nombre as proyecto_nombre,
    COALESCE(mo.total, 0) as costo_mano_obra,
    COALESCE(mat.total, 0) as costo_materiales,
    COALESCE(eq.total, 0) as costo_equipos,
    COALESCE(sub.total, 0) as costo_subcontratos
FROM partidas p
LEFT JOIN proyectos pr ON p.proyecto_id = pr.id
LEFT JOIN (
    SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
    FROM partida_recursos pr
    JOIN recursos r ON pr.recurso_id = r.id
    JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
    WHERE tr.nombre = 'mano_obra'
    GROUP BY pr.partida_id
) mo ON p.id = mo.partida_id
LEFT JOIN (
    SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
    FROM partida_recursos pr
    JOIN recursos r ON pr.recurso_id = r.id
    JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
    WHERE tr.nombre = 'materiales'
    GROUP BY pr.partida_id
) mat ON p.id = mat.partida_id
LEFT JOIN (
    SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
    FROM partida_recursos pr
    JOIN recursos r ON pr.recurso_id = r.id
    JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
    WHERE tr.nombre = 'equipos'
    GROUP BY pr.partida_id
) eq ON p.id = eq.partida_id
LEFT JOIN (
    SELECT pr.partida_id, SUM(pr.cantidad * pr.precio) as total
    FROM partida_recursos pr
    JOIN recursos r ON pr.recurso_id = r.id
    JOIN tipos_recurso tr ON r.tipo_recurso_id = tr.id
    WHERE tr.nombre = 'subcontratos'
    GROUP BY pr.partida_id
) sub ON p.id = sub.partida_id;

-- Tabla de sesiones de usuario (opcional para logout global)
CREATE TABLE sesiones_usuario (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de likes en proyectos
CREATE TABLE proyecto_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    proyecto_id UUID REFERENCES proyectos(id) ON DELETE CASCADE,
    usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proyecto_id, usuario_id)
);

-- Función para actualizar contador de likes
CREATE OR REPLACE FUNCTION update_proyecto_likes_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE proyectos 
        SET likes_count = (
            SELECT COUNT(*) FROM proyecto_likes WHERE proyecto_id = OLD.proyecto_id
        )
        WHERE id = OLD.proyecto_id;
        RETURN OLD;
    ELSE
        UPDATE proyectos 
        SET likes_count = (
            SELECT COUNT(*) FROM proyecto_likes WHERE proyecto_id = NEW.proyecto_id
        )
        WHERE id = NEW.proyecto_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Trigger para actualizar likes_count automáticamente
CREATE TRIGGER update_proyecto_likes_count_trigger
    AFTER INSERT OR DELETE ON proyecto_likes
    FOR EACH ROW EXECUTE FUNCTION update_proyecto_likes_count();

-- Índices adicionales
CREATE INDEX idx_sesiones_usuario_id ON sesiones_usuario(usuario_id);
CREATE INDEX idx_sesiones_expires_at ON sesiones_usuario(expires_at);
CREATE INDEX idx_proyecto_likes_proyecto_id ON proyecto_likes(proyecto_id);
CREATE INDEX idx_proyecto_likes_usuario_id ON proyecto_likes(usuario_id);

-- Insertar organización por defecto
INSERT INTO organizaciones (id, nombre, descripcion) VALUES 
('00000000-0000-0000-0000-000000000001', 'PresupuestosAI', 'Organización principal del sistema');

-- Insertar usuario administrador por defecto (password: admin123)
INSERT INTO usuarios (id, email, password_hash, nombre, rol, organizacion_id) VALUES 
('00000000-0000-0000-0000-000000000001', 'admin@presupuestosai.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Administrador', 'admin', '00000000-0000-0000-0000-000000000001');

-- Limpiar sesiones expiradas (función de mantenimiento)
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sesiones_usuario WHERE expires_at < CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;