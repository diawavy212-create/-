package main

import (
	"log"

	"teacher-platform/server/internal/config"
	"teacher-platform/server/internal/infra/mysql"
	"teacher-platform/server/internal/router"
)

func main() {
	cfg := config.Load()
	db, err := mysql.Open(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	engine := router.New(cfg, db)

	log.Printf("api gateway listening on %s", cfg.HTTPAddr)
	if err := engine.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
