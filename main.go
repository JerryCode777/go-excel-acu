package main

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/xuri/excelize/v2"
)

// Estructura para recursos individuales
type Recurso struct {
	Codigo      string  `json:"codigo"`
	Descripcion string  `json:"descripcion"`
	Unidad      string  `json:"unidad"`
	Cuadrilla   float64 `json:"cuadrilla,omitempty"`
	Cantidad    float64 `json:"cantidad"`
	Precio      float64 `json:"precio"`
}

// Estructura para partidas
type Partida struct {
	Codigo      string    `json:"codigo"`
	Descripcion string    `json:"descripcion"`
	Unidad      string    `json:"unidad"`
	Rendimiento float64   `json:"rendimiento"`
	ManoObra    []Recurso `json:"mano_obra"`
	Materiales  []Recurso `json:"materiales"`
	Equipos     []Recurso `json:"equipos"`
}

func main() {
	// Leer archivo JSON desde argumentos o usar por defecto
	archivoJSON := "partidas.json"
	if len(os.Args) > 1 {
		archivoJSON = os.Args[1]
	}
	
	fmt.Printf("📖 Leyendo archivo: %s\n", archivoJSON)
	
	// Verificar si el archivo existe
	if _, err := os.Stat(archivoJSON); os.IsNotExist(err) {
		fmt.Printf("❌ El archivo %s no existe\n", archivoJSON)
		return
	}
	
	// Leer y procesar JSON
	data, err := os.ReadFile(archivoJSON)
	if err != nil {
		fmt.Printf("❌ Error leyendo archivo: %v\n", err)
		return
	}

	var partidas []Partida
	if err := json.Unmarshal(data, &partidas); err != nil {
		fmt.Printf("❌ Error parseando JSON: %v\n", err)
		return
	}

	if len(partidas) == 0 {
		fmt.Println("❌ No se encontraron partidas en el archivo")
		return
	}

	// Generar Excel
	nombreExcel := "ACUs_Consolidado.xlsx"
	if err := generarExcel(partidas, nombreExcel); err != nil {
		fmt.Printf("❌ Error generando Excel: %v\n", err)
		return
	}

	fmt.Printf("✅ Archivo generado: %s\n", nombreExcel)
	fmt.Printf("📊 %d partidas procesadas\n", len(partidas))
}

func generarExcel(partidas []Partida, nombreArchivo string) error {
	f := excelize.NewFile()
	defer f.Close()
	
	sheet := "ACUs"
	f.SetSheetName("Sheet1", sheet)
	
	// Crear hoja de resumen
	sheetResumen := "Resumen"
	f.NewSheet(sheetResumen)
	
	// Configurar columnas
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 45)
	f.SetColWidth(sheet, "C", "C", 10)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 15)
	f.SetColWidth(sheet, "G", "G", 15)

	// Estilos
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

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		NumFmt: 4, // Formato con 2 decimales
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
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

	// Título principal
	f.MergeCell(sheet, "A1", "G1")
	f.SetCellValue(sheet, "A1", "ANÁLISIS DE COSTOS UNITARIOS - CONSOLIDADO")
	f.SetCellStyle(sheet, "A1", "G1", headerStyle)

	row := 3
	var datosResumen []map[string]interface{}

	// Procesar cada partida
	for i, partida := range partidas {
		fmt.Printf("Procesando partida %d/%d: %s\n", i+1, len(partidas), partida.Codigo)
		
		// Validar partida
		if partida.Codigo == "" || partida.Descripcion == "" {
			fmt.Printf("⚠️  Saltando partida %d: datos incompletos\n", i+1)
			continue
		}

		// Calcular totales
		totalMO := calcularTotal(partida.ManoObra)
		totalMat := calcularTotal(partida.Materiales)
		totalEq := calcularTotal(partida.Equipos)
		costoTotal := totalMO + totalMat + totalEq

		// Guardar para resumen
		datosResumen = append(datosResumen, map[string]interface{}{
			"codigo":      partida.Codigo,
			"descripcion": partida.Descripcion,
			"unidad":      partida.Unidad,
			"rendimiento": partida.Rendimiento,
			"costo_mo":    totalMO,
			"costo_mat":   totalMat,
			"costo_eq":    totalEq,
			"costo_total": costoTotal,
		})

		// Encabezado de partida
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), 
			fmt.Sprintf("PARTIDA %s - %s", partida.Codigo, partida.Descripcion))
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), partidaStyle)
		row++

		// Info de la partida
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Unidad:")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), partida.Unidad)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Rendimiento:")
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), partida.Rendimiento)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Costo Total:")
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), costoTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), totalStyle)
		row++

		// Cabeceras de tabla
		headers := []string{"Código", "Descripción", "Unidad", "Cuadrilla", "Cantidad", "Precio S/", "Parcial S/"}
		for j, header := range headers {
			f.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+j, row), header)
			f.SetCellStyle(sheet, fmt.Sprintf("%c%d", 'A'+j, row), fmt.Sprintf("%c%d", 'A'+j, row), sectionStyle)
		}
		row++

		// Mano de obra
		if len(partida.ManoObra) > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "MANO DE OBRA")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), sectionStyle)
			row++
			
			row = agregarRecursos(f, sheet, partida.ManoObra, row, dataStyle, numberStyle)
			
			// Subtotal MO
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL MANO DE OBRA")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), totalMO)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), sectionStyle)
			row++
		}

		// Materiales
		if len(partida.Materiales) > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "MATERIALES")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), sectionStyle)
			row++
			
			row = agregarRecursos(f, sheet, partida.Materiales, row, dataStyle, numberStyle)
			
			// Subtotal Materiales
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL MATERIALES")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), totalMat)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), sectionStyle)
			row++
		}

		// Equipos
		if len(partida.Equipos) > 0 {
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "EQUIPOS")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), sectionStyle)
			row++
			
			row = agregarRecursos(f, sheet, partida.Equipos, row, dataStyle, numberStyle)
			
			// Subtotal Equipos
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "SUBTOTAL EQUIPOS")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), totalEq)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), sectionStyle)
			row++
		}

		// Costo total de la partida
		f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("COSTO TOTAL - PARTIDA %s", partida.Codigo))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), costoTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), totalStyle)
		row += 3 // Espaciado entre partidas
	}

	// Crear hoja resumen
	crearResumen(f, sheetResumen, datosResumen)

	return f.SaveAs(nombreArchivo)
}

func agregarRecursos(f *excelize.File, sheet string, recursos []Recurso, startRow int, dataStyle, numberStyle int) int {
	row := startRow
	for _, recurso := range recursos {
		// Validar recurso
		if recurso.Codigo == "" || recurso.Descripcion == "" {
			continue
		}
		
		parcial := recurso.Cantidad * recurso.Precio
		
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), recurso.Codigo)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), recurso.Descripcion)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), recurso.Unidad)
		
		// Cuadrilla solo si es mayor a 0
		if recurso.Cuadrilla > 0 {
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), recurso.Cuadrilla)
		} else {
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "-")
		}
		
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), recurso.Cantidad)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), recurso.Precio)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), parcial)
		
		// Aplicar estilos
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("G%d", row), numberStyle)
		row++
	}
	return row
}

func crearResumen(f *excelize.File, sheet string, datos []map[string]interface{}) {
	if len(datos) == 0 {
		return
	}

	// Configurar columnas para resumen
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 45)
	f.SetColWidth(sheet, "C", "C", 10)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 15)
	f.SetColWidth(sheet, "F", "F", 15)
	f.SetColWidth(sheet, "G", "G", 15)
	f.SetColWidth(sheet, "H", "H", 15)

	// Estilos para resumen
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
	f.MergeCell(sheet, "A1", "H1")
	f.SetCellValue(sheet, "A1", "RESUMEN DE COSTOS UNITARIOS")
	f.SetCellStyle(sheet, "A1", "H1", titleStyle)

	// Cabeceras
	headers := []string{"Código", "Descripción", "Unidad", "Rendimiento", "Mano Obra", "Materiales", "Equipos", "Costo Total"}
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
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), dato["costo_total"])
		
		// Aplicar formato numérico a las columnas de números
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("H%d", row), numberStyle)
		row++
	}
}

func calcularTotal(recursos []Recurso) float64 {
	total := 0.0
	for _, recurso := range recursos {
		total += recurso.Cantidad * recurso.Precio
	}
	return total
}