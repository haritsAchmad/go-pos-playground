package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-inventory-playground/internal/entity"
	"go-inventory-playground/internal/pkg/response"
	"go-inventory-playground/internal/repository"
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
		data, err := h.repo.Customers(r.Context())
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
		data, err := h.repo.Transactions(r.Context(), r.URL.Query().Get("type"))
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
	data, err := h.repo.Debts(r.Context())
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
