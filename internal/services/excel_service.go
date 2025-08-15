package services

import (
	"fmt"

	"github.com/xuri/excelize/v2"
	"goexcel/config"
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

