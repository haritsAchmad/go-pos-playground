package main

import (
	"context"
	"log"
	"net/http"

	"go-pos-playground/internal/auth"
	"go-pos-playground/internal/config"
	"go-pos-playground/internal/database"
	"go-pos-playground/internal/handler"
	"go-pos-playground/internal/repository"
	"go-pos-playground/internal/router"
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
	tokens, err := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiryMinutes)
	if err != nil {
		log.Fatal(err)
	}
	refreshTokens, err := auth.NewRefreshManager(cfg.RefreshExpiryDays)
	if err != nil {
		log.Fatal(err)
	}
	authRepo := repository.NewAuthRepository(db, cfg.DBSchema)
	if err := authRepo.SeedAdmin(ctx, cfg.AdminName, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatal(err)
	}
	authHandler := handler.NewAuthHandler(authRepo, tokens, refreshTokens)
	auditRepo := repository.NewAuditRepository(db, cfg.DBSchema)
	auditHandler := handler.NewAuditHandler(auditRepo)

	itemRepo := repository.NewItemRepository(db, cfg.DBSchema)
	itemHandler := handler.NewItemHandler(itemRepo)
	supplierRepo := repository.NewSupplierRepository(db, cfg.DBSchema)
	supplierHandler := handler.NewSupplierHandler(supplierRepo)
	cooperativeRepo := repository.NewCooperativeRepository(db, cfg.DBSchema)
	cooperativeHandler := handler.NewCooperativeHandler(cooperativeRepo)

	r := router.New(itemHandler, supplierHandler, cooperativeHandler, authHandler, auditHandler, tokens, authRepo, auditRepo)

	log.Println("Server running at :" + cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, r); err != nil {
		log.Fatal(err)
	}
}
