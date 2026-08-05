package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/middleware"
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

func (h *CooperativeHandler) DeletedCustomers(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.DeletedCustomers(r.Context())
	if err != nil {
		response.Error(w, 500, "failed to get deleted customers")
		return
	}
	response.Success(w, 200, "deleted customers fetched", data)
}

func (h *CooperativeHandler) RestoreCustomer(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/customers/"), "/restore")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		response.Error(w, 400, "ID pelanggan tidak valid")
		return
	}
	if err := h.repo.RestoreCustomer(r.Context(), id); err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	response.Success(w, 200, "pelanggan berhasil dipulihkan", nil)
}

func (h *CooperativeHandler) BulkSoftDeleteRestore(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.Method != http.MethodPost || len(parts) != 3 || parts[0] != "bulk" ||
		(parts[2] != "delete" && parts[2] != "restore" && parts[2] != "settle" && parts[2] != "reset-stock") {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req entity.BulkRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || len(req.IDs) == 0 || len(req.IDs) > 100 {
		response.Error(w, http.StatusBadRequest, "pilih 1 sampai 100 data")
		return
	}
	seen := make(map[int64]bool, len(req.IDs))
	ids := make([]int64, 0, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		response.Error(w, http.StatusBadRequest, "ID data tidak valid")
		return
	}
	targets := make([]string, len(ids))
	for i, id := range ids {
		targets[i] = strconv.FormatInt(id, 10)
	}
	w.Header().Set("X-Audit-Targets", strings.Join(targets, ","))
	var result entity.BulkResult
	if parts[1] == "debts" && parts[2] == "settle" {
		result = h.repo.BulkSettleDebts(r.Context(), ids)
	} else if parts[1] == "items" && parts[2] == "reset-stock" {
		result = h.repo.BulkResetStock(r.Context(), ids)
	} else if parts[2] == "delete" {
		result = h.repo.BulkSoftDelete(r.Context(), parts[1], ids)
	} else {
		result = h.repo.BulkRestore(r.Context(), parts[1], ids)
	}
	if len(result.Results) == 0 {
		response.Error(w, http.StatusBadRequest, "jenis data tidak valid")
		return
	}
	response.Success(w, http.StatusOK, "bulk action completed", result)
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
		req.IdempotencyKey = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if req.IdempotencyKey != "" && (len(req.IdempotencyKey) < 8 || len(req.IdempotencyKey) > 128 || !regexp.MustCompile(`^[A-Za-z0-9._:-]+$`).MatchString(req.IdempotencyKey)) {
			response.Error(w, http.StatusBadRequest, "Idempotency-Key must be 8-128 safe characters")
			return
		}
		data, err := h.repo.CreateTransaction(r.Context(), req)
		if err != nil {
			if errors.Is(err, repository.ErrIdempotencyConflict) {
				response.Error(w, http.StatusConflict, err.Error())
				return
			}
			response.Error(w, 400, err.Error())
			return
		}
		if data.IdempotencyReplay {
			w.Header().Set("Idempotency-Replayed", "true")
			response.Success(w, http.StatusOK, "transaction already created", data)
			return
		}
		response.Success(w, 201, "transaction created", data)
		return
	}
	response.Error(w, 405, "method not allowed")
}

func (h *CooperativeHandler) SimulatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/simulate") {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	part := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/payments/"), "/simulate")
	id, err := strconv.ParseInt(strings.Trim(part, "/"), 10, 64)
	if err != nil || id < 1 {
		response.Error(w, http.StatusBadRequest, "invalid payment ID")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		response.Error(w, http.StatusBadRequest, "invalid payment data")
		return
	}
	payment, err := h.repo.SetDummyPaymentStatus(r.Context(), id, strings.ToUpper(req.Status))
	if err != nil {
		if errors.Is(err, repository.ErrPaymentExpired) {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(w, http.StatusOK, "dummy payment updated", payment)
}

func (h *CooperativeHandler) PaymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	part := strings.Trim(strings.TrimPrefix(r.URL.Path, "/payments/"), "/")
	id, err := strconv.ParseInt(part, 10, 64)
	if err != nil || id < 1 {
		response.Error(w, http.StatusBadRequest, "invalid payment ID")
		return
	}
	payment, err := h.repo.DummyPayment(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(w, http.StatusOK, "payment status fetched", payment)
}

func (h *CooperativeHandler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/cancel") {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	part := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/payments/"), "/cancel")
	id, err := strconv.ParseInt(strings.Trim(part, "/"), 10, 64)
	if err != nil || id < 1 {
		response.Error(w, http.StatusBadRequest, "invalid payment ID")
		return
	}
	payment, err := h.repo.SetDummyPaymentStatus(r.Context(), id, "FAILED")
	if err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	response.Success(w, http.StatusOK, "pending payment cancelled", payment)
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

func (h *CooperativeHandler) DebtPayments(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.Method != http.MethodGet || len(parts) != 3 || parts[0] != "debts" || parts[2] != "payments" {
		response.Error(w, http.StatusNotFound, "payment history path not found")
		return
	}
	debtID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid debt id")
		return
	}
	payments, err := h.repo.DebtPayments(r.Context(), debtID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get payment history")
		return
	}
	response.Success(w, http.StatusOK, "payment history fetched", payments)
}

func (h *CooperativeHandler) ReverseDebtPayment(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.Method != http.MethodPost || len(parts) != 5 || parts[0] != "debts" || parts[2] != "payments" || parts[4] != "reverse" {
		response.Error(w, http.StatusNotFound, "payment reversal path not found")
		return
	}
	debtID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid debt id")
		return
	}
	paymentID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid payment id")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		response.Error(w, http.StatusBadRequest, "invalid reversal data")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid user identity")
		return
	}
	if err := h.repo.ReverseDebtPayment(r.Context(), debtID, paymentID, userID, claims.Name, req.Reason); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(w, http.StatusOK, "payment reversed", nil)
}

func (h *CooperativeHandler) StockMovements(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.Method != http.MethodGet || len(parts) != 3 || parts[0] != "items" || parts[2] != "stock-movements" {
		response.Error(w, http.StatusNotFound, "stock movement path not found")
		return
	}
	itemID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid item id")
		return
	}
	movements, err := h.repo.StockMovements(r.Context(), itemID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get stock movements")
		return
	}
	response.Success(w, http.StatusOK, "stock movements fetched", movements)
}
