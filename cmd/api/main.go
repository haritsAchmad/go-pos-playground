package main

import (
	"log"
	"net/http"

	"go-inventory-playground/internal/router"
)

func main() {

	r := router.New()

	log.Println("Server running at :8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
