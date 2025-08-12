package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	sheet := "ACU"
	f.SetSheetName("Sheet1", sheet)

	// Configurar anchos de columnas
	f.SetColWidth(sheet, "A", "A", 12)  // Código
	f.SetColWidth(sheet, "B", "B", 45)  // Descripción Recurso
	f.SetColWidth(sheet, "C", "C", 10)  // Unidad
	f.SetColWidth(sheet, "D", "D", 12)  // Cuadrilla
	f.SetColWidth(sheet, "E", "E", 12)  // Cantidad
	f.SetColWidth(sheet, "F", "F", 15)  // Precio S/.
	f.SetColWidth(sheet, "G", "G", 15)  // Parcial S/.

	// Encabezado principal de la partida
	f.MergeCell(sheet, "A1", "G1")
	f.SetCellValue(sheet, "A1", "Partida        01.02                    MURO PARAPETO DE SOGA LADRILLO SILICO CALCAREO KK CON CEMENTO-ARENA (RVG)")
	
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A1", "G1", headerStyle)

	// Información de rendimiento y costo
	f.MergeCell(sheet, "A2", "G2")
	f.SetCellValue(sheet, "A2", "Rendimiento        m2/DIA        MO. 9.8000                    EQ.  9.8000                                        Costo unitario directo por : m2                    33.65")
	f.SetCellStyle(sheet, "A2", "G2", headerStyle)

	// Cabeceras de la tabla
	headers := []string{"Código", "Descripción Recurso", "Unidad", "Cuadrilla", "Cantidad", "Precio S/.", "Parcial S/."}
	subHeader := []string{"", "Mano de Obra", "", "", "", "", ""}
	
	columnStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Establecer cabeceras principales
	for i, header := range headers {
		cell := fmt.Sprintf("%c3", 'A'+i)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, columnStyle)
	}

	// Establecer subcabecera para Mano de Obra
	for i, subH := range subHeader {
		cell := fmt.Sprintf("%c4", 'A'+i)
		f.SetCellValue(sheet, cell, subH)
		f.SetCellStyle(sheet, cell, cell, columnStyle)
	}

	// Estilo para datos
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		NumFmt: 4, // Formato numérico con 2 decimales
	})

	row := 5
	
	// MANO DE OBRA
	manoObra := []map[string]interface{}{
		{"codigo": "0147010001", "descripcion": "CAPATAZ", "unidad": "hh", "cuadrilla": 0.1000, "cantidad": 0.0816, "precio": 15.00, "parcial": 1.22},
		{"codigo": "0147010002", "descripcion": "OPERARIO", "unidad": "hh", "cuadrilla": 1.0000, "cantidad": 0.8163, "precio": 14.97, "parcial": 12.22},
		{"codigo": "0147010004", "descripcion": "PEON", "unidad": "hh", "cuadrilla": 0.7500, "cantidad": 0.6122, "precio": 11.73, "parcial": 7.18},
	}

	for _, item := range manoObra {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item["codigo"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item["descripcion"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item["unidad"])
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item["cuadrilla"])
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item["cantidad"])
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item["precio"])
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item["parcial"])
		
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("G%d", row), numberStyle)
		row++
	}

	// Subtotal Mano de Obra
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), 20.62)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), columnStyle)
	row++

	// MATERIALES
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Materiales")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), columnStyle)
	row++

	materiales := []map[string]interface{}{
		{"codigo": "0202010005", "descripcion": "CLAVOS PARA MADERA CON CABEZA DE 3\"", "unidad": "kg", "cuadrilla": "", "cantidad": 0.0200, "precio": 4.20, "parcial": 0.08},
		{"codigo": "0205010004", "descripcion": "ARENA GRUESA", "unidad": "m3", "cuadrilla": "", "cantidad": 0.0300, "precio": 25.50, "parcial": 0.77},
		{"codigo": "0217510002", "descripcion": "BLOQUE SILICO ESTANDARD 14 X 25 X 9 cm", "unidad": "u", "cuadrilla": "", "cantidad": 38.0000, "precio": 0.21, "parcial": 7.98},
		{"codigo": "0221000001", "descripcion": "CEMENTO PORTLAND TIPO I (42.5 kg)", "unidad": "bls", "cuadrilla": "", "cantidad": 0.1100, "precio": 15.55, "parcial": 1.71},
		{"codigo": "0239050000", "descripcion": "AGUA", "unidad": "m3", "cuadrilla": "", "cantidad": 0.0080, "precio": 1.83, "parcial": 0.01},
		{"codigo": "0243040000", "descripcion": "MADERA TORNILLO", "unidad": "p2", "cuadrilla": "", "cantidad": 0.5800, "precio": 3.20, "parcial": 1.86},
	}

	for _, item := range materiales {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item["codigo"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item["descripcion"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item["unidad"])
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item["cuadrilla"])
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item["cantidad"])
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item["precio"])
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item["parcial"])
		
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("G%d", row), numberStyle)
		row++
	}

	// Subtotal Materiales
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), 12.41)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), columnStyle)
	row++

	// EQUIPOS
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Equipos")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), columnStyle)
	row++

	equipos := []map[string]interface{}{
		{"codigo": "0337010001", "descripcion": "HERRAMIENTAS MANUALES", "unidad": "%MO", "cuadrilla": "", "cantidad": 3.0000, "precio": 20.62, "parcial": 0.62},
	}

	for _, item := range equipos {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item["codigo"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item["descripcion"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item["unidad"])
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item["cuadrilla"])
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item["cantidad"])
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item["precio"])
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item["parcial"])
		
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("G%d", row), numberStyle)
		row++
	}

	// Subtotal Equipos
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), 0.62)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), columnStyle)

	// Guardar archivo
	if err := f.SaveAs("ACU_Partida_Completo.xlsx"); err != nil {
		fmt.Println("Error guardando archivo:", err)
	} else {
		fmt.Println("✅ ACU generado correctamente - Formato completo")
	}
}