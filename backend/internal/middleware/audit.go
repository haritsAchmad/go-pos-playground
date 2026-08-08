package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go-pos-playground/internal/entity"
)

type AuditStore interface {
	Record(context.Context, entity.AuditEntry) error
}

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *auditResponseWriter) Write(value []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(value)
}

func Audit(store AuditStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paymentStatusRead := r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/payments/")
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete && !paymentStatusRead {
			next(w, r)
			return
		}
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			next(w, r)
			return
		}
		writer := &auditResponseWriter{ResponseWriter: w}
		next(writer, r)
		if paymentStatusRead && writer.Header().Get("X-Audit-Action") == "" {
			return
		}
		statusCode := writer.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			return
		}
		entityType, entityID := auditTarget(r.URL.Path)
		if targets := writer.Header().Get("X-Audit-Targets"); targets != "" {
			entityID = targets
		}
		action := auditAction(r.Method, r.URL.Path)
		if headerAction := writer.Header().Get("X-Audit-Action"); headerAction != "" {
			action = headerAction
		}
		_ = store.Record(r.Context(), entity.AuditEntry{
			UserID: userID, UserName: claims.Name, UserEmail: claims.Email,
			Action:     action,
			EntityType: entityType, EntityID: entityID,
			Method: r.Method, Path: r.URL.Path, StatusCode: statusCode,
			IPAddress: requestIP(r), UserAgent: r.UserAgent(),
		})
	}
}

func auditAction(method, path string) string {
	if strings.HasPrefix(path, "/payments/") && strings.HasSuffix(path, "/simulate") {
		return "PAYMENT_SIMULATION"
	}
	if strings.HasPrefix(path, "/payments/") && strings.HasSuffix(path, "/cancel") {
		return "PAYMENT_CANCEL"
	}
	if strings.HasSuffix(path, "/reverse") {
		return "PAYMENT_REVERSAL"
	}
	if strings.HasPrefix(path, "/bulk/") {
		if strings.HasSuffix(path, "/restore") {
			return "BULK_RESTORE"
		}
		if strings.HasSuffix(path, "/settle") {
			return "BULK_SETTLEMENT"
		}
		if strings.HasSuffix(path, "/reset-stock") {
			return "BULK_STOCK_RESET"
		}
		return "BULK_DELETE"
	}
	if strings.HasSuffix(path, "/void") {
		return "VOID"
	}
	if strings.HasSuffix(path, "/payments") {
		return "PAYMENT"
	}
	return map[string]string{
		http.MethodPost:   "CREATE",
		http.MethodPut:    "UPDATE",
		http.MethodDelete: "DELETE",
	}[method]
}

func auditTarget(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "unknown", ""
	}
	entityType := parts[0]
	if entityType == "bulk" && len(parts) > 1 {
		entityType = parts[1]
	}
	if entityType == "masters" && len(parts) > 1 {
		entityType += "/" + parts[1]
	}
	for _, part := range parts[1:] {
		if _, err := strconv.ParseInt(part, 10, 64); err == nil {
			return entityType, part
		}
	}
	return entityType, ""
}

func requestIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
