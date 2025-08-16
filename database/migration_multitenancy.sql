-- Migración para agregar multi-tenancy al sistema existente

-- Crear tablas de usuarios y organizaciones
CREATE TABLE IF NOT EXISTS organizaciones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT,
    logo_url TEXT,
    activo BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS usuarios (
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

-- Agregar columnas a la tabla proyectos existente (solo si no existen)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'usuario_id') THEN
        ALTER TABLE proyectos ADD COLUMN usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'organizacion_id') THEN
        ALTER TABLE proyectos ADD COLUMN organizacion_id UUID REFERENCES organizaciones(id) ON DELETE SET NULL;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'visibility') THEN
        ALTER TABLE proyectos ADD COLUMN visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('private', 'public', 'featured'));
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'template_categoria') THEN
        ALTER TABLE proyectos ADD COLUMN template_categoria VARCHAR(50);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'imagen_portada') THEN
        ALTER TABLE proyectos ADD COLUMN imagen_portada TEXT;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'likes_count') THEN
        ALTER TABLE proyectos ADD COLUMN likes_count INTEGER DEFAULT 0;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'proyectos' AND column_name = 'vistas_count') THEN
        ALTER TABLE proyectos ADD COLUMN vistas_count INTEGER DEFAULT 0;
    END IF;
END $$;

-- Crear tabla de sesiones de usuario
CREATE TABLE IF NOT EXISTS sesiones_usuario (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Crear tabla de likes en proyectos
CREATE TABLE IF NOT EXISTS proyecto_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    proyecto_id UUID REFERENCES proyectos(id) ON DELETE CASCADE,
    usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proyecto_id, usuario_id)
);

-- Crear triggers para updated_at en nuevas tablas
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_organizaciones_updated_at') THEN
        CREATE TRIGGER update_organizaciones_updated_at BEFORE UPDATE ON organizaciones
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_usuarios_updated_at') THEN
        CREATE TRIGGER update_usuarios_updated_at BEFORE UPDATE ON usuarios
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

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
DROP TRIGGER IF EXISTS update_proyecto_likes_count_trigger ON proyecto_likes;
CREATE TRIGGER update_proyecto_likes_count_trigger
    AFTER INSERT OR DELETE ON proyecto_likes
    FOR EACH ROW EXECUTE FUNCTION update_proyecto_likes_count();

-- Crear índices adicionales solo si no existen
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_usuarios_email') THEN
        CREATE INDEX idx_usuarios_email ON usuarios(email);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_usuarios_organizacion') THEN
        CREATE INDEX idx_usuarios_organizacion ON usuarios(organizacion_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_proyectos_usuario_id') THEN
        CREATE INDEX idx_proyectos_usuario_id ON proyectos(usuario_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_proyectos_organizacion_id') THEN
        CREATE INDEX idx_proyectos_organizacion_id ON proyectos(organizacion_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_proyectos_visibility') THEN
        CREATE INDEX idx_proyectos_visibility ON proyectos(visibility);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sesiones_usuario_id') THEN
        CREATE INDEX idx_sesiones_usuario_id ON sesiones_usuario(usuario_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sesiones_expires_at') THEN
        CREATE INDEX idx_sesiones_expires_at ON sesiones_usuario(expires_at);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_proyecto_likes_proyecto_id') THEN
        CREATE INDEX idx_proyecto_likes_proyecto_id ON proyecto_likes(proyecto_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_proyecto_likes_usuario_id') THEN
        CREATE INDEX idx_proyecto_likes_usuario_id ON proyecto_likes(usuario_id);
    END IF;
END $$;

-- Insertar organización por defecto si no existe
INSERT INTO organizaciones (id, nombre, descripcion) 
VALUES ('00000000-0000-0000-0000-000000000001', 'PresupuestosAI', 'Organización principal del sistema')
ON CONFLICT (id) DO NOTHING;

-- Insertar usuario administrador por defecto si no existe (password: admin123)
INSERT INTO usuarios (id, email, password_hash, nombre, rol, organizacion_id) 
VALUES ('00000000-0000-0000-0000-000000000001', 'admin@presupuestosai.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Administrador', 'admin', '00000000-0000-0000-0000-000000000001')
ON CONFLICT (email) DO NOTHING;

-- Función para limpiar sesiones expiradas
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