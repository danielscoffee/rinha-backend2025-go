package main

import (
	"log"
	"os"

	"github.com/danielscoffee/rinha-backend2025-go/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := app.NewServer()
	log.Printf("Starting server on port %s", port)
	if err := server.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
