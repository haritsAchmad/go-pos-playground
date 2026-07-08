package router

import (
	"net/http"

	"go-inventory-playground/internal/handler"
)

func New(itemHandler *handler.ItemHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/items", func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		switch r.Method {

		case http.MethodGet:
			itemHandler.FindAll(w, r)

		case http.MethodPost:
			itemHandler.Create(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			itemHandler.FindByID(w, r)
		case http.MethodPut:
			itemHandler.Update(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
