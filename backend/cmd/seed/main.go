package main

import (
	"context"
	"flag"
	"log"

	"go-pos-playground/internal/config"
	"go-pos-playground/internal/database"
	"go-pos-playground/internal/seed"
)

func main() {
	var o seed.Options
	flag.IntVar(&o.Items, "items", 20, "number of demo items")
	flag.IntVar(&o.Customers, "customers", 30, "number of demo customers")
	flag.IntVar(&o.Suppliers, "suppliers", 8, "number of demo suppliers")
	flag.IntVar(&o.Purchases, "purchases", 60, "number of demo purchases")
	flag.IntVar(&o.Sales, "sales", 150, "number of demo sales")
	flag.IntVar(&o.Months, "months", 6, "spread transactions across this many months")
	flag.Float64Var(&o.DebtRate, "debt-rate", 0.2, "fraction of sales that create debt")
	flag.Int64Var(&o.RandomSeed, "seed", 20260720, "deterministic random seed")
	flag.Parse()
	ctx := context.Background()
	cfg := config.Load()
	db, err := database.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(ctx, db, cfg.DBSchema); err != nil {
		log.Fatal(err)
	}
	result, err := seed.Generate(ctx, db, cfg.DBSchema, o)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seed complete: %+v", result)
}
