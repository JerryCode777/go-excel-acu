package main

import (
	"encoding/json"
	"fmt"
	"os"

	"goexcel/internal/services"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run test_parser.go <archivo.acu>")
		os.Exit(1)
	}

	archivo := os.Args[1]
	
	// Leer archivo
	content, err := os.ReadFile(archivo)
	if err != nil {
		fmt.Printf("Error leyendo archivo: %v\n", err)
		os.Exit(1)
	}

	// Crear parser
	parser := services.NewACUJerarquicoParser()
	
	// Parsear contenido
	result, err := parser.ParseACUJerarquico(string(content))
	if err != nil {
		fmt.Printf("Error parseando ACU: %v\n", err)
		os.Exit(1)
	}

	// Mostrar resultados
	fmt.Printf("=== PRESUPUESTO ===\n")
	fmt.Printf("Código: %s\n", result.Presupuesto.Codigo)
	fmt.Printf("Nombre: %s\n", result.Presupuesto.Nombre)
	if result.Presupuesto.Cliente != nil {
		fmt.Printf("Cliente: %s\n", *result.Presupuesto.Cliente)
	}
	fmt.Printf("Moneda: %s\n", result.Presupuesto.Moneda)
	fmt.Println()

	if len(result.Subpresupuestos) > 0 {
		fmt.Printf("=== SUBPRESUPUESTOS ===\n")
		for _, sub := range result.Subpresupuestos {
			fmt.Printf("- %s: %s\n", sub.Codigo, sub.Nombre)
		}
		fmt.Println()
	}

	fmt.Printf("=== TÍTULOS ===\n")
	for _, titulo := range result.Titulos {
		fmt.Printf("%s (Nivel %d): %s\n", titulo.CodigoCompleto, titulo.Nivel, titulo.Nombre)
	}
	fmt.Println()

	fmt.Printf("=== PARTIDAS ===\n")
	for _, partida := range result.Partidas {
		fmt.Printf("%s: %s [%s] (Rend: %.1f)\n", 
			partida.Codigo, partida.Descripcion, partida.Unidad, partida.Rendimiento)
		
		if len(partida.ManoObra) > 0 {
			fmt.Printf("  Mano de Obra:\n")
			for _, recurso := range partida.ManoObra {
				fmt.Printf("    - %s: %s (%.4f %s @ S/ %.2f)\n", 
					recurso.Codigo, recurso.Descripcion, recurso.Cantidad, recurso.Unidad, recurso.Precio)
			}
		}
		
		if len(partida.Materiales) > 0 {
			fmt.Printf("  Materiales:\n")
			for _, recurso := range partida.Materiales {
				fmt.Printf("    - %s: %s (%.4f %s @ S/ %.2f)\n", 
					recurso.Codigo, recurso.Descripcion, recurso.Cantidad, recurso.Unidad, recurso.Precio)
			}
		}
		
		if len(partida.Equipos) > 0 {
			fmt.Printf("  Equipos:\n")
			for _, recurso := range partida.Equipos {
				fmt.Printf("    - %s: %s (%.4f %s @ S/ %.2f)\n", 
					recurso.Codigo, recurso.Descripcion, recurso.Cantidad, recurso.Unidad, recurso.Precio)
			}
		}
		fmt.Println()
	}

	// Mostrar JSON para debug
	fmt.Printf("=== JSON DEBUG ===\n")
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}