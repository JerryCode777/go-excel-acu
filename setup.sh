#!/bin/bash

# Setup script para GoExcel v2.0
set -e

echo "🏗️  GoExcel v2.0 - Setup Script"
echo "================================"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para imprimir con colores
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Verificar prerrequisitos
echo "🔍 Verificando prerrequisitos..."

# Verificar Go
if ! command -v go &> /dev/null; then
    print_error "Go no está instalado. Por favor instalar Go 1.21 o superior."
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
print_status "Go $GO_VERSION encontrado"

# Verificar PostgreSQL
if ! command -v psql &> /dev/null; then
    print_warning "PostgreSQL no encontrado en PATH"
    print_info "Por favor asegurar que PostgreSQL esté instalado y accesible"
else
    print_status "PostgreSQL encontrado"
fi

# Crear directorios necesarios
echo
echo "📁 Creando directorios..."
mkdir -p bin output temp logs
print_status "Directorios creados"

# Copiar archivo de configuración si no existe
if [ ! -f .env ]; then
    echo
    echo "⚙️  Configurando variables de entorno..."
    cp .env.example .env
    print_status "Archivo .env creado desde .env.example"
    print_warning "Por favor editar .env con tus credenciales de PostgreSQL"
else
    print_info "Archivo .env ya existe"
fi

# Descargar dependencias
echo
echo "📦 Descargando dependencias..."
go mod download
go mod tidy
print_status "Dependencias descargadas"

# Compilar aplicación
echo
echo "🔨 Compilando aplicación..."
go build -ldflags "-s -w" -o bin/goexcel cmd/goexcel/main.go
print_status "Aplicación compilada: bin/goexcel"

# Verificar que el binario funcione
echo
echo "🧪 Verificando instalación..."
if ./bin/goexcel > /dev/null 2>&1; then
    print_status "Aplicación funciona correctamente"
else
    print_warning "La aplicación puede requerir configuración de base de datos"
fi

echo
echo "🎉 Setup completado!"
echo
print_info "Próximos pasos:"
echo "1. Configurar PostgreSQL y editar .env con tus credenciales"
echo "2. Crear base de datos: createdb goexcel_db"
echo "3. Ejecutar: ./bin/goexcel (para ver opciones)"
echo "4. Migrar datos: ./bin/goexcel migrate partidas.json"
echo
print_info "Comandos útiles:"
echo "  make help           - Ver comandos disponibles"
echo "  ./bin/goexcel help  - Ver ayuda de la aplicación"
echo "  make run-legacy     - Ejecutar en modo legacy (sin BD)"
echo

# Verificar si Make está disponible
if command -v make &> /dev/null; then
    print_status "Make disponible - puedes usar comandos 'make'"
else
    print_warning "Make no disponible - usa comandos go directamente"
fi

echo "================================"
echo "🚀 GoExcel v2.0 listo para usar!"