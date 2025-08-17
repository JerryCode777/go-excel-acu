#!/bin/bash

# Script para aplicar migración de metrados
set -e

echo "🏗️  Aplicando migración de metrados"
echo "=================================="

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Cargar variables de entorno
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    print_status "Variables de entorno cargadas desde .env"
else
    print_error "Archivo .env no encontrado"
    exit 1
fi

# Verificar que las variables estén configuradas
if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ]; then
    print_error "Variables de base de datos no configuradas en .env"
    exit 1
fi

# Construir string de conexión
DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

print_info "Conectando a: ${DB_HOST}:${DB_PORT}/${DB_NAME}"

# Verificar conexión
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
    print_error "No se puede conectar a la base de datos"
    print_info "Verificar que PostgreSQL esté corriendo y las credenciales sean correctas"
    exit 1
fi

print_status "Conexión a base de datos exitosa"

# Aplicar migración de metrados
echo
echo "📊 Aplicando migración de metrados..."

PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f database/metrados_migration.sql

if [ $? -eq 0 ]; then
    print_status "Migración de metrados aplicada exitosamente"
else
    print_error "Error aplicando migración de metrados"
    exit 1
fi

# Verificar que las tablas se crearon correctamente
echo
echo "🔍 Verificando estructura de base de datos..."

TABLES_CHECK=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT EXISTS (
    SELECT FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_name = 'metrados_partidas'
);")

if [[ "$TABLES_CHECK" =~ "t" ]]; then
    print_status "Tabla metrados_partidas creada correctamente"
else
    print_error "Tabla metrados_partidas no fue creada"
    exit 1
fi

# Verificar vista
VIEW_CHECK=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT EXISTS (
    SELECT FROM information_schema.views 
    WHERE table_schema = 'public' 
    AND table_name = 'vista_metrados_completos'
);")

if [[ "$VIEW_CHECK" =~ "t" ]]; then
    print_status "Vista vista_metrados_completos creada correctamente"
else
    print_warning "Vista vista_metrados_completos no fue creada"
fi

# Mostrar funciones creadas
echo
print_info "Funciones PostgreSQL disponibles:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT proname as function_name 
FROM pg_proc 
WHERE proname IN ('calcular_costo_total_proyecto', 'obtener_resumen_proyecto')
ORDER BY proname;"

echo
echo "🎉 Migración de metrados completada!"
echo
print_info "Nuevas funcionalidades disponibles:"
echo "  • Tabla: metrados_partidas - Almacena metrados por proyecto/partida"
echo "  • Vista: vista_metrados_completos - Vista completa con información de partidas"
echo "  • Función: calcular_costo_total_proyecto() - Calcula costo total con metrados"
echo "  • Función: obtener_resumen_proyecto() - Resumen financiero del proyecto"
echo
print_info "API endpoints disponibles:"
echo "  • GET    /api/v1/projects/{id}/metrados"
echo "  • POST   /api/v1/projects/{id}/metrados"
echo "  • PUT    /api/v1/projects/{id}/metrados/batch"
echo "  • DELETE /api/v1/projects/{id}/metrados/{codigo}"
echo "  • GET    /api/v1/projects/{id}/resumen"
echo "  • GET    /api/v1/projects/{id}/costo-total"
echo
echo "=================================="
echo "🚀 Sistema de metrados listo para usar!"