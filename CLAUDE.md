# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**GoExcel v2.0** - A scalable Go application that processes construction unit cost analysis (ACU) data with PostgreSQL backend and generates formatted Excel reports. The project reads construction project partidas (work items) with their associated labor, materials, equipment, and subcontract costs, then creates comprehensive Excel workbooks with detailed cost breakdowns and summary sheets.

### Key Features v2.0
- PostgreSQL database for data persistence
- Scalable project structure with separate packages
- Support for subcontracts (previously missing)
- Migration from JSON to database
- API-ready structure for future React integration
- Backward compatibility with legacy JSON mode

## Comandos Principales

### Configuración Inicial
```bash
# Ejecutar script de configuración completa
./setup.sh

# O configuración manual:
make build          # Compilar aplicación
make setup-dirs     # Crear directorios necesarios
```

### Aplicación CLI (Un solo main.go)
La aplicación tiene **un solo punto de entrada**: `cmd/goexcel/main.go`

```bash
# Listar proyectos disponibles en la base de datos
./bin/goexcel

# Migrar datos JSON a PostgreSQL
./bin/goexcel migrate partidas.json

# Generar Excel desde base de datos
./bin/goexcel generate <project_id>

# Modo legacy (sin base de datos)
./bin/goexcel legacy partidas.json

# Iniciar servidor API REST
./bin/goexcel server
```

### Desarrollo Local
```bash
# Compilar
make build
go build -o bin/goexcel cmd/goexcel/main.go

# Ejecutar en modo desarrollo
go run cmd/goexcel/main.go legacy partidas.json
go run cmd/goexcel/main.go server

# Herramientas de desarrollo
make test          # Ejecutar tests
make fmt           # Formatear código
make clean         # Limpiar archivos generados
go mod tidy        # Limpiar dependencias
```

## Architecture v2.0

### Project Structure
```
goexcel/
├── cmd/goexcel/main.go      # Main application entry point
├── config/                  # Configuration management  
├── database/                # SQL schemas and migrations
├── internal/
│   ├── database/           # DB connection and repositories
│   ├── models/             # Data structures and DTOs
│   └── services/           # Business logic
├── bin/                    # Compiled binaries
├── output/                 # Generated Excel files
└── setup.sh               # Installation script
```

### Core Data Models

- **Proyecto**: Project information with metadata
- **Partida**: Work items belonging to projects  
- **Recurso**: Individual resources (labor, materials, equipment, subcontracts)
- **PartidaRecurso**: Many-to-many relationship between partidas and recursos
- **TipoRecurso**: Resource type classification (mano_obra, materiales, equipos, subcontratos)

### Application Modes

1. **Database Mode** (New): 
   - PostgreSQL backend for data persistence
   - Project-based organization
   - Full CRUD operations support
   
2. **Legacy Mode**: 
   - Direct JSON file processing (backward compatibility)
   - Single Excel output
   - No database required

### Main Application Flow (Database Mode)

1. **Migration**: Import existing JSON data into PostgreSQL
2. **Project Management**: Organize partidas by projects
3. **Data Processing**: Read from database with proper relationships
4. **Excel Generation**: Create formatted reports with all resource types including subcontracts
5. **Analysis Storage**: Save analysis metadata for historical tracking

### Key Components

- `database.DB`: PostgreSQL connection with migration support
- `repositories.*`: Data access layer for each entity
- `services.ExcelService`: Excel generation with database integration  
- `config.Config`: Environment-based configuration management

### Dependencies

- `github.com/xuri/excelize/v2`: Excel file generation
- `github.com/lib/pq`: PostgreSQL driver
- `github.com/joho/godotenv`: Environment variable management
- `github.com/google/uuid`: UUID generation for primary keys

## Input Data Format

### JSON Format (Legacy and Migration)
The application supports JSON files with arrays of partidas containing:
- Labor resources (mano_obra)
- Material resources (materiales) 
- Equipment resources (equipos)
- **Subcontract resources (subcontratos)** ⭐ NEW in v2.0
- Each resource includes code, description, unit, quantity, price, and optional cuadrilla

### Database Schema
- Normalized PostgreSQL schema with proper relationships
- UUID primary keys for all entities
- Automatic cost calculations via triggers
- Historical analysis tracking

## Output

Generates Excel files with:
- Detailed cost analysis sheets per partida with **subcontract support** ⭐
- Enhanced "Resumen" summary sheet with subcontract column ⭐
- Professional formatting suitable for construction documentation
- Automatic calculations and totals including subcontracts
- Project-based organization and metadata

## Configuración de Servidor y Base de Datos

### Variables de Entorno
Crear archivo `.env` en la raíz del proyecto:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=goexcel_user
DB_PASSWORD=your_password
DB_NAME=goexcel_db
DB_SSLMODE=disable
```

### Inicio del Servidor API
```bash
# Compilar y ejecutar servidor
make build
./bin/goexcel server

# O ejecutar en modo desarrollo
go run cmd/goexcel/main.go server
```

El servidor se ejecuta por defecto en puerto **8080** y proporciona:
- API REST para manejo de proyectos y partidas
- Endpoints para frontend React
- CORS habilitado para desarrollo local

**Acceso al servidor:**
- API: `http://localhost:8080/api/v1`
- Health Check: `http://localhost:8080/api/v1/health`

**Si el puerto está en uso:**
- Cambiar `SERVER_PORT=8081` en `.env`
- O usar otro puerto libre disponible
- Verificar procesos: `lsof -i :8080`

### Para Clonar en Otra PC
1. **Requisitos**: Go 1.23+, PostgreSQL
2. **Clonar**: `git clone <tu-repo-local>`
3. **Configurar**: Copiar `.env` y ajustar configuración de BD
4. **Instalar**: `go mod download`
5. **Compilar**: `make build` o `./setup.sh`
6. **Ejecutar**: `./bin/goexcel server`

## Testing

```bash
make test                    # Run all tests
go test ./...               # Direct test execution  
make lint                   # Run linter (requires golangci-lint)
```