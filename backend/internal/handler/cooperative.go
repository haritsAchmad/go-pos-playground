package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"
)

type CooperativeHandler struct {
	repo *repository.CooperativeRepository
}

func NewCooperativeHandler(repo *repository.CooperativeRepository) *CooperativeHandler {
	return &CooperativeHandler{repo: repo}
}

func (h *CooperativeHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year < 2000 || year > 2100 {
		year = time.Now().Year()
	}
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}
	data, err := h.repo.Dashboard(r.Context(), year, month)
	if err != nil {
		response.Error(w, 500, "failed to load dashboard")
		return
	}
	response.Success(w, 200, "dashboard fetched", data)
}

func (h *CooperativeHandler) Masters(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/masters/"), "/"), "/")
	table := parts[0]
	var id int64
	if len(parts) > 1 {
		id, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	if r.Method == http.MethodGet {
		data, err := h.repo.Masters(r.Context(), table)
		if err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		response.Success(w, 200, "master data fetched", data)
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Name string `json:"name" validate:"required,min=2,max=100"`
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || validate.Struct(req) != nil {
			response.Error(w, 400, "invalid master data")
			return
		}
		if err := h.repo.CreateMaster(r.Context(), table, strings.TrimSpace(req.Name)); err != nil {
			response.Error(w, 400, "failed to create master data")
			return
		}
		response.Success(w, 201, "master data created", nil)
		return
	}
	if r.Method == http.MethodPut && id > 0 {
		var req struct {
			Name string `json:"name"`
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || len(strings.TrimSpace(req.Name)) < 2 {
			response.Error(w, 400, "nama minimal 2 karakter")
			return
		}
		if err := h.repo.UpdateMaster(r.Context(), table, id, strings.TrimSpace(req.Name)); err != nil {
			response.Error(w, 400, "data masih dipakai atau tidak ditemukan")
			return
		}
		response.Success(w, 200, "data berhasil diubah", nil)
		return
	}
	if r.Method == http.MethodDelete && id > 0 {
		if err := h.repo.DeleteMaster(r.Context(), table, id); err != nil {
			response.Error(w, 400, "data tidak dapat dihapus karena masih dipakai")
			return
		}
		response.Success(w, 200, "data berhasil dihapus", nil)
		return
	}
	response.Error(w, 405, "method not allowed")
}

func (h *CooperativeHandler) Customers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query, err := listquery.Parse(r.URL.Query(), listquery.Config{
			DefaultSort: "name",
			Sorts: map[string]bool{
				"id": true, "code": true, "name": true, "customer_type": true, "created_at": true,
			},
			Filters: map[string]bool{"customer_type": true},
		})
		if err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if customerType := query.Values["customer_type"]; customerType != "" && customerType != "MEMBER" && customerType != "NON_MEMBER" {
			response.Error(w, http.StatusBadRequest, "customer_type must be MEMBER or NON_MEMBER")
			return
		}
		params, paginated, ok := paginationParams(w, r)
		if !ok {
			return
		}
		if paginated {
			data, err := h.repo.CustomersPageQuery(r.Context(), params, query)
			if err != nil {
				response.Error(w, 500, "failed to get customers")
				return
			}
			response.Success(w, 200, "customers fetched", data)
			return
		}
		data, err := h.repo.CustomersQuery(r.Context(), query)
		if err != nil {
			response.Error(w, 500, "failed to get customers")
			return
		}
		response.Success(w, 200, "customers fetched", data)
		return
	}
	if r.Method == http.MethodPost {
		var req entity.Customer
		if json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.Code) == "" || len(strings.TrimSpace(req.Name)) < 3 || (req.Phone != "" && !regexp.MustCompile(`^[0-9]{8,20}$`).MatchString(req.Phone)) {
			response.Error(w, 400, "kode dan nama wajib diisi; telepon harus 8-20 digit angka")
			return
		}
		if req.CustomerType == "" {
			req.CustomerType = "MEMBER"
		}
		if req.CustomerType != "MEMBER" && req.CustomerType != "NON_MEMBER" {
			response.Error(w, 400, "invalid customer type")
			return
		}
		if err := h.repo.CreateCustomer(r.Context(), req); err != nil {
			response.Error(w, 400, "failed to create customer")
			return
		}
		response.Success(w, 201, "customer created", nil)
		return
	}
	response.Error(w, 405, "method not allowed")
}

func (h *CooperativeHandler) CustomerDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/customers/"), 10, 64)
	if err != nil {
		response.Error(w, 400, "ID pelanggan tidak valid")
		return
	}
	if r.Method == http.MethodDelete {
		if err := h.repo.DeleteCustomer(r.Context(), id); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		response.Success(w, 200, "pelanggan berhasil dihapus", nil)
		return
	}
	if r.Method == http.MethodPut {
		var req entity.Customer
		if json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.Code) == "" || len(strings.TrimSpace(req.Name)) < 3 || (req.Phone != "" && !regexp.MustCompile(`^[0-9]{8,20}$`).MatchString(req.Phone)) {
			response.Error(w, 400, "kode dan nama wajib diisi; telepon harus 8-20 digit angka")
			return
		}
		if err := h.repo.UpdateCustomer(r.Context(), id, req); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		response.Success(w, 200, "pelanggan berhasil diubah", nil)
		return
	}
	response.Error(w, 405, "method not allowed")
}

func (h *CooperativeHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query, err := listquery.Parse(r.URL.Query(), listquery.Config{
			DefaultSort: "transaction_date",
			Sorts: map[string]bool{
				"id": true, "invoice_no": true, "transaction_date": true,
				"grand_total": true, "payment_status": true, "status": true,
			},
			Filters: map[string]bool{
				"payment_status": true, "status": true, "date_from": true, "date_to": true,
			},
		})
		if err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		kind := r.URL.Query().Get("type")
		if kind != "" && kind != "SALE" && kind != "PURCHASE" {
			response.Error(w, http.StatusBadRequest, "type must be SALE or PURCHASE")
			return
		}
		if value := query.Values["payment_status"]; value != "" && value != "PAID" && value != "UNPAID" && value != "PARTIAL" {
			response.Error(w, http.StatusBadRequest, "payment_status must be PAID, UNPAID, or PARTIAL")
			return
		}
		if value := query.Values["status"]; value != "" && value != "ACTIVE" && value != "VOID" {
			response.Error(w, http.StatusBadRequest, "status must be ACTIVE or VOID")
			return
		}
		for _, key := range []string{"date_from", "date_to"} {
			if value := query.Values[key]; value != "" {
				if _, err := time.Parse("2006-01-02", value); err != nil {
					response.Error(w, http.StatusBadRequest, key+" must use YYYY-MM-DD")
					return
				}
			}
		}
		if from, to := query.Values["date_from"], query.Values["date_to"]; from != "" && to != "" && from > to {
			response.Error(w, http.StatusBadRequest, "date_from must not exceed date_to")
			return
		}
		params, paginated, ok := paginationParams(w, r)
		if !ok {
			return
		}
		if paginated {
			data, err := h.repo.TransactionsPageQuery(r.Context(), kind, params, query)
			if err != nil {
				response.Error(w, 500, "failed to get transactions")
				return
			}
			response.Success(w, 200, "transactions fetched", data)
			return
		}
		data, err := h.repo.TransactionsQuery(r.Context(), kind, query)
		if err != nil {
			response.Error(w, 500, "failed to get transactions")
			return
		}
		response.Success(w, 200, "transactions fetched", data)
		return
	}
	if r.Method == http.MethodPost {
		var req entity.CreateTransactionRequest
		if json.NewDecoder(r.Body).Decode(&req) != nil || validate.Struct(req) != nil {
			response.Error(w, 400, "invalid transaction data")
			return
		}
		data, err := h.repo.CreateTransaction(r.Context(), req)
		if err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		response.Success(w, 201, "transaction created", data)
		return
	}
	response.Error(w, 405, "method not allowed")
}

func (h *CooperativeHandler) VoidTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut && !strings.HasSuffix(r.URL.Path, "/void") {
		part := strings.Trim(strings.TrimPrefix(r.URL.Path, "/transactions/"), "/")
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			response.Error(w, 400, "ID transaksi tidak valid")
			return
		}
		var req entity.CreateTransactionRequest
		if json.NewDecoder(r.Body).Decode(&req) != nil || validate.Struct(req) != nil {
			response.Error(w, 400, "invalid transaction data")
			return
		}
		data, err := h.repo.UpdateTransaction(r.Context(), id, req)
		if err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		response.Success(w, 200, "transaksi berhasil diubah dan stok telah disesuaikan", data)
		return
	}
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/void") {
		response.Error(w, 405, "method not allowed")
		return
	}
	part := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/transactions/"), "/void")
	id, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		response.Error(w, 400, "ID transaksi tidak valid")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || len(strings.TrimSpace(req.Reason)) < 5 {
		response.Error(w, 400, "alasan pembatalan minimal 5 karakter")
		return
	}
	if err := h.repo.VoidTransaction(r.Context(), id, strings.TrimSpace(req.Reason)); err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.Success(w, 200, "transaksi berhasil dibatalkan dan stok telah disesuaikan", nil)
}

func (h *CooperativeHandler) Debts(w http.ResponseWriter, r *http.Request) {
	query, err := listquery.Parse(r.URL.Query(), listquery.Config{
		DefaultSort: "created_at",
		Sorts: map[string]bool{
			"id": true, "invoice_no": true, "customer_name": true,
			"original_amount": true, "remaining_amount": true, "status": true, "created_at": true,
		},
		Filters: map[string]bool{"status": true, "min_remaining": true, "max_remaining": true},
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if status := query.Values["status"]; status != "" && status != "OPEN" && status != "PAID" {
		response.Error(w, http.StatusBadRequest, "status must be OPEN or PAID")
		return
	}
	minRemaining, hasMin, err := query.NonNegativeInt("min_remaining")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	maxRemaining, hasMax, err := query.NonNegativeInt("max_remaining")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if hasMin && hasMax && minRemaining > maxRemaining {
		response.Error(w, http.StatusBadRequest, "min_remaining must not exceed max_remaining")
		return
	}
	params, paginated, ok := paginationParams(w, r)
	if !ok {
		return
	}
	if paginated {
		data, err := h.repo.DebtsPageQuery(r.Context(), params, query)
		if err != nil {
			response.Error(w, 500, "failed to get debts")
			return
		}
		response.Success(w, 200, "debts fetched", data)
		return
	}
	data, err := h.repo.DebtsQuery(r.Context(), query)
	if err != nil {
		response.Error(w, 500, "failed to get debts")
		return
	}
	response.Success(w, 200, "debts fetched", data)
}

func (h *CooperativeHandler) PayDebt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, 405, "method not allowed")
		return
	}
	part := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/debts/"), "/payments")
	id, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		response.Error(w, 400, "invalid debt id")
		return
	}
	var req struct {
		Amount int64  `json:"amount"`
		Notes  string `json:"notes"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		response.Error(w, 400, "invalid payment data")
		return
	}
	if err := h.repo.PayDebt(r.Context(), id, req.Amount, req.Notes); err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.Success(w, 201, "debt payment recorded", nil)
}
