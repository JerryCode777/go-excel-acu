package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jerryandersonh/goexcel/config"
	"github.com/jerryandersonh/goexcel/internal/database"
	"github.com/jerryandersonh/goexcel/internal/database/repositories"
	"github.com/jerryandersonh/goexcel/internal/legacy"
	"github.com/jerryandersonh/goexcel/internal/models"
)

func main() {
	fmt.Println("🧪 Test de migración simple")

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// Conectar a base de datos
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Error conectando a base de datos: %v", err)
	}
	defer db.Close()

	// Leer JSON
	data, err := os.ReadFile("partidas.json")
	if err != nil {
		log.Fatalf("Error leyendo JSON: %v", err)
	}

	var partidasJSON []legacy.PartidaLegacy
	if err := json.Unmarshal(data, &partidasJSON); err != nil {
		log.Fatalf("Error parseando JSON: %v", err)
	}

	fmt.Printf("📊 JSON cargado: %d partidas\n", len(partidasJSON))

	// Crear solo un proyecto para probar
	proyectoRepo := repositories.NewProyectoRepository(db)
	proyectoReq := &models.ProyectoCreateRequest{
		Nombre: "Test Migration Simple",
		Moneda: "PEN",
	}

	proyecto, err := proyectoRepo.Create(proyectoReq)
	if err != nil {
		log.Fatalf("Error creando proyecto: %v", err)
	}

	fmt.Printf("✅ Proyecto creado: %s (ID: %s)\n", proyecto.Nombre, proyecto.ID.String())

	// Verificar que se persistió
	proyectos, err := proyectoRepo.GetAll()
	if err != nil {
		log.Fatalf("Error obteniendo proyectos: %v", err)
	}

	fmt.Printf("📋 Proyectos en BD: %d\n", len(proyectos))
	for _, p := range proyectos {
		fmt.Printf("   - %s (ID: %s)\n", p.Nombre, p.ID.String()[:8])
	}
}