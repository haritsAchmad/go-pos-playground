package main

import (
	"context"
	"log"
	"net/http"

	"go-inventory-playground/internal/config"
	"go-inventory-playground/internal/database"
	"go-inventory-playground/internal/handler"
	"go-inventory-playground/internal/repository"
	"go-inventory-playground/internal/router"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	db, err := database.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.Migrate(ctx, db, cfg.DBSchema); err != nil {
		log.Fatal(err)
	}
	log.Println("Database migration completed")

	itemRepo := repository.NewItemRepository(db, cfg.DBSchema)
	itemHandler := handler.NewItemHandler(itemRepo)
	supplierRepo := repository.NewSupplierRepository(db, cfg.DBSchema)
	supplierHandler := handler.NewSupplierHandler(supplierRepo)
	cooperativeRepo := repository.NewCooperativeRepository(db, cfg.DBSchema)
	cooperativeHandler := handler.NewCooperativeHandler(cooperativeRepo)

	r := router.New(itemHandler, supplierHandler, cooperativeHandler)

	log.Println("Server running at :" + cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, r); err != nil {
		log.Fatal(err)
	}
}
