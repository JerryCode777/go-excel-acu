package services

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/jerryandersonh/goexcel/config"
	"github.com/jerryandersonh/goexcel/internal/models"
)

type ExcelService struct {
	config *config.Config
}

func NewExcelService(config *config.Config) *ExcelService {
	return &ExcelService{
		config: config,
	}
}

func (s *ExcelService) GetConfig() *config.Config {
	return s.config
}

func (s *ExcelService) GenerarExcelJerarquicoFromDB(proyecto *models.Proyecto, hierarchySvc *HierarchyService) (string, error) {
	f := excelize.NewFile()
	defer f.Close()
	
	sheet := "ACUs"
	f.SetSheetName("Sheet1", sheet)
	
	// Crear hoja de resumen
	sheetResumen := "Resumen"
	f.NewSheet(sheetResumen)
	
	// Configurar estilos y columnas (reutilizar del código original)
	s.configurarColumnas(f, sheet)
	estilos := s.crearEstilos(f)
	
	// Título principal
	f.MergeCell(sheet, "A1", "G1")
	f.SetCellValue(sheet, "A1", fmt.Sprintf("ANÁLISIS DE COSTOS UNITARIOS - %s", proyecto.Nombre))
	f.SetCellStyle(sheet, "A1", "G1", estilos["header"])

	row := 3
	var datosResumen []map[string]interface{}

	// Procesar cada partida
	for i, partida := range partidas {
		fmt.Printf("Procesando partida %d/%d: %s\n", i+1, len(partidas), partida.Codigo)
		
		// Guardar para resumen
		datosResumen = append(datosResumen, map[string]interface{}{
			"codigo":      partida.Codigo,
			"descripcion": partida.Descripcion,
			"unidad":      partida.Unidad,
			"rendimiento": partida.Rendimiento,
			"costo_mo":    partida.CostoManoObra,
			"costo_mat":   partida.CostoMateriales,
			"costo_eq":    partida.CostoEquipos,
			"costo_sub":   partida.CostoSubcontratos,
			"costo_total": partida.CostoTotal,
		})

		// Encabezado de partida
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), 
			fmt.Sprintf("PARTIDA %s - %s", partida.Codigo, partida.Descripcion))
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["partida"])
		row++

		// Info de la partida
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Unidad:")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), partida.Unidad)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Rendimiento:")
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), partida.Rendimiento)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Costo Total:")
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), partida.CostoTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), estilos["total"])
		row++

		// Cabeceras de tabla
		headers := []string{"Código", "Descripción", "Unidad", "Cuadrilla", "Cantidad", "Precio S/", "Parcial S/"}
		for j, header := range headers {
			f.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+j, row), header)
			f.SetCellStyle(sheet, fmt.Sprintf("%c%d", 'A'+j, row), fmt.Sprintf("%c%d", 'A'+j, row), estilos["section"])
		}
		row++

		// Aquí se agregarían los recursos por tipo
		// Por ahora, solo mostramos los subtotales
		
		// Mano de obra
		if partida.CostoManoObra > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "MANO DE OBRA")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
			
			// Subtotal MO
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL MANO DE OBRA")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoManoObra)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
		}

		// Materiales
		if partida.CostoMateriales > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "MATERIALES")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
			
			// Subtotal Materiales
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL MATERIALES")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoMateriales)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
		}

		// Equipos
		if partida.CostoEquipos > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "EQUIPOS")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
			
			// Subtotal Equipos
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL EQUIPOS")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoEquipos)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
		}

		// Subcontratos
		if partida.CostoSubcontratos > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBCONTRATOS")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
			
			// Subtotal Subcontratos
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL SUBCONTRATOS")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoSubcontratos)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["section"])
			row++
		}

		// Costo total de la partida
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("COSTO TOTAL - PARTIDA %s", partida.Codigo))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), partida.CostoTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), estilos["total"])
		row += 3 // Espaciado entre partidas
	}

	// Crear hoja resumen
	s.crearResumen(f, sheetResumen, datosResumen)

	// Generar nombre de archivo único
	timestamp := time.Now().Format("20060102_150405")
	nombreArchivo := fmt.Sprintf("ACUs_%s_%s.xlsx", proyecto.Nombre, timestamp)
	rutaCompleta := filepath.Join(s.config.Files.ExcelOutputDir, nombreArchivo)

	// Asegurar que el directorio existe
	// os.MkdirAll(s.config.Files.ExcelOutputDir, 0755)

	return rutaCompleta, f.SaveAs(rutaCompleta)
}

func (s *ExcelService) configurarColumnas(f *excelize.File, sheet string) {
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 45)
	f.SetColWidth(sheet, "C", "C", 10)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 15)
	f.SetColWidth(sheet, "G", "G", 15)
}

func (s *ExcelService) crearEstilos(f *excelize.File) map[string]int {
	estilos := make(map[string]int)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

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

	sectionStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E2F3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#70AD47"}, Pattern: 1},
		NumFmt: 4,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	estilos["header"] = headerStyle
	estilos["partida"] = partidaStyle
	estilos["section"] = sectionStyle
	estilos["total"] = totalStyle

	return estilos
}

func (s *ExcelService) crearResumen(f *excelize.File, sheet string, datos []map[string]interface{}) {
	if len(datos) == 0 {
		return
	}

	// Configurar columnas para resumen (con columna adicional para subcontratos)
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 45)
	f.SetColWidth(sheet, "C", "C", 10)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 15)
	f.SetColWidth(sheet, "F", "F", 15)
	f.SetColWidth(sheet, "G", "G", 15)
	f.SetColWidth(sheet, "H", "H", 15)
	f.SetColWidth(sheet, "I", "I", 15)

	// Estilos
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#2F5597"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// Título
	f.MergeCell(sheet, "A1", "I1")
	f.SetCellValue(sheet, "A1", "RESUMEN DE COSTOS UNITARIOS")
	f.SetCellStyle(sheet, "A1", "I1", titleStyle)

	// Cabeceras
	headers := []string{"Código", "Descripción", "Unidad", "Rendimiento", "Mano Obra", "Materiales", "Equipos", "Subcontratos", "Costo Total"}
	for i, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c3", 'A'+i), header)
		f.SetCellStyle(sheet, fmt.Sprintf("%c3", 'A'+i), fmt.Sprintf("%c3", 'A'+i), headerStyle)
	}

	// Datos
	row := 4
	for _, dato := range datos {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), dato["codigo"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), dato["descripcion"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), dato["unidad"])
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), dato["rendimiento"])
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), dato["costo_mo"])
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), dato["costo_mat"])
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), dato["costo_eq"])
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), dato["costo_sub"])
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), dato["costo_total"])
		
		// Aplicar formato numérico a las columnas de números
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("I%d", row), numberStyle)
		row++
	}
}

// GenerarExcelProfesionalAPU genera Excel en formato profesional APU y Presupuesto
func (s *ExcelService) GenerarExcelProfesionalAPU(proyecto *models.Proyecto, partidas []models.PartidaCompleta, metrados map[string]float64) (string, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Crear las dos hojas requeridas
	apuSheet := "Análisis de Precios Unitarios"
	presupuestoSheet := "Presupuesto General"
	
	f.SetSheetName("Sheet1", apuSheet)
	f.NewSheet(presupuestoSheet)

	// Configurar estilos profesionales
	estilos := s.crearEstilosProfesionales(f)

	// Generar hoja de APU
	if err := s.generarHojaAPU(f, apuSheet, proyecto, partidas, estilos); err != nil {
		return "", fmt.Errorf("error generando hoja APU: %v", err)
	}

	// Generar hoja de Presupuesto
	if err := s.generarHojaPresupuesto(f, presupuestoSheet, proyecto, partidas, metrados, estilos); err != nil {
		return "", fmt.Errorf("error generando hoja de Presupuesto: %v", err)
	}

	// Guardar archivo
	filename := fmt.Sprintf("APU_Presupuesto_%s_%d.xlsx", 
		proyecto.Nombre, time.Now().Unix())
	filepath := filepath.Join(s.config.Files.ExcelOutputDir, filename)

	if err := f.SaveAs(filepath); err != nil {
		return "", fmt.Errorf("error guardando archivo: %v", err)
	}

	return filepath, nil
}