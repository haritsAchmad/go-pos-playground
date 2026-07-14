package router

import (
	"net/http"

	"go-inventory-playground/internal/handler"
)

func New(
	itemHandler *handler.ItemHandler,
	supplierHandler *handler.SupplierHandler,
	cooperativeHandler *handler.CooperativeHandler,
) *http.ServeMux {
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
		case http.MethodDelete:
			itemHandler.Delete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/suppliers", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.Method {
		case http.MethodGet:
			supplierHandler.FindAll(w, r)
		case http.MethodPost:
			supplierHandler.Create(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/suppliers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			supplierHandler.FindByID(w, r)
		case http.MethodPut:
			supplierHandler.Update(w, r)
		case http.MethodDelete:
			supplierHandler.Delete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/dashboard", cooperativeHandler.Dashboard)
	mux.HandleFunc("/masters/", cooperativeHandler.Masters)
	mux.HandleFunc("/customers", cooperativeHandler.Customers)
	mux.HandleFunc("/customers/", cooperativeHandler.CustomerDetail)
	mux.HandleFunc("/transactions", cooperativeHandler.Transactions)
	mux.HandleFunc("/transactions/", cooperativeHandler.VoidTransaction)
	mux.HandleFunc("/debts", cooperativeHandler.Debts)
	mux.HandleFunc("/debts/", cooperativeHandler.PayDebt)

	return mux
}
