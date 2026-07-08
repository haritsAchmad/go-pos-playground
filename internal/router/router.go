package router

import (
	"net/http"

	"go-inventory-playground/internal/handler"
)

func New() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)

	return mux
}
