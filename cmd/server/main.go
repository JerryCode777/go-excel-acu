package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jerryandersonh/goexcel/config"
	"github.com/jerryandersonh/goexcel/internal/database"
	"github.com/jerryandersonh/goexcel/internal/server"
)

func main() {
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

	// Ejecutar migraciones
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("❌ Error ejecutando migraciones: %v", err)
	}

	// Crear servidor
	srv := server.NewServer(cfg, db)

	// Manejo de señales para shutdown graceful
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Señal de interrupción recibida, cerrando servidor...")
		srv.Stop()
		os.Exit(0)
	}()

	// Iniciar servidor
	log.Fatal(srv.Start())
}