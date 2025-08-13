# GoExcel - Generador de ACUs 🏗️

Sistema para procesamiento de Análisis de Costos Unitarios (ACU) que lee datos de partidas con sus recursos (mano de obra, materiales, equipos y subcontratos) y genera reportes Excel profesionales.

## 🚀 Versión 2.0 - Con PostgreSQL

Esta versión incluye:
- ✅ Base de datos PostgreSQL para almacenar proyectos y partidas
- ✅ API REST para integración con frontend React
- ✅ Migración desde archivos JSON existentes
- ✅ Modo legacy para compatibilidad con versión anterior
- ✅ Estructura de archivos escalable

## 📋 Características

- **Procesamiento de ACUs**: Análisis completo de costos unitarios por partida
- **Múltiples tipos de recursos**: Mano de obra, materiales, equipos y subcontratos
- **Generación Excel**: Reportes profesionales con formato y estilos
- **Base de datos**: PostgreSQL para persistencia de datos
- **Migración de datos**: Importación desde archivos JSON existentes
- **API REST**: Endpoints para integración con aplicaciones web

## 🛠️ Instalación

### Prerrequisitos

- Go 1.21 o superior
- PostgreSQL 12 o superior
- Make (opcional, para usar Makefile)

### 1. Clonar repositorio

```bash
git clone https://github.com/jerryandersonh/goexcel.git
cd goexcel
```

### 2. Configurar base de datos

```bash
# Crear base de datos PostgreSQL
createdb goexcel_db

# Crear usuario (opcional)
psql -c "CREATE USER goexcel_user WITH PASSWORD 'your_password';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE goexcel_db TO goexcel_user;"
```

### 3. Configurar variables de entorno

```bash
# Copiar archivo de configuración
cp .env.example .env

# Editar .env con tus credenciales
nano .env
```

### 4. Instalar dependencias

```bash
go mod download
```

### 5. Compilar aplicación

```bash
make build
# o
go build -o bin/goexcel cmd/goexcel/main.go
```

## 🎯 Uso

### Modo Base de Datos

#### 1. Migrar datos desde JSON
```bash
./bin/goexcel migrate partidas.json
```

#### 2. Listar proyectos disponibles
```bash
./bin/goexcel
```

#### 3. Generar Excel desde base de datos
```bash
./bin/goexcel generate <proyecto_id>
```

### Modo Legacy (Compatible con v1.0)

```bash
# Usar archivo JSON directamente
./bin/goexcel legacy partidas.json

# Con archivo por defecto
./bin/goexcel legacy
```

### Con Makefile

```bash
# Ejecutar en modo legacy
make run-legacy

# Migrar desde JSON
make run-migrate

# Listar proyectos
make run-generate
```

## 📊 Estructura de Datos

### JSON de entrada (formato original)
```json
[
  {
    "codigo": "01.01.01.01",
    "descripcion": "CARTEL DE IDENTIFICACIÓN DE OBRA",
    "unidad": "und",
    "rendimiento": 1.0000,
    "mano_obra": [
      {
        "codigo": "010101002",
        "descripcion": "CAPATAZ",
        "unidad": "hh",
        "cuadrilla": 1.0000,
        "cantidad": 8.0000,
        "precio": 29.08
      }
    ],
    "materiales": [...],
    "equipos": [...],
    "subcontratos": [...]
  }
]
```

### Base de datos PostgreSQL

El sistema utiliza las siguientes tablas principales:

- **proyectos**: Información de proyectos
- **partidas**: Partidas de cada proyecto
- **recursos**: Catálogo de recursos (MO, materiales, equipos, subcontratos)
- **partida_recursos**: Relación many-to-many entre partidas y recursos
- **tipos_recurso**: Clasificación de recursos
- **analisis_historicos**: Historial de análisis generados

## 📁 Estructura del Proyecto

```
goexcel/
├── cmd/goexcel/           # Aplicación principal
├── config/                # Configuración
├── database/              # Scripts SQL y migraciones
├── internal/
│   ├── database/          # Conexión y repositorios
│   ├── models/            # Estructuras de datos
│   ├── services/          # Lógica de negocio
│   └── handlers/          # Handlers HTTP (futuro)
├── pkg/
│   ├── excel/             # Utilidades Excel
│   └── utils/             # Utilidades generales
├── output/                # Archivos Excel generados
├── temp/                  # Archivos temporales
├── .env                   # Variables de entorno
├── Makefile              # Comandos de construcción
└── README.md
```

## 🧪 Testing

```bash
# Ejecutar tests
make test

# Con coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔧 Desarrollo

### Comandos Make disponibles

```bash
make help          # Ver todos los comandos
make build         # Compilar aplicación
make clean         # Limpiar archivos generados
make test          # Ejecutar tests
make deps          # Descargar dependencias
make fmt           # Formatear código
make lint          # Ejecutar linter
make setup-dirs    # Crear directorios necesarios
```

### Base de datos

```bash
# Ejecutar migraciones manualmente
psql goexcel_db < database/schema.sql

# Migrar datos desde JSON
psql goexcel_db -f database/migrate_json.sql
```

## 🌐 API REST (En desarrollo)

La aplicación incluirá endpoints REST para:

- `GET /api/proyectos` - Listar proyectos
- `POST /api/proyectos` - Crear proyecto
- `GET /api/proyectos/{id}/partidas` - Listar partidas de un proyecto
- `POST /api/proyectos/{id}/analisis` - Generar análisis Excel

## 🔄 Migración desde v1.0

Si tienes datos en formato JSON de la versión anterior:

```bash
# 1. Migrar datos a PostgreSQL
./bin/goexcel migrate tu_archivo.json

# 2. Verificar migración
./bin/goexcel

# 3. Generar Excel desde BD
./bin/goexcel generate <proyecto_id>
```

## 📝 Variables de Entorno

```bash
# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=goexcel_user
DB_PASSWORD=your_password
DB_NAME=goexcel_db
DB_SSLMODE=disable

# Servidor (futuro)
SERVER_PORT=8080
SERVER_HOST=localhost

# Archivos
EXCEL_OUTPUT_DIR=./output
TEMP_DIR=./temp
```

## 🤝 Contribuir

1. Fork del repositorio
2. Crear rama de feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit de cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver archivo `LICENSE` para más detalles.

## 📞 Soporte

Para reportar bugs o solicitar funcionalidades, crear un issue en el repositorio de GitHub.

---

**GoExcel v2.0** - Sistema de Análisis de Costos Unitarios con PostgreSQL