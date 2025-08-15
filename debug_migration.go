package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jerryandersonh/goexcel/config"
	"github.com/jerryandersonh/goexcel/internal/database"
	"github.com/jerryandersonh/goexcel/internal/legacy"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("🔍 Debug migración PostgreSQL")

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	// Conectar a base de datos
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Error conectando a base de datos: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Conectado a PostgreSQL")

	// Leer JSON
	data, err := os.ReadFile("partidas.json")
	if err != nil {
		log.Fatalf("❌ Error leyendo JSON: %v", err)
	}

	var partidasJSON []legacy.PartidaLegacy
	if err := json.Unmarshal(data, &partidasJSON); err != nil {
		log.Fatalf("❌ Error parseando JSON: %v", err)
	}

	fmt.Printf("✅ JSON parseado: %d partidas\n", len(partidasJSON))

	// PASO 1: Crear proyecto manualmente
	fmt.Println("\n🔸 PASO 1: Crear proyecto")
	proyectoID := uuid.New()
	query := `INSERT INTO proyectos (id, nombre, moneda) VALUES ($1, $2, $3)`
	
	_, err = db.Exec(query, proyectoID, "Debug Migration Test", "PEN")
	if err != nil {
		log.Fatalf("❌ Error creando proyecto: %v", err)
	}
	
	fmt.Printf("✅ Proyecto creado: %s\n", proyectoID.String()[:8])

	// Verificar que se guardó
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM proyectos").Scan(&count)
	if err != nil {
		log.Fatalf("❌ Error verificando proyectos: %v", err)
	}
	fmt.Printf("✅ Proyectos en BD: %d\n", count)

	// PASO 2: Obtener tipos de recurso
	fmt.Println("\n🔸 PASO 2: Verificar tipos de recurso")
	var tipoSubcontratosID uuid.UUID
	err = db.QueryRow("SELECT id FROM tipos_recurso WHERE nombre = 'subcontratos'").Scan(&tipoSubcontratosID)
	if err != nil {
		log.Fatalf("❌ Error obteniendo tipo subcontratos: %v", err)
	}
	fmt.Printf("✅ Tipo subcontratos: %s\n", tipoSubcontratosID.String()[:8])

	// PASO 3: Insertar UNA partida con subcontratos
	fmt.Println("\n🔸 PASO 3: Insertar partida de prueba")
	
	// Buscar partida con subcontratos (01.01.01.03)
	var partidaPrueba legacy.PartidaLegacy
	for _, p := range partidasJSON {
		if p.Codigo == "01.01.01.03" {
			partidaPrueba = p
			break
		}
	}
	
	if partidaPrueba.Codigo == "" {
		log.Fatal("❌ No se encontró partida 01.01.01.03")
	}
	
	fmt.Printf("✅ Partida encontrada: %s - %s\n", partidaPrueba.Codigo, partidaPrueba.Descripcion)
	fmt.Printf("   Subcontratos: %d\n", len(partidaPrueba.Subcontratos))

	// Insertar partida
	partidaID := uuid.New()
	partidaQuery := `INSERT INTO partidas (id, proyecto_id, codigo, descripcion, unidad, rendimiento) 
					 VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err = db.Exec(partidaQuery, partidaID, proyectoID, partidaPrueba.Codigo, 
		partidaPrueba.Descripcion, partidaPrueba.Unidad, partidaPrueba.Rendimiento)
	if err != nil {
		log.Fatalf("❌ Error insertando partida: %v", err)
	}
	
	fmt.Printf("✅ Partida insertada: %s\n", partidaID.String()[:8])

	// PASO 4: Insertar recursos y relaciones
	fmt.Println("\n🔸 PASO 4: Insertar subcontratos")
	
	for i, subcontrato := range partidaPrueba.Subcontratos {
		fmt.Printf("   Procesando: %s - %s\n", subcontrato.Codigo, subcontrato.Descripcion)
		
		// Insertar recurso
		recursoID := uuid.New()
		recursoQuery := `INSERT INTO recursos (id, codigo, descripcion, unidad, precio_base, tipo_recurso_id) 
						 VALUES ($1, $2, $3, $4, $5, $6)`
		
		_, err = db.Exec(recursoQuery, recursoID, subcontrato.Codigo, subcontrato.Descripcion,
			subcontrato.Unidad, subcontrato.Precio, tipoSubcontratosID)
		if err != nil {
			log.Printf("   ⚠️  Error insertando recurso %s: %v", subcontrato.Codigo, err)
			continue
		}
		
		// Insertar relación partida-recurso
		relacionQuery := `INSERT INTO partida_recursos (partida_id, recurso_id, cantidad, precio) 
						  VALUES ($1, $2, $3, $4)`
		
		_, err = db.Exec(relacionQuery, partidaID, recursoID, subcontrato.Cantidad, subcontrato.Precio)
		if err != nil {
			log.Printf("   ❌ Error insertando relación %s: %v", subcontrato.Codigo, err)
			continue
		}
		
		fmt.Printf("   ✅ Subcontrato %d insertado\n", i+1)
	}

	// PASO 5: Verificar resultados
	fmt.Println("\n🔸 PASO 5: Verificar resultados")
	
	err = db.QueryRow("SELECT COUNT(*) FROM proyectos").Scan(&count)
	if err == nil {
		fmt.Printf("✅ Proyectos: %d\n", count)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM partidas").Scan(&count)
	if err == nil {
		fmt.Printf("✅ Partidas: %d\n", count)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM recursos").Scan(&count)
	if err == nil {
		fmt.Printf("✅ Recursos: %d\n", count)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM partida_recursos").Scan(&count)
	if err == nil {
		fmt.Printf("✅ Relaciones: %d\n", count)
	}

	// Verificar costo total calculado
	var costoTotal float64
	err = db.QueryRow("SELECT costo_total FROM partidas WHERE id = $1", partidaID).Scan(&costoTotal)
	if err == nil {
		fmt.Printf("✅ Costo total calculado: %.2f\n", costoTotal)
	}

	fmt.Println("\n🎉 Debug completado")
}