package main

import (
	"log"

	"org-structure-api/internal/app"
	"org-structure-api/internal/config"
)

func main() {
	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	log.Printf("starting server on %s", cfg.HTTPAddr)
	if err := application.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
