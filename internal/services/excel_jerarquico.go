package services

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
	"goexcel/config"
	"goexcel/internal/models"
)

type ExcelJerarquicoService struct {
	config *config.Config
}

func NewExcelJerarquicoService(config *config.Config) *ExcelJerarquicoService {
	return &ExcelJerarquicoService{
		config: config,
	}
}

// GenerarExcelJerarquico genera Excel con verdadera estructura jerárquica desde BD
func (s *ExcelJerarquicoService) GenerarExcelJerarquico(proyecto *models.Proyecto, hierarchySvc *HierarchyService) (string, error) {
	f := excelize.NewFile()
	defer f.Close()
	
	// Crear hojas con nombres profesionales
	apuSheet := "Análisis de Precios Unitarios"
	presupuestoSheet := "Presupuesto General"
	
	f.SetSheetName("Sheet1", apuSheet)
	f.NewSheet(presupuestoSheet)
	
	// Obtener jerarquía completa desde BD
	jerarquia, err := hierarchySvc.ObtenerJerarquiaCompleta(proyecto.ID.String())
	if err != nil {
		return "", fmt.Errorf("error obteniendo jerarquía: %v", err)
	}
	
	// Obtener partidas con recursos completos
	partidasConRecursos, err := hierarchySvc.ObtenerPartidasConJerarquia(proyecto.ID.String())
	if err != nil {
		return "", fmt.Errorf("error obteniendo partidas: %v", err)
	}
	
	// Crear estilos profesionales
	estilos := s.crearEstilosProfesionales(f)
	
	// Generar hoja APU con jerarquía real
	if err := s.generarHojaAPUJerarquica(f, apuSheet, proyecto, jerarquia, partidasConRecursos, estilos); err != nil {
		return "", fmt.Errorf("error generando hoja APU: %v", err)
	}
	
	// Generar hoja Presupuesto con jerarquía real  
	if err := s.generarHojaPresupuestoJerarquica(f, presupuestoSheet, proyecto, jerarquia, partidasConRecursos, estilos); err != nil {
		return "", fmt.Errorf("error generando hoja Presupuesto: %v", err)
	}
	
	// Generar nombre de archivo único
	timestamp := time.Now().Format("20060102_150405")
	nombreArchivo := fmt.Sprintf("APU_Presupuesto_%s_%s.xlsx", proyecto.Nombre, timestamp)
	rutaCompleta := filepath.Join(s.config.Files.ExcelOutputDir, nombreArchivo)
	
	return rutaCompleta, f.SaveAs(rutaCompleta)
}

// generarHojaAPUJerarquica genera la hoja de APU con estilo profesional y jerarquía real
func (s *ExcelJerarquicoService) generarHojaAPUJerarquica(f *excelize.File, sheet string, proyecto *models.Proyecto, jerarquia []ElementoJerarquico, partidas []models.PartidaCompleta, estilos map[string]int) error {
	// Configurar columnas
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 50)
	f.SetColWidth(sheet, "C", "C", 8)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 15)
	f.SetColWidth(sheet, "G", "G", 15)
	
	// Encabezado principal
	f.MergeCell(sheet, "A1", "G1")
	f.SetCellValue(sheet, "A1", "ANÁLISIS DE PRECIOS UNITARIOS")
	f.SetCellStyle(sheet, "A1", "G1", estilos["titulo_principal"])
	
	// Información del proyecto
	row := 3
	f.SetCellValue(sheet, "A3", "Proyecto:")
	f.SetCellValue(sheet, "B3", proyecto.Nombre)
	f.SetCellStyle(sheet, "A3", "A3", estilos["etiqueta"])
	f.SetCellStyle(sheet, "B3", "B3", estilos["datos"])
	
	row = 5
	
	// Crear mapa de partidas por código para acceso rápido
	partidasMap := make(map[string]models.PartidaCompleta)
	for _, partida := range partidas {
		partidasMap[partida.Codigo] = partida
	}
	
	// Mostrar jerarquía recursivamente con partidas detalladas
	row = s.mostrarJerarquiaAPU(f, sheet, jerarquia, partidasMap, row, estilos, 0)
	
	return nil
}

// mostrarJerarquiaAPU muestra la jerarquía recursivamente en formato APU
func (s *ExcelJerarquicoService) mostrarJerarquiaAPU(f *excelize.File, sheet string, elementos []ElementoJerarquico, partidasMap map[string]models.PartidaCompleta, row int, estilos map[string]int, nivel int) int {
	for _, elem := range elementos {
		if elem.TipoElemento == "titulo" {
			// Mostrar título jerárquico
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			titulo := fmt.Sprintf("%s %s", elem.Codigo, elem.Descripcion)
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), titulo)
			
			// Aplicar estilo según nivel
			estiloNivel := s.obtenerEstiloPorNivel(estilos, nivel)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estiloNivel)
			row += 2
			
			// Mostrar hijos recursivamente
			row = s.mostrarJerarquiaAPU(f, sheet, elem.Hijos, partidasMap, row, estilos, nivel+1)
			
		} else {
			// Es una partida - mostrar detalle completo
			if partida, existe := partidasMap[elem.Codigo]; existe {
				row = s.mostrarPartidaDetalladaAPU(f, sheet, partida, row, estilos)
			}
		}
	}
	return row
}

// mostrarPartidaDetalladaAPU muestra una partida con todos sus recursos
func (s *ExcelJerarquicoService) mostrarPartidaDetalladaAPU(f *excelize.File, sheet string, partida models.PartidaCompleta, row int, estilos map[string]int) int {
	// Encabezado de partida
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("Partida %s - %s", partida.Codigo, partida.Descripcion))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["partida"])
	row++
	
	// Información básica
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Unidad:")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), partida.Unidad)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Rendimiento:")
	f.SetCellValue(sheet, fmt.Sprintf("D%d", row), partida.Rendimiento)
	f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Costo Total:")
	f.SetCellValue(sheet, fmt.Sprintf("F%d", row), partida.CostoTotal)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), estilos["etiqueta"])
	f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), estilos["numero"])
	row++
	
	// Cabeceras de tabla de recursos
	headers := []string{"Código", "Descripción", "Unidad", "Cuadrilla", "Cantidad", "Precio S/.", "Parcial S/."}
	for j, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+j, row), header)
		f.SetCellStyle(sheet, fmt.Sprintf("%c%d", 'A'+j, row), fmt.Sprintf("%c%d", 'A'+j, row), estilos["cabecera"])
	}
	row++
	
	// Mostrar recursos por tipo con detalles completos
	row = s.mostrarRecursosPorTipo(f, sheet, partida, row, estilos)
	
	// Espacio entre partidas
	row += 2
	
	return row
}

// generarHojaPresupuestoJerarquica genera la hoja de presupuesto con jerarquía colapsable
func (s *ExcelJerarquicoService) generarHojaPresupuestoJerarquica(f *excelize.File, sheet string, proyecto *models.Proyecto, jerarquia []ElementoJerarquico, partidas []models.PartidaCompleta, estilos map[string]int) error {
	// Configurar columnas para presupuesto
	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 50)
	f.SetColWidth(sheet, "C", "C", 8)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 15)
	f.SetColWidth(sheet, "F", "F", 18)
	
	// Título principal
	f.MergeCell(sheet, "A1", "F1")
	f.SetCellValue(sheet, "A1", "PRESUPUESTO GENERAL")
	f.SetCellStyle(sheet, "A1", "F1", estilos["titulo_principal"])
	
	// Información del proyecto
	f.SetCellValue(sheet, "A3", "Proyecto:")
	f.SetCellValue(sheet, "B3", proyecto.Nombre)
	f.SetCellStyle(sheet, "A3", "A3", estilos["etiqueta"])
	f.SetCellStyle(sheet, "B3", "B3", estilos["datos"])
	
	// Cabeceras
	row := 5
	headers := []string{"Ítem", "Descripción", "Und.", "Metrado", "Precio S/.", "Parcial S/."}
	for i, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+i, row), header)
		f.SetCellStyle(sheet, fmt.Sprintf("%c%d", 'A'+i, row), fmt.Sprintf("%c%d", 'A'+i, row), estilos["cabecera"])
	}
	row++
	
	// Crear mapa de partidas por código
	partidasMap := make(map[string]models.PartidaCompleta)
	for _, partida := range partidas {
		partidasMap[partida.Codigo] = partida
	}
	
	totalGeneral := 0.0
	
	// Mostrar jerarquía de presupuesto
	row, totalGeneral = s.mostrarJerarquiaPresupuesto(f, sheet, jerarquia, partidasMap, row, estilos, 0, &totalGeneral)
	
	// Total general
	row++
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("E%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "COSTO DIRECTO")
	f.SetCellValue(sheet, fmt.Sprintf("F%d", row), totalGeneral)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), estilos["total"])
	
	return nil
}

// mostrarJerarquiaPresupuesto muestra jerarquía con subtotales por grupo
func (s *ExcelJerarquicoService) mostrarJerarquiaPresupuesto(f *excelize.File, sheet string, elementos []ElementoJerarquico, partidasMap map[string]models.PartidaCompleta, row int, estilos map[string]int, nivel int, totalGeneral *float64) (int, float64) {
	subtotalGrupo := 0.0
	
	for _, elem := range elementos {
		if elem.TipoElemento == "titulo" {
			// Título de grupo
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			titulo := fmt.Sprintf("%s %s", elem.Codigo, elem.Descripcion)
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), titulo)
			
			estiloNivel := s.obtenerEstiloPorNivel(estilos, nivel)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), estiloNivel)
			row++
			
			// Procesar hijos y acumular subtotal
			var subtotalHijos float64
			row, subtotalHijos = s.mostrarJerarquiaPresupuesto(f, sheet, elem.Hijos, partidasMap, row, estilos, nivel+1, totalGeneral)
			subtotalGrupo += subtotalHijos
			
			// Mostrar subtotal si el grupo tiene partidas
			if subtotalHijos > 0 && nivel < 2 { // Solo mostrar subtotales en niveles principales
				f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("E%d", row))
				f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("Subtotal %s", elem.Codigo))
				f.SetCellValue(sheet, fmt.Sprintf("F%d", row), subtotalHijos)
				f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), estilos["subtotal"])
				row++
			}
			
		} else {
			// Partida individual
			if partida, existe := partidasMap[elem.Codigo]; existe {
				metrado := 1.0 // Por defecto
				costoUnitario := partida.CostoTotal
				parcial := costoUnitario * metrado
				
				f.SetCellValue(sheet, fmt.Sprintf("A%d", row), partida.Codigo)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", row), partida.Descripcion)
				f.SetCellValue(sheet, fmt.Sprintf("C%d", row), partida.Unidad)
				f.SetCellValue(sheet, fmt.Sprintf("D%d", row), metrado)
				f.SetCellValue(sheet, fmt.Sprintf("E%d", row), costoUnitario)
				f.SetCellValue(sheet, fmt.Sprintf("F%d", row), parcial)
				
				f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), estilos["datos"])
				f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("F%d", row), estilos["numero"])
				
				subtotalGrupo += parcial
				*totalGeneral += parcial
				row++
			}
		}
	}
	
	return row, subtotalGrupo
}

// mostrarRecursosPorTipo muestra los recursos de una partida por categorías
func (s *ExcelJerarquicoService) mostrarRecursosPorTipo(f *excelize.File, sheet string, partida models.PartidaCompleta, row int, estilos map[string]int) int {
	// En la implementación completa, aquí obtendríamos los recursos desde la BD
	// Por ahora mostramos solo los subtotales que ya están calculados
	
	// Mano de obra
	if partida.CostoManoObra > 0 {
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "MANO DE OBRA")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["nivel_3"])
		row++
		
		// Subtotal MO
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL MANO DE OBRA")
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoManoObra)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["subtotal"])
		row++
	}

	// Materiales
	if partida.CostoMateriales > 0 {
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "MATERIALES")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["nivel_3"])
		row++
		
		// Subtotal Materiales
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL MATERIALES")
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoMateriales)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["subtotal"])
		row++
	}

	// Equipos
	if partida.CostoEquipos > 0 {
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "EQUIPOS")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["nivel_3"])
		row++
		
		// Subtotal Equipos
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL EQUIPOS")
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoEquipos)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["subtotal"])
		row++
	}

	// Subcontratos
	if partida.CostoSubcontratos > 0 {
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBCONTRATOS")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["nivel_3"])
		row++
		
		// Subtotal Subcontratos
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL SUBCONTRATOS")
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoSubcontratos)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["subtotal"])
		row++
	}

	// Costo total de la partida
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("COSTO TOTAL - PARTIDA %s", partida.Codigo))
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoTotal)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["total"])
	row++

	return row
}

// obtenerEstiloPorNivel devuelve el estilo apropiado según el nivel jerárquico
func (s *ExcelJerarquicoService) obtenerEstiloPorNivel(estilos map[string]int, nivel int) int {
	switch nivel {
	case 0:
		return estilos["nivel_1"]
	case 1:
		return estilos["nivel_2"]
	case 2:
		return estilos["nivel_3"]
	default:
		return estilos["nivel_4"]
	}
}

// crearEstilosProfesionales crea todos los estilos necesarios
func (s *ExcelJerarquicoService) crearEstilosProfesionales(f *excelize.File) map[string]int {
	estilos := make(map[string]int)
	
	// Título principal
	tituloStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#2F5597"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Nivel 1 (01, 02, 03...)
	nivel1Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Nivel 2 (01.01, 01.02...)
	nivel2Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#8DB4E2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Nivel 3 (01.01.01, 01.01.02...)
	nivel3Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#C5D9F1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Nivel 4+ (01.01.01.01...)
	nivel4Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E7E6E6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Partida
	partidaStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#305496"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Cabecera
	cabeceraStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Datos normales
	datosStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 9},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Números
	numeroStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 9},
		NumFmt: 4,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Etiquetas
	etiquetaStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 9},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	
	// Subtotal
	subtotalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E7E6E6"}, Pattern: 1},
		NumFmt: 4,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	// Total
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#70AD47"}, Pattern: 1},
		NumFmt: 4,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	
	estilos["titulo_principal"] = tituloStyle
	estilos["nivel_1"] = nivel1Style
	estilos["nivel_2"] = nivel2Style
	estilos["nivel_3"] = nivel3Style
	estilos["nivel_4"] = nivel4Style
	estilos["partida"] = partidaStyle
	estilos["cabecera"] = cabeceraStyle
	estilos["datos"] = datosStyle
	estilos["numero"] = numeroStyle
	estilos["etiqueta"] = etiquetaStyle
	estilos["subtotal"] = subtotalStyle
	estilos["total"] = totalStyle
	
	return estilos
}

