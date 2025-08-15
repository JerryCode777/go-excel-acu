# Makefile para GoExcel

# Variables
BINARY_NAME=goexcel
MAIN_PACKAGE=./cmd/goexcel
BUILD_DIR=./bin

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build build-server build-all clean test deps fmt help run-legacy run-migrate run-generate

all: clean deps fmt test build

# Compilar la aplicación
build:
	@echo "🔨 Compilando $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ Compilación completada: $(BUILD_DIR)/$(BINARY_NAME)"

# Compilar el servidor
build-server:
	@echo "🔨 Compilando servidor..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/server ./cmd/server
	@echo "✅ Compilación completada: $(BUILD_DIR)/server"

# Compilar ambos binarios
build-all: build build-server

# Limpiar archivos generados
clean:
	@echo "🧹 Limpiando archivos generados..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f *.xlsx
	@echo "✅ Limpieza completada"

# Ejecutar tests
test:
	@echo "🧪 Ejecutando tests..."
	$(GOTEST) -v ./...

# Descargar dependencias
deps:
	@echo "📦 Descargando dependencias..."
	$(GOMOD) download
	$(GOMOD) tidy

# Formatear código
fmt:
	@echo "🎨 Formateando código..."
	$(GOFMT) ./...

# Ejecutar en modo legacy
run-legacy:
	@echo "🔄 Ejecutando en modo legacy..."
	$(GOCMD) run $(MAIN_PACKAGE) legacy partidas.json

# Ejecutar migración desde JSON
run-migrate:
	@echo "📊 Ejecutando migración desde JSON..."
	$(GOCMD) run $(MAIN_PACKAGE) migrate partidas.json

# Generar Excel desde BD (requiere proyecto_id)
run-generate:
	@echo "📋 Listando proyectos disponibles..."
	$(GOCMD) run $(MAIN_PACKAGE)

# Instalar herramientas de desarrollo
dev-tools:
	@echo "🛠️ Instalando herramientas de desarrollo..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint del código
lint:
	@echo "🔍 Ejecutando linter..."
	golangci-lint run

# Crear directorio de salida
setup-dirs:
	@echo "📁 Creando directorios necesarios..."
	@mkdir -p ./output ./temp ./logs

# Ejecutar servidor de desarrollo (para futuro API)
dev-server:
	@echo "🚀 Iniciando servidor de desarrollo..."
	$(GOCMD) run $(MAIN_PACKAGE) server

# Backup de base de datos (requiere PostgreSQL configurado)
backup-db:
	@echo "💾 Creando backup de base de datos..."
	@pg_dump $(DB_NAME) > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Restaurar backup de base de datos
restore-db:
	@echo "📥 Restaurando base de datos desde backup..."
	@psql $(DB_NAME) < $(BACKUP_FILE)

# Versión del binario
version:
	@echo "📄 GoExcel version 2.0.0-alpha"

# Ayuda
help:
	@echo "🏗️  GoExcel - Makefile Commands"
	@echo ""
	@echo "Comandos principales:"
	@echo "  make build         - Compilar la aplicación"
	@echo "  make run-legacy    - Ejecutar modo legacy (sin BD)"
	@echo "  make run-migrate   - Migrar datos desde JSON"
	@echo "  make run-generate  - Listar proyectos disponibles"
	@echo ""
	@echo "Desarrollo:"
	@echo "  make clean         - Limpiar archivos generados"
	@echo "  make test          - Ejecutar tests"
	@echo "  make deps          - Descargar dependencias"
	@echo "  make fmt           - Formatear código"
	@echo "  make lint          - Ejecutar linter"
	@echo ""
	@echo "Configuración:"
	@echo "  make setup-dirs    - Crear directorios necesarios"
	@echo "  make dev-tools     - Instalar herramientas de desarrollo"