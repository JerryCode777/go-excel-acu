# API Reference

## Base URL
```
http://localhost:8080/api/v1
```

## 🔍 Health Check

### GET /health
Verifica el estado del servidor y la base de datos.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "2.0.0",
  "database": "connected"
}
```

## 🏗️ Projects

### GET /projects
Obtiene la lista de todos los proyectos.

**Response:**
```json
{
  "success": true,
  "projects": [
    {
      "id": "uuid-string",
      "nombre": "Proyecto Edificación",
      "descripcion": "Descripción del proyecto",
      "moneda": "PEN",
      "created_at": "2024-01-15 10:30:00",
      "updated_at": "2024-01-15 10:30:00"
    }
  ],
  "total": 1
}
```

### POST /projects
Crea un nuevo proyecto desde datos ACU.

**Request:**
```json
{
  "proyecto": {
    "nombre": "Mi Proyecto de Construcción",
    "descripcion": "Descripción detallada del proyecto",
    "moneda": "PEN"
  },
  "partidas": [
    {
      "codigo": "01.01.01",
      "descripcion": "EXCAVACIÓN MANUAL EN TERRENO NORMAL",
      "unidad": "m3",
      "rendimiento": 8.0,
      "mano_obra": [
        {
          "codigo": "470101",
          "descripcion": "OPERARIO",
          "unidad": "hh",
          "cantidad": 1.0,
          "precio": 25.00,
          "cuadrilla": 1.0
        }
      ],
      "materiales": [
        {
          "codigo": "020101",
          "descripcion": "AGUA",
          "unidad": "m3",
          "cantidad": 0.2,
          "precio": 8.00
        }
      ],
      "equipos": [],
      "subcontratos": []
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Proyecto creado exitosamente",
  "project_id": "uuid-string",
  "project": {
    "id": "uuid-string",
    "nombre": "Mi Proyecto de Construcción",
    "descripcion": "Descripción detallada del proyecto",
    "moneda": "PEN",
    "created_at": "2024-01-15 10:30:00",
    "updated_at": "2024-01-15 10:30:00"
  }
}
```

### GET /projects/{id}
Obtiene un proyecto específico con sus partidas.

**Response:**
```json
{
  "success": true,
  "project": {
    "id": "uuid-string",
    "nombre": "Mi Proyecto",
    "descripcion": "Descripción",
    "moneda": "PEN"
  },
  "partidas": [
    {
      "id": "uuid-string",
      "codigo": "01.01.01",
      "descripcion": "EXCAVACIÓN MANUAL",
      "unidad": "m3",
      "rendimiento": 8.0,
      "costo_total": 125.50,
      "mano_obra": [
        {
          "id": "uuid-string",
          "codigo": "470101",
          "descripcion": "OPERARIO",
          "unidad": "hh",
          "cantidad": 1.0,
          "precio": 25.00,
          "cuadrilla": 1.0,
          "parcial": 25.00
        }
      ],
      "materiales": [],
      "equipos": [],
      "subcontratos": []
    }
  ],
  "stats": {
    "total_partidas": 15,
    "total_recursos": 45,
    "costo_total": 15000.00,
    "costo_mano_obra": 5000.00,
    "costo_materiales": 7000.00,
    "costo_equipos": 2000.00,
    "costo_subcontratos": 1000.00
  }
}
```

### PUT /projects/{id}
Actualiza un proyecto existente.

**Request:** (Mismo formato que POST /projects)

**Response:**
```json
{
  "success": true,
  "message": "Proyecto actualizado exitosamente",
  "project_id": "uuid-string"
}
```

### DELETE /projects/{id}
Elimina un proyecto.

**Response:**
```json
{
  "success": true,
  "message": "Proyecto eliminado exitosamente"
}
```

### GET /projects/{id}/export
Exporta un proyecto en diferentes formatos.

**Query Parameters:**
- `format`: excel | acu | json (default: excel)

**Examples:**
- `/projects/uuid/export?format=excel` → Archivo Excel
- `/projects/uuid/export?format=acu` → Archivo .acu
- `/projects/uuid/export?format=json` → JSON completo

## 🔍 Validation

### POST /validate-acu
Valida la sintaxis de código .acu.

**Request:**
```json
{
  "acu_content": "@partida{excavacion,\n  codigo = \"01.01.01\",\n  descripcion = \"EXCAVACIÓN MANUAL\"\n}"
}
```

**Response:**
```json
{
  "valid": true,
  "message": "ACU válido"
}
```

**Error Response:**
```json
{
  "valid": false,
  "message": "Error de sintaxis: campo 'unidad' requerido"
}
```

## ❌ Error Responses

Todos los endpoints pueden retornar errores en este formato:

```json
{
  "success": false,
  "error": "Descripción del error",
  "code": "ERROR_CODE"
}
```

### Códigos de estado HTTP
- `200`: Éxito
- `400`: Bad Request (datos inválidos)
- `404`: Not Found (recurso no encontrado)
- `500`: Internal Server Error (error del servidor)

## 📝 Notas de implementación

### Validación de datos
- Todos los campos marcados como `required` son obligatorios
- Los valores numéricos deben ser ≥ 0
- Los códigos deben ser únicos dentro del proyecto

### Limitaciones actuales
- `GET /projects/{id}`: Retorna estructura básica (en desarrollo)
- `PUT /projects/{id}`: Funcionalidad básica (en desarrollo)
- `DELETE /projects/{id}`: Funcionalidad básica (en desarrollo)
- Exportación: Solo placeholders (en desarrollo)

### Próximas funcionalidades
- Filtrado y paginación en listados
- Búsqueda de proyectos
- Estadísticas avanzadas
- Bulk operations