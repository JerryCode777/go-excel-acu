# Formato .acu - Especificación

## 📖 Introducción

El formato .acu es una sintaxis estilo LaTeX diseñada específicamente para definir Análisis de Costos Unitarios (ACUs) de manera legible, versionable y compatible con IA.

## 🎯 Filosofía

- **Human-readable**: Fácil de leer y editar manualmente
- **IA-friendly**: Simple para que IAs generen desde prompts
- **Versionable**: Compatible con Git y sistemas de control de versiones
- **Estándar**: Formato consistente para toda la industria

## 📝 Sintaxis básica

### Estructura de archivo
```acu
@proyecto{id_proyecto,
  campo = "valor",
  campo2 = valor_numerico
}

@partida{id_partida,
  campo = "valor",
  recursos = {
    {campo = "valor", campo2 = numero},
    {campo = "valor", campo2 = numero}
  }
}
```

### Reglas de sintaxis
- **Bloques**: Definidos con `@tipo{id, contenido}`
- **Campos**: Formato `campo = valor`
- **Strings**: Entre comillas dobles `"texto"`
- **Números**: Sin comillas `123.45`
- **Arrays**: Entre llaves `{elemento1, elemento2}`
- **Comentarios**: No implementados aún

## 🏗️ Definición de proyecto

```acu
@proyecto{id_proyecto,
  nombre = "Nombre del Proyecto",
  descripcion = "Descripción detallada del proyecto",
  moneda = "PEN"
}
```

### Campos del proyecto
| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `nombre` | String | ✅ | Nombre del proyecto |
| `descripcion` | String | ❌ | Descripción del proyecto |
| `moneda` | String | ❌ | Código de moneda (default: "PEN") |

## 📋 Definición de partidas

```acu
@partida{id_partida,
  codigo = "01.01.01",
  descripcion = "EXCAVACIÓN MANUAL EN TERRENO NORMAL",
  unidad = "m3",
  rendimiento = 8.00,
  
  mano_obra = {
    {codigo = "470101", desc = "OPERARIO", unidad = "hh", cantidad = 1.0, precio = 25.00, cuadrilla = 1.0},
    {codigo = "470102", desc = "OFICIAL", unidad = "hh", cantidad = 0.5, precio = 22.41}
  },
  
  materiales = {
    {codigo = "020101", desc = "AGUA", unidad = "m3", cantidad = 0.2, precio = 8.00}
  },
  
  equipos = {
    {codigo = "480101", desc = "COMPACTADORA", unidad = "hm", cantidad = 0.5, precio = 45.00}
  },
  
  subcontratos = {
    {codigo = "SUB001", desc = "FLETE TERRESTRE", unidad = "glb", cantidad = 1.0, precio = 500.00}
  }
}
```

### Campos de partida
| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `codigo` | String | ✅ | Código único de la partida |
| `descripcion` | String | ✅ | Descripción de la partida |
| `unidad` | String | ✅ | Unidad de medida |
| `rendimiento` | Número | ✅ | Rendimiento en unidades/día |
| `mano_obra` | Array | ❌ | Recursos de mano de obra |
| `materiales` | Array | ❌ | Recursos de materiales |
| `equipos` | Array | ❌ | Recursos de equipos |
| `subcontratos` | Array | ❌ | Recursos de subcontratos |

### Campos de recurso
| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `codigo` | String | ✅ | Código único del recurso |
| `desc` | String | ✅ | Descripción del recurso |
| `unidad` | String | ✅ | Unidad de medida |
| `cantidad` | Número | ✅ | Cantidad utilizada |
| `precio` | Número | ✅ | Precio unitario |
| `cuadrilla` | Número | ❌ | Factor de cuadrilla (solo mano de obra) |

## 📚 Ejemplos completos

### Ejemplo 1: Partida simple
```acu
@proyecto{vivienda_basica,
  nombre = "Vivienda Unifamiliar",
  descripcion = "Construcción de vivienda de 120m2",
  moneda = "PEN"
}

@partida{excavacion,
  codigo = "01.01.01",
  descripcion = "EXCAVACIÓN MANUAL",
  unidad = "m3",
  rendimiento = 8.0,
  
  mano_obra = {
    {codigo = "470101", desc = "OPERARIO", unidad = "hh", cantidad = 1.0, precio = 25.00}
  }
}
```

### Ejemplo 2: Partida compleja
```acu
@partida{concreto_armado,
  codigo = "03.01.01",
  descripcion = "CONCRETO ARMADO F'C=210 KG/CM2",
  unidad = "m3",
  rendimiento = 10.0,
  
  mano_obra = {
    {codigo = "470101", desc = "OPERARIO", unidad = "hh", cantidad = 1.5, precio = 25.00, cuadrilla = 1.0},
    {codigo = "470102", desc = "OFICIAL", unidad = "hh", cantidad = 1.0, precio = 22.41},
    {codigo = "470103", desc = "PEÓN", unidad = "hh", cantidad = 8.0, precio = 19.68}
  },
  
  materiales = {
    {codigo = "210101", desc = "CEMENTO PORTLAND TIPO I", unidad = "bol", cantidad = 9.73, precio = 23.39},
    {codigo = "210201", desc = "ARENA GRUESA", unidad = "m3", cantidad = 0.54, precio = 45.00},
    {codigo = "210301", desc = "PIEDRA CHANCADA 1/2\"", unidad = "m3", cantidad = 0.81, precio = 55.00},
    {codigo = "020101", desc = "AGUA", unidad = "m3", cantidad = 0.18, precio = 8.00}
  },
  
  equipos = {
    {codigo = "490101", desc = "MEZCLADORA DE CONCRETO", unidad = "hm", cantidad = 0.8, precio = 35.00},
    {codigo = "490201", desc = "VIBRADOR DE CONCRETO", unidad = "hm", cantidad = 0.5, precio = 15.00}
  }
}
```

### Ejemplo 3: Solo subcontratos
```acu
@partida{transporte,
  codigo = "12.01.01",
  descripcion = "TRANSPORTE DE MATERIALES",
  unidad = "glb",
  rendimiento = 1.0,
  
  subcontratos = {
    {codigo = "SUB001", desc = "FLETE TERRESTRE LOCAL", unidad = "glb", cantidad = 1.0, precio = 2500.00}
  }
}
```

## 🔍 Validación

### Reglas de validación
1. **Proyecto**: Debe existir exactamente uno
2. **Partidas**: Al menos una partida requerida
3. **Códigos únicos**: Códigos de partidas deben ser únicos
4. **Valores numéricos**: Deben ser ≥ 0
5. **Campos requeridos**: No pueden estar vacíos

### Errores comunes
```acu
# ❌ Error: Falta campo requerido
@partida{bad_partida,
  codigo = "01.01.01"
  # Falta descripcion, unidad, rendimiento
}

# ❌ Error: Sintaxis incorrecta
@partida{bad_syntax,
  codigo = 01.01.01  # Falta comillas
}

# ❌ Error: Valor negativo
@partida{bad_values,
  codigo = "01.01.01",
  rendimiento = -5.0  # No puede ser negativo
}
```

## 🛠️ Herramientas

### Validación
```bash
# Validar sintaxis
./goexcel validate-acu proyecto.acu

# Convertir a JSON
./goexcel acu-to-json proyecto.acu output.json

# Importar directamente
./goexcel import-acu proyecto.acu
```

### Editor recomendado
Para una mejor experiencia de edición:
- **VS Code**: Con extensión de syntax highlighting
- **Vim**: Con syntax highlighting personalizado
- **Sublime Text**: Con custom syntax

## 🚀 Casos de uso

### 1. Generación por IA
```
Usuario: "Crea un ACU para excavación manual"
IA: Genera código .acu → Usuario copia/pega → Sistema procesa
```

### 2. OCR de imágenes
```
Imagen ACU → OCR → IA procesa → Código .acu → Validación → Sistema
```

### 3. Templates
```acu
# Template para excavación
@partida{excavacion_template,
  codigo = "XX.XX.XX",
  descripcion = "EXCAVACIÓN [TIPO] EN [MATERIAL]",
  unidad = "m3",
  rendimiento = 0.0,
  
  mano_obra = {
    {codigo = "470101", desc = "OPERARIO", unidad = "hh", cantidad = 0.0, precio = 0.0}
  }
}
```

## 📈 Roadmap

### Próximas funcionalidades
- **Comentarios**: `# Esto es un comentario`
- **Variables**: `@var{precio_operario = 25.00}`
- **Includes**: `@include{partidas_comunes.acu}`
- **Macros**: `@macro{operario_basico = {...}}`
- **Validación avanzada**: Rangos de precios, unidades válidas

### Extensiones futuras
- **Metadata**: Fechas, versiones, autores
- **Fórmulas**: Cálculos automáticos
- **Dependencias**: Relaciones entre partidas
- **Localization**: Soporte multi-idioma