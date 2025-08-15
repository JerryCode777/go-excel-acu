# Arquitectura del Sistema

## 🏗️ Visión general

GoExcel es un sistema distribuido de microservicios diseñado para la gestión integral de Análisis de Costos Unitarios (ACUs) en construcción.

## 📐 Diagrama de arquitectura

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  React Frontend │    │ Python AI Service│    │   PostgreSQL    │
│                 │    │                 │    │    Database     │
│ • ACU Editor    │    │ • OCR Processing │    │                 │
│ • Project Mgmt  │    │ • AI Generation │    │ • Projects      │
│ • Visualization │    │ • Optimization  │    │ • Partidas      │
│                 │    │ • Specs Gen     │    │ • Resources     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │ HTTP/JSON            │ HTTP/JSON            │ SQL
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Go Backend API                               │
│                                                                 │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │  Handlers   │ │  Services   │ │ Repository  │ │  Database   │ │
│ │             │ │             │ │             │ │             │ │
│ │ • HTTP API  │ │ • Business  │ │ • Data      │ │ • Migrations│ │
│ │ • Validation│ │   Logic     │ │   Access    │ │ • Schema    │ │
│ │ • Responses │ │ • ACU Parse │ │ • Queries   │ │ • Indexes   │ │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
│                                                                 │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │    Models   │ │    Config   │ │   Legacy    │ │     CLI     │ │
│ │             │ │             │ │             │ │             │ │
│ │ • API DTOs  │ │ • Database  │ │ • Excel Gen │ │ • Commands  │ │
│ │ • DB Models │ │ • Server    │ │ • JSON      │ │ • Tools     │ │
│ │ • ACU Types │ │ • CORS      │ │ • Migration │ │ • Scripts   │ │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 🔧 Componentes principales

### 1. Frontend (React)
**Responsabilidades:**
- Interface de usuario para crear/editar ACUs
- Editor de texto para formato .acu
- Visualización de proyectos y estadísticas
- Parser JavaScript para validación local

**Tecnologías:**
- React 18+
- TypeScript
- Axios para HTTP
- CSS-in-JS o Styled Components

### 2. Backend API (Go)
**Responsabilidades:**
- API REST para operaciones CRUD
- Procesamiento de archivos .acu
- Normalización de datos
- Migración a base de datos
- Generación de reportes

**Tecnologías:**
- Go 1.21+
- Gorilla Mux (routing)
- PostgreSQL driver
- UUID generation

### 3. AI Service (Python) - Futuro
**Responsabilidades:**
- OCR de imágenes → .acu
- Generación de especificaciones técnicas
- Optimización de ACUs con IA
- Estimación inteligente de costos

**Tecnologías:**
- FastAPI
- Tesseract OCR
- OpenAI/Claude APIs
- PIL/OpenCV

### 4. Base de datos (PostgreSQL)
**Responsabilidades:**
- Almacenamiento persistente
- Integridad referencial
- Índices optimizados
- Triggers para cálculos

## 🔄 Flujos de datos

### Flujo 1: Creación de proyecto desde React
```
React → Parse .acu → JSON → Go API → Normalize → PostgreSQL
  ↓                                                   ↓
Project Created ←─────────── Response ←──────────── Stored
```

### Flujo 2: OCR de imagen (futuro)
```
React → Upload Image → Python AI → .acu Code → React → Go API
                         ↓
                    OCR + AI Processing
```

### Flujo 3: CLI tradicional
```
.acu file → Go CLI → Parse → Normalize → PostgreSQL
    ↓                                        ↓
JSON file ←─────── Export ←─────────── Query Results
```

## 📁 Estructura de directorios

```
goexcel/
├── cmd/                      # Puntos de entrada
│   ├── goexcel/             # CLI principal
│   └── server/              # Servidor HTTP
├── internal/                # Código interno
│   ├── handlers/            # Controladores HTTP
│   ├── services/            # Lógica de negocio
│   ├── database/            # Acceso a datos
│   ├── models/              # Estructuras de datos
│   ├── server/              # Configuración servidor
│   └── legacy/              # Código legacy
├── config/                  # Configuración
├── docs/                    # Documentación
├── web/                     # Frontend React
├── ai-service/              # Microservicio Python (futuro)
├── database/                # Scripts SQL
└── bin/                     # Binarios compilados
```

## 🌐 Interfaces y contratos

### HTTP API
- **Base URL**: `http://localhost:8080/api/v1`
- **Formato**: JSON
- **Auth**: Ninguna (futuro: JWT)
- **CORS**: Habilitado para desarrollo

### Database Schema
- **Engine**: PostgreSQL 13+
- **Features**: UUIDs, triggers, índices
- **Migration**: Automática en startup

### File Formats
- **Input**: .acu, JSON
- **Output**: Excel, JSON, .acu
- **Validation**: Sintaxis y datos

## 🔒 Seguridad

### Actual
- Validación de entrada básica
- Sanitización de SQL (usando drivers)
- CORS configurado

### Futuro
- Autenticación JWT
- Rate limiting
- Input validation avanzada
- File upload security

## 📈 Escalabilidad

### Horizontal
- Múltiples instancias del Go backend
- Load balancer (Nginx/HAProxy)
- Python AI service independiente

### Vertical
- Connection pooling en DB
- Índices optimizados
- Caching (Redis futuro)

### Performance
- Consultas SQL optimizadas
- Parsing asíncrono
- Streaming para archivos grandes

## 🔍 Monitoreo y observabilidad

### Logging
- Structured logging en Go
- Niveles: DEBUG, INFO, WARN, ERROR
- Formato JSON para producción

### Métricas (futuro)
- Prometheus metrics
- Request duration
- Database query performance
- Error rates

### Health checks
- `/api/v1/health` endpoint
- Database connectivity
- Service dependencies

## 🚀 Deployment

### Desarrollo
```bash
# Backend
go run cmd/server/main.go

# Frontend
npm start

# Database
docker run postgres:15
```

### Producción (futuro)
```yaml
# docker-compose.yml
services:
  backend:
    image: goexcel:latest
    ports: ["8080:8080"]
  
  frontend:
    image: nginx:alpine
    volumes: ["./build:/usr/share/nginx/html"]
  
  ai-service:
    image: goexcel-ai:latest
    ports: ["8001:8001"]
  
  database:
    image: postgres:15
    volumes: ["pgdata:/var/lib/postgresql/data"]
```

## 🔧 Configuración

### Variables de entorno
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=goexcel_db

# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

# AI Service (futuro)
AI_SERVICE_URL=http://localhost:8001
OPENAI_API_KEY=sk-...
```

## 📋 Patrones y principios

### Backend (Go)
- **Repository Pattern**: Separación de datos
- **Service Layer**: Lógica de negocio
- **Dependency Injection**: Testabilidad
- **Clean Architecture**: Separación de responsabilidades

### Frontend (React)
- **Component Composition**: Reutilización
- **Custom Hooks**: Lógica compartida
- **Context API**: Estado global
- **Error Boundaries**: Manejo de errores

### Database
- **Normalization**: Tercera forma normal
- **Foreign Keys**: Integridad referencial
- **Indexes**: Performance queries
- **Triggers**: Cálculos automáticos

## 🔄 Testing strategy

### Unit Tests
- Servicios de negocio
- Parsers y validadores
- Database repositories

### Integration Tests
- HTTP endpoints
- Database operations
- File processing

### E2E Tests (futuro)
- User workflows
- React components
- API interactions

## 🚧 Roadmap técnico

### Fase 1 (Actual)
- ✅ API básica funcional
- ✅ Formato .acu implementado
- ✅ CLI completo
- ✅ Database schema

### Fase 2 (Próxima)
- 🔄 React frontend completo
- 🔄 Python AI service
- 🔄 OCR processing
- 🔄 Excel generation desde DB

### Fase 3 (Futuro)
- Authentication/authorization
- Multi-tenant support
- Real-time collaboration
- Mobile app

## 📊 Métricas de éxito

### Performance
- API response time < 200ms
- Database queries < 50ms
- File processing < 2s

### Reliability
- Uptime > 99.9%
- Error rate < 0.1%
- Data consistency 100%

### Usability
- ACU creation time < 5min
- Learning curve < 1 hour
- User satisfaction > 90%