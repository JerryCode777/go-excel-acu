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

## Commands

### Setup (First Time)
```bash
# Run setup script
./setup.sh

# Or manual setup:
make build          # Compile application
make setup-dirs     # Create necessary directories
```

### Database Operations
```bash
# Migrate JSON data to PostgreSQL
./bin/goexcel migrate partidas.json

# List available projects
./bin/goexcel

# Generate Excel from database
./bin/goexcel generate <project_id>
```

### Legacy Mode (JSON only)
```bash
# Use original functionality (no database)
./bin/goexcel legacy partidas.json
./bin/goexcel legacy  # uses partidas.json by default
```

### Development
```bash
# Using Makefile
make build         # Compile application
make run-legacy    # Run in legacy mode
make run-migrate   # Migrate JSON to DB
make test          # Run tests
make clean         # Clean build artifacts
make fmt           # Format code

# Direct Go commands
go build -o bin/goexcel cmd/goexcel/main.go
go run cmd/goexcel/main.go legacy partidas.json
go mod tidy
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

## Environment Setup

Required environment variables in `.env`:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=goexcel_user
DB_PASSWORD=your_password
DB_NAME=goexcel_db
DB_SSLMODE=disable
```

## Testing

```bash
make test                    # Run all tests
go test ./...               # Direct test execution  
make lint                   # Run linter (requires golangci-lint)
```