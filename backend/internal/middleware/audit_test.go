package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pos-playground/internal/auth"
	"go-pos-playground/internal/entity"
)

type auditStoreStub struct {
	entries []entity.AuditEntry
}

func (s *auditStoreStub) Record(_ context.Context, entry entity.AuditEntry) error {
	s.entries = append(s.entries, entry)
	return nil
}

func TestAuditRecordsAuthenticatedMutation(t *testing.T) {
	store := &auditStoreStub{}
	next := Audit(store, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	request := httptest.NewRequest(http.MethodPost, "/transactions/42/void", nil)
	request.Header.Set("X-Forwarded-For", "192.0.2.10, 10.0.0.1")
	ctx := context.WithValue(request.Context(), claimsKey, auth.Claims{Subject: "7"})
	recorder := httptest.NewRecorder()

	next(recorder, request.WithContext(ctx))

	if len(store.entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(store.entries))
	}
	entry := store.entries[0]
	if entry.UserID != 7 || entry.Action != "VOID" || entry.EntityType != "transactions" ||
		entry.EntityID != "42" || entry.StatusCode != http.StatusCreated || entry.IPAddress != "192.0.2.10" {
		t.Fatalf("unexpected audit entry: %+v", entry)
	}
}

func TestAuditSkipsReads(t *testing.T) {
	store := &auditStoreStub{}
	next := Audit(store, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	next(httptest.NewRecorder(), request)
	if len(store.entries) != 0 {
		t.Fatalf("entries = %d, want 0", len(store.entries))
	}
}
