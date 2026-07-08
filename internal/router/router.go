package router

import (
	"net/http"

	"go-inventory-playground/internal/handler"
)

func New(itemHandler *handler.ItemHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/items", itemHandler.FindAll)

	return mux
}
