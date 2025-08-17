package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"goexcel/internal/models"
)

// ACUJerarquicoParser parsea el nuevo formato ACU jerárquico
type ACUJerarquicoParser struct{}

// NewACUJerarquicoParser crea una nueva instancia del parser jerárquico
func NewACUJerarquicoParser() *ACUJerarquicoParser {
	return &ACUJerarquicoParser{}
}

// ParseACUJerarquico parsea contenido ACU en formato jerárquico
func (p *ACUJerarquicoParser) ParseACUJerarquico(content string) (*models.ACUJerarquico, error) {
	// Limpiar contenido
	content = strings.ReplaceAll(content, "\r", "")
	lines := strings.Split(content, "\n")

	result := &models.ACUJerarquico{
		Subpresupuestos: []models.SubpresupuestoData{},
		Titulos:         []models.TituloData{},
		Partidas:        []models.PartidaData{},
	}

	// Stack para mantener la jerarquía de títulos
	tituloStack := make([]string, 10) // hasta 10 niveles
	partidaCounters := make(map[string]int) // key: título código, value: siguiente número de partida

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse @presupuesto{}
		if strings.HasPrefix(line, "@presupuesto{") {
			presupuesto, err := p.parsePresupuesto(lines, &i)
			if err != nil {
				return nil, fmt.Errorf("error parsing presupuesto: %v", err)
			}
			result.Presupuesto = *presupuesto
			continue
		}

		// Parse @subpresupuesto{}
		if strings.HasPrefix(line, "@subpresupuesto{") {
			subpresupuesto, err := p.parseSubpresupuesto(lines, &i)
			if err != nil {
				return nil, fmt.Errorf("error parsing subpresupuesto: %v", err)
			}
			result.Subpresupuestos = append(result.Subpresupuestos, *subpresupuesto)
			continue
		}

		// Parse @titulo{nivel, ...}
		if strings.HasPrefix(line, "@titulo{") {
			titulo, err := p.parseTituloConJerarquia(lines, &i, tituloStack)
			if err != nil {
				return nil, fmt.Errorf("error parsing titulo: %v", err)
			}
			result.Titulos = append(result.Titulos, *titulo)
			continue
		}

		// Parse @partida{}
		if strings.HasPrefix(line, "@partida{") {
			// Construir el código del título actual basado en el stack
			currentTituloCodigo := p.construirCodigoActual(tituloStack)
			
			partida, err := p.parsePartida(lines, &i, currentTituloCodigo, partidaCounters)
			if err != nil {
				return nil, fmt.Errorf("error parsing partida: %v", err)
			}
			result.Partidas = append(result.Partidas, *partida)
			continue
		}
	}

	return result, nil
}

// parsePresupuesto parsea un bloque @presupuesto{}
func (p *ACUJerarquicoParser) parsePresupuesto(lines []string, index *int) (*models.PresupuestoData, error) {
	presupuesto := &models.PresupuestoData{
		Moneda: "PEN", // Default
	}

	// Extraer código del presupuesto de la declaración
	line := lines[*index]
	re := regexp.MustCompile(`@presupuesto\{([^,}]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		presupuesto.Codigo = strings.TrimSpace(matches[1])
	}

	// Parsear contenido del bloque
	for *index+1 < len(lines) {
		*index++
		line := strings.TrimSpace(lines[*index])
		
		if line == "}" {
			break
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"',`)

				switch key {
				case "nombre":
					presupuesto.Nombre = value
				case "cliente":
					presupuesto.Cliente = &value
				case "lugar":
					presupuesto.Lugar = &value
				case "moneda":
					presupuesto.Moneda = value
				}
			}
		}
	}

	return presupuesto, nil
}

// parseSubpresupuesto parsea un bloque @subpresupuesto{}
func (p *ACUJerarquicoParser) parseSubpresupuesto(lines []string, index *int) (*models.SubpresupuestoData, error) {
	subpresupuesto := &models.SubpresupuestoData{}

	// Extraer código del subpresupuesto
	line := lines[*index]
	re := regexp.MustCompile(`@subpresupuesto\{([^,}]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		subpresupuesto.Codigo = strings.TrimSpace(matches[1])
	}

	// Parsear contenido
	for *index+1 < len(lines) {
		*index++
		line := strings.TrimSpace(lines[*index])
		
		if line == "}" {
			break
		}

		if strings.Contains(line, "=") || strings.Contains(line, ":") {
			var parts []string
			if strings.Contains(line, "=") {
				parts = strings.SplitN(line, "=", 2)
			} else {
				parts = strings.SplitN(line, ":", 2)
			}
			
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"',`)

				switch key {
				case "nombre":
					subpresupuesto.Nombre = value
				}
			}
		}
	}

	return subpresupuesto, nil
}

// parseTituloConJerarquia parsea un bloque @titulo{nivel, ...} y maneja la jerarquía
func (p *ACUJerarquicoParser) parseTituloConJerarquia(lines []string, index *int, tituloStack []string) (*models.TituloData, error) {
	titulo := &models.TituloData{}

	// Extraer nivel del título
	line := lines[*index]
	re := regexp.MustCompile(`@titulo\{(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		nivel, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid nivel in titulo: %s", matches[1])
		}
		titulo.Nivel = nivel
	}

	// Parsear contenido para obtener el nombre
	for *index+1 < len(lines) {
		*index++
		line := strings.TrimSpace(lines[*index])
		
		if line == "}" {
			break
		}

		if strings.Contains(line, "=") || strings.Contains(line, ":") {
			var parts []string
			if strings.Contains(line, "=") {
				parts = strings.SplitN(line, "=", 2)
			} else {
				parts = strings.SplitN(line, ":", 2)
			}
			
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"',`)

				switch key {
				case "nombre":
					titulo.Nombre = value
				}
			}
		}
	}

	// Generar código jerárquico automático basado en el stack
	titulo.CodigoCompleto = p.generarCodigoJerarquico(titulo.Nivel, tituloStack)

	return titulo, nil
}

// parsePartida parsea un bloque @partida{} con el formato jerárquico
func (p *ACUJerarquicoParser) parsePartida(lines []string, index *int, tituloCodigo string, partidaCounters map[string]int) (*models.PartidaData, error) {
	partida := &models.PartidaData{
		ManoObra:     []models.RecursoData{},
		Materiales:   []models.RecursoData{},
		Equipos:      []models.RecursoData{},
		Subcontratos: []models.RecursoData{},
	}

	// Generar código automático de partida
	partida.Codigo = p.generarCodigoPartida(tituloCodigo, partidaCounters)

	var currentSection string
	var bracketLevel int

	// Parsear contenido del bloque
	for *index+1 < len(lines) {
		*index++
		line := strings.TrimSpace(lines[*index])
		
		if line == "}" && bracketLevel == 0 {
			break
		}

		// Parsear campos básicos
		if strings.Contains(line, "=") && bracketLevel == 0 {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"',`)

				switch key {
				case "descripcion":
					partida.Descripcion = value
				case "unidad":
					partida.Unidad = value
				case "rendimiento":
					if rendimiento, err := strconv.ParseFloat(value, 64); err == nil {
						partida.Rendimiento = rendimiento
					}
				}
			}
		}

		// Parsear secciones de recursos
		if strings.Contains(line, "= {") {
			sectionMatch := regexp.MustCompile(`(\w+)\s*=\s*\{`).FindStringSubmatch(line)
			if len(sectionMatch) > 1 {
				currentSection = sectionMatch[1]
				bracketLevel = 1
				continue
			}
		}

		// Contar brackets para nivel de anidamiento
		if currentSection != "" {
			bracketLevel += strings.Count(line, "{")
			bracketLevel -= strings.Count(line, "}")

			// Parsear recursos
			if strings.HasPrefix(line, "{") && strings.Contains(line, "codigo =") {
				recurso, err := p.parseRecurso(lines, index, currentSection)
				if err != nil {
					return nil, fmt.Errorf("error parsing recurso: %v", err)
				}

				switch currentSection {
				case "mano_obra":
					partida.ManoObra = append(partida.ManoObra, *recurso)
				case "materiales":
					partida.Materiales = append(partida.Materiales, *recurso)
				case "equipos":
					partida.Equipos = append(partida.Equipos, *recurso)
				case "subcontratos":
					partida.Subcontratos = append(partida.Subcontratos, *recurso)
				}
			}

			// Fin de sección
			if bracketLevel == 0 {
				currentSection = ""
			}
		}
	}

	return partida, nil
}

// parseRecurso parsea un recurso individual
func (p *ACUJerarquicoParser) parseRecurso(lines []string, index *int, seccion string) (*models.RecursoData, error) {
	recurso := &models.RecursoData{}

	line := lines[*index]
	
	// Manejar recursos multilinea
	for !strings.Contains(line, "}") && *index < len(lines)-1 {
		*index++
		line += " " + strings.TrimSpace(lines[*index])
	}

	// Extraer campos del recurso
	resourceContent := strings.Trim(line, "{}")
	fieldMatches := regexp.MustCompile(`(\w+)\s*=\s*([^,}]+)`).FindAllStringSubmatch(resourceContent, -1)
	
	for _, match := range fieldMatches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			value = strings.Trim(value, `"'`)

			switch key {
			case "codigo":
				recurso.Codigo = value
			case "desc":
				recurso.Descripcion = value
			case "unidad":
				recurso.Unidad = value
			case "cantidad":
				if cantidad, err := strconv.ParseFloat(value, 64); err == nil {
					recurso.Cantidad = cantidad
				}
			case "precio":
				if precio, err := strconv.ParseFloat(value, 64); err == nil {
					recurso.Precio = precio
				}
			case "cuadrilla":
				if seccion == "mano_obra" {
					if cuadrilla, err := strconv.ParseFloat(value, 64); err == nil {
						recurso.Cuadrilla = &cuadrilla
					}
				}
			}
		}
	}

	return recurso, nil
}

// generarCodigoJerarquico genera código jerárquico automático basado en el stack de títulos
func (p *ACUJerarquicoParser) generarCodigoJerarquico(nivel int, tituloStack []string) string {
	// Usar contadores simples por nivel almacenados en el stack mismo
	// tituloStack[i] almacena el número actual del nivel i+1
	
	// Limpiar niveles superiores al actual
	for i := nivel; i < len(tituloStack); i++ {
		tituloStack[i] = ""
	}
	
	// Incrementar contador del nivel actual
	if nivel > 0 && nivel <= len(tituloStack) {
		currentNum, _ := strconv.Atoi(tituloStack[nivel-1])
		currentNum++
		tituloStack[nivel-1] = strconv.Itoa(currentNum)
	}
	
	// Construir código jerárquico
	var parts []string
	for i := 0; i < nivel; i++ {
		num := "01" // por defecto
		if tituloStack[i] != "" {
			if n, err := strconv.Atoi(tituloStack[i]); err == nil {
				num = fmt.Sprintf("%02d", n)
			}
		}
		parts = append(parts, num)
	}
	
	return strings.Join(parts, ".")
}

// construirCodigoActual construye el código del título actual basado en el stack
func (p *ACUJerarquicoParser) construirCodigoActual(tituloStack []string) string {
	var parts []string
	
	// Encontrar el último nivel no vacío
	ultimoNivel := 0
	for i := 0; i < len(tituloStack); i++ {
		if tituloStack[i] != "" {
			ultimoNivel = i + 1
		}
	}
	
	if ultimoNivel == 0 {
		return "" // No hay títulos
	}
	
	// Construir código hasta el último nivel
	for i := 0; i < ultimoNivel; i++ {
		num := "01" // por defecto
		if tituloStack[i] != "" {
			if n, err := strconv.Atoi(tituloStack[i]); err == nil {
				num = fmt.Sprintf("%02d", n)
			}
		}
		parts = append(parts, num)
	}
	
	return strings.Join(parts, ".")
}

// generarCodigoPartida genera código automático para partidas
func (p *ACUJerarquicoParser) generarCodigoPartida(tituloCodigo string, partidaCounters map[string]int) string {
	if tituloCodigo == "" {
		// Partida sin título (no recomendado)
		partidaCounters["sin_titulo"]++
		return fmt.Sprintf("%02d", partidaCounters["sin_titulo"])
	}

	partidaCounters[tituloCodigo]++
	return fmt.Sprintf("%s.%02d", tituloCodigo, partidaCounters[tituloCodigo])
}

// ConvertToACUJerarquico convierte estructura de partidas a formato ACU jerárquico
func (p *ACUJerarquicoParser) ConvertToACUJerarquico(data *models.ACUJerarquico) string {
	var acuContent strings.Builder

	// Generar @presupuesto{}
	acuContent.WriteString(fmt.Sprintf("@presupuesto{%s,\n", data.Presupuesto.Codigo))
	acuContent.WriteString(fmt.Sprintf("  nombre = \"%s\",\n", data.Presupuesto.Nombre))
	if data.Presupuesto.Cliente != nil {
		acuContent.WriteString(fmt.Sprintf("  cliente = \"%s\",\n", *data.Presupuesto.Cliente))
	}
	if data.Presupuesto.Lugar != nil {
		acuContent.WriteString(fmt.Sprintf("  lugar = \"%s\",\n", *data.Presupuesto.Lugar))
	}
	acuContent.WriteString(fmt.Sprintf("  moneda = \"%s\"\n", data.Presupuesto.Moneda))
	acuContent.WriteString("}\n\n")

	// Generar @subpresupuesto{}
	for _, sub := range data.Subpresupuestos {
		acuContent.WriteString(fmt.Sprintf("@subpresupuesto{%s,\n", sub.Codigo))
		acuContent.WriteString(fmt.Sprintf("  nombre = \"%s\"\n", sub.Nombre))
		acuContent.WriteString("}\n\n")
	}

	// Generar @titulo{}
	for _, titulo := range data.Titulos {
		acuContent.WriteString(fmt.Sprintf("@titulo{%d,\n", titulo.Nivel))
		acuContent.WriteString(fmt.Sprintf("  nombre = \"%s\"\n", titulo.Nombre))
		acuContent.WriteString("}\n\n")
	}

	// Generar @partida{}
	for _, partida := range data.Partidas {
		acuContent.WriteString(fmt.Sprintf("@partida{%s,\n", strings.ToLower(strings.ReplaceAll(partida.Descripcion, " ", "_"))))
		acuContent.WriteString(fmt.Sprintf("  descripcion = \"%s\",\n", partida.Descripcion))
		acuContent.WriteString(fmt.Sprintf("  unidad = \"%s\",\n", partida.Unidad))
		acuContent.WriteString(fmt.Sprintf("  rendimiento = %.1f,\n", partida.Rendimiento))

		// Agregar secciones de recursos
		if len(partida.ManoObra) > 0 {
			acuContent.WriteString("  \n  mano_obra = {\n")
			for _, recurso := range partida.ManoObra {
				acuContent.WriteString(fmt.Sprintf("    {codigo = \"%s\", desc = \"%s\", unidad = \"%s\", cantidad = %.4f, precio = %.2f",
					recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.Cantidad, recurso.Precio))
				if recurso.Cuadrilla != nil {
					acuContent.WriteString(fmt.Sprintf(", cuadrilla = %.4f", *recurso.Cuadrilla))
				}
				acuContent.WriteString("},\n")
			}
			acuContent.WriteString("  },\n")
		}

		if len(partida.Materiales) > 0 {
			acuContent.WriteString("  \n  materiales = {\n")
			for _, recurso := range partida.Materiales {
				acuContent.WriteString(fmt.Sprintf("    {codigo = \"%s\", desc = \"%s\", unidad = \"%s\", cantidad = %.4f, precio = %.2f},\n",
					recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.Cantidad, recurso.Precio))
			}
			acuContent.WriteString("  },\n")
		}

		if len(partida.Equipos) > 0 {
			acuContent.WriteString("  \n  equipos = {\n")
			for _, recurso := range partida.Equipos {
				acuContent.WriteString(fmt.Sprintf("    {codigo = \"%s\", desc = \"%s\", unidad = \"%s\", cantidad = %.4f, precio = %.2f},\n",
					recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.Cantidad, recurso.Precio))
			}
			acuContent.WriteString("  },\n")
		}

		if len(partida.Subcontratos) > 0 {
			acuContent.WriteString("  \n  subcontratos = {\n")
			for _, recurso := range partida.Subcontratos {
				acuContent.WriteString(fmt.Sprintf("    {codigo = \"%s\", desc = \"%s\", unidad = \"%s\", cantidad = %.4f, precio = %.2f},\n",
					recurso.Codigo, recurso.Descripcion, recurso.Unidad, recurso.Cantidad, recurso.Precio))
			}
			acuContent.WriteString("  },\n")
		}

		acuContent.WriteString("}\n\n")
	}

	return acuContent.String()
}