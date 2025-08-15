# 🧪 Guía de Testing - GoExcel API

## 🚀 Cómo probar los endpoints

### 1. Iniciar el servidor
```bash
# Desde la raíz del proyecto
./bin/server
```

El servidor estará disponible en: `http://localhost:8080`

### 2. Probar endpoints básicos

#### Health Check
```bash
curl http://localhost:8080/api/v1/health
```

#### Listar proyectos
```bash
curl http://localhost:8080/api/v1/projects
```

#### Validar ACU
```bash
curl -X POST http://localhost:8080/api/v1/validate-acu \
  -H "Content-Type: application/json" \
  -d '{
    "acu_content": "@partida{test, codigo=\"01.01.01\", descripcion=\"TEST\"}"
  }'
```

#### Crear proyecto desde JSON
```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "proyecto": {
      "nombre": "Proyecto Test API",
      "descripcion": "Creado desde API REST",
      "moneda": "PEN"
    },
    "partidas": [
      {
        "codigo": "01.01.01",
        "descripcion": "EXCAVACIÓN MANUAL TEST",
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
  }'
```

### 3. Verificar en base de datos
```bash
# Verificar que el proyecto se creó
./bin/goexcel

# O consultar directamente PostgreSQL
psql postgresql://postgres:postgres@localhost:5432/goexcel_db \
  -c "SELECT nombre, created_at FROM proyectos ORDER BY created_at DESC LIMIT 5;"
```

## 🔗 Para frontend React

### Estructura de request esperada
```javascript
const projectData = {
  proyecto: {
    nombre: "Mi Proyecto",
    descripcion: "Descripción del proyecto",
    moneda: "PEN"
  },
  partidas: [
    {
      codigo: "01.01.01",
      descripcion: "PARTIDA DE EJEMPLO",
      unidad: "m3",
      rendimiento: 8.0,
      mano_obra: [
        {
          codigo: "470101",
          descripcion: "OPERARIO",
          unidad: "hh",
          cantidad: 1.0,
          precio: 25.00,
          cuadrilla: 1.0  // opcional
        }
      ],
      materiales: [...],
      equipos: [...],
      subcontratos: [...]
    }
  ]
};

// Enviar al endpoint
fetch('http://localhost:8080/api/v1/projects', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(projectData)
});
```

### Response esperada
```javascript
{
  "success": true,
  "message": "Proyecto creado exitosamente",
  "project_id": "uuid-string",
  "project": {
    "id": "uuid-string",
    "nombre": "Mi Proyecto",
    "descripcion": "Descripción del proyecto",
    "moneda": "PEN",
    "created_at": "2024-01-15 10:30:00",
    "updated_at": "2024-01-15 10:30:00"
  }
}
```

## 📝 Testing con Postman

### Colección de requests
```json
{
  "info": {
    "name": "GoExcel API",
    "description": "Testing GoExcel endpoints"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "url": "http://localhost:8080/api/v1/health"
      }
    },
    {
      "name": "Get Projects",
      "request": {
        "method": "GET",
        "url": "http://localhost:8080/api/v1/projects"
      }
    },
    {
      "name": "Create Project",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/v1/projects",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\\"proyecto\\": {...}, \\"partidas\\": [...]}"
        }
      }
    }
  ]
}
```

## ⚠️ Notas importantes

1. **CORS**: Configurado para `http://localhost:3000` (React dev server)
2. **Base de datos**: Debe estar corriendo PostgreSQL
3. **Validación**: Los campos `codigo`, `descripcion`, `unidad` y `rendimiento` son obligatorios
4. **Cuadrilla**: Solo aplica para recursos de mano de obra
5. **Precios**: Deben ser números ≥ 0

## 🐛 Debugging

Si algo no funciona:

1. **Verificar logs del servidor**: Buscar mensajes de error
2. **Comprobar BD**: `./bin/goexcel` para ver proyectos
3. **Validar JSON**: Usar herramientas de formato JSON
4. **CORS**: Verificar origen de requests desde React