package main

import (
	"log"

	"ao-regears/backend/internal/config"
	"ao-regears/backend/internal/httpapi"
	"ao-regears/backend/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		log.Fatal(err)
	}

	api := httpapi.New(cfg, store.New(db))
	if err := api.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
