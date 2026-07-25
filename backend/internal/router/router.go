package router

import (
	"net/http"

	"go-pos-playground/internal/auth"
	"go-pos-playground/internal/handler"
	"go-pos-playground/internal/middleware"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"
)

func New(
	itemHandler *handler.ItemHandler,
	supplierHandler *handler.SupplierHandler,
	cooperativeHandler *handler.CooperativeHandler,
	authHandler *handler.AuthHandler,
	auditHandler *handler.AuditHandler,
	tokens *auth.Manager,
	authRepo *repository.AuthRepository,
	auditRepo *repository.AuditRepository,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.Login(w, r)
	})
	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		authHandler.Refresh(w, r)
	})
	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		authHandler.Logout(w, r)
	})
	mux.HandleFunc("/auth/me", middleware.Authenticate(tokens, authRepo, authHandler.Me))
	protect := func(next http.HandlerFunc, roles ...string) http.HandlerFunc {
		return middleware.Authenticate(tokens, authRepo, middleware.Audit(auditRepo, middleware.Authorize(next, roles...)))
	}
	mux.HandleFunc("/users", protect(authHandler.Users, "admin"))
	mux.HandleFunc("/users/", protect(authHandler.UserDetail, "admin"))
	mux.HandleFunc("/audit-logs", protect(auditHandler.List, "admin"))
	mux.HandleFunc("/items", func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		switch r.Method {

		case http.MethodGet:
			protect(itemHandler.FindAll, "admin", "cashier", "viewer")(w, r)

		case http.MethodPost:
			protect(itemHandler.Create, "admin", "cashier")(w, r)

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
			protect(itemHandler.FindByID, "admin", "cashier", "viewer")(w, r)
		case http.MethodPut:
			protect(itemHandler.Update, "admin", "cashier")(w, r)
		case http.MethodDelete:
			protect(itemHandler.Delete, "admin", "cashier")(w, r)
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
			protect(supplierHandler.FindAll, "admin", "cashier", "viewer")(w, r)
		case http.MethodPost:
			protect(supplierHandler.Create, "admin", "cashier")(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/suppliers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			protect(supplierHandler.FindByID, "admin", "cashier", "viewer")(w, r)
		case http.MethodPut:
			protect(supplierHandler.Update, "admin", "cashier")(w, r)
		case http.MethodDelete:
			protect(supplierHandler.Delete, "admin", "cashier")(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/dashboard", protect(cooperativeHandler.Dashboard, "admin", "cashier", "viewer"))
	mux.HandleFunc("/masters/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			protect(cooperativeHandler.Masters, "admin", "cashier", "viewer")(w, r)
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/masters/brands" {
			protect(cooperativeHandler.Masters, "admin", "cashier")(w, r)
			return
		}
		protect(cooperativeHandler.Masters, "admin")(w, r)
	})
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			protect(cooperativeHandler.Customers, "admin", "cashier", "viewer")(w, r)
			return
		}
		protect(cooperativeHandler.Customers, "admin", "cashier")(w, r)
	})
	mux.HandleFunc("/customers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			protect(cooperativeHandler.CustomerDetail, "admin", "cashier", "viewer")(w, r)
			return
		}
		protect(cooperativeHandler.CustomerDetail, "admin", "cashier")(w, r)
	})
	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			protect(cooperativeHandler.Transactions, "admin", "cashier", "viewer")(w, r)
			return
		}
		protect(cooperativeHandler.Transactions, "admin", "cashier")(w, r)
	})
	mux.HandleFunc("/transactions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			protect(cooperativeHandler.VoidTransaction, "admin", "cashier", "viewer")(w, r)
			return
		}
		protect(cooperativeHandler.VoidTransaction, "admin", "cashier")(w, r)
	})
	mux.HandleFunc("/debts", protect(cooperativeHandler.Debts, "admin", "cashier", "viewer"))
	mux.HandleFunc("/debts/", protect(cooperativeHandler.PayDebt, "admin", "cashier"))

	return mux
}
