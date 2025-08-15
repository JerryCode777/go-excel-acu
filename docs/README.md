# GoExcel - Sistema de Gestión de ACUs

## 📋 Descripción

GoExcel es un sistema completo para la gestión de Análisis de Costos Unitarios (ACUs) en construcción. Permite crear, gestionar y generar reportes de ACUs usando tanto formato JSON tradicional como el innovador formato .acu (estilo LaTeX).

## 🏗️ Arquitectura

```
React Frontend ↔ Go API Backend ↔ PostgreSQL Database
                      ↕
              Python AI Service (futuro)
```

## 🚀 Características

### ✅ Implementado
- **Formato .acu**: Sintaxis estilo LaTeX para ACUs
- **API REST**: Endpoints completos para gestión de proyectos
- **Base de datos**: PostgreSQL con normalización
- **CLI avanzado**: Herramientas de línea de comandos
- **Validación**: Sintaxis .acu y datos JSON

### 🔄 En desarrollo
- **Python AI Service**: OCR, especificaciones técnicas, optimización
- **React Frontend**: Interface gráfica completa
- **Generación Excel**: Desde base de datos

## 📁 Estructura del proyecto

```
goexcel/
├── cmd/
│   ├── goexcel/          # CLI principal
│   └── server/           # Servidor HTTP
├── internal/
│   ├── handlers/         # Controladores HTTP
│   ├── models/           # Estructuras de datos
│   ├── services/         # Lógica de negocio
│   ├── database/         # Acceso a datos
│   └── server/           # Configuración servidor
├── config/               # Configuración
├── docs/                 # Documentación
└── web/                  # Frontend React (futuro)
```

## 🛠️ Instalación y configuración

### Prerrequisitos
- Go 1.21+
- PostgreSQL 13+
- Node.js 18+ (para frontend)

### Configuración base de datos
```bash
# Crear base de datos
createdb goexcel_db

# Configurar variables de entorno
echo "DB_USER=postgres" > .env
echo "DB_PASSWORD=tu_password" >> .env
echo "DB_HOST=localhost" >> .env
echo "DB_PORT=5432" >> .env
echo "DB_NAME=goexcel_db" >> .env
```

### Compilación
```bash
# CLI
go build -o bin/goexcel cmd/goexcel/main.go

# Servidor HTTP
go build -o bin/server cmd/server/main.go
```

## 🎯 Uso

### CLI
```bash
# Validar archivo .acu
./bin/goexcel validate-acu proyecto.acu

# Importar proyecto desde .acu
./bin/goexcel import-acu proyecto.acu

# Listar proyectos
./bin/goexcel

# Ver ayuda completa
./bin/goexcel --help
```

### API Server
```bash
# Iniciar servidor
./bin/server

# El servidor estará disponible en:
# http://localhost:8080/api/v1
```

## 📚 Documentación

- [API Reference](./api-reference.md) - Documentación completa de endpoints
- [ACU Format](./acu-format.md) - Especificación del formato .acu
- [Database Schema](./database-schema.md) - Esquema de base de datos
- [Frontend Integration](./frontend-integration.md) - Guía para React
- [Architecture](./architecture.md) - Arquitectura del sistema

## 🤝 Desarrollo

### Comandos útiles
```bash
# Formato código
go fmt ./...

# Tests
go test ./...

# Dependencias
go mod tidy

# Migración DB manual
./bin/goexcel migrate
```

### Flujo de desarrollo
1. Hacer cambios en código
2. Compilar: `go build`
3. Probar con CLI o servidor
4. Ejecutar tests
5. Commit y push

## 📄 Licencia

MIT License