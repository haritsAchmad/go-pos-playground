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

func TestAuditNamesPaymentReversal(t *testing.T) {
	store := &auditStoreStub{}
	next := Audit(store, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	request := httptest.NewRequest(http.MethodPost, "/debts/12/payments/34/reverse", nil)
	ctx := context.WithValue(request.Context(), claimsKey, auth.Claims{Subject: "7"})
	next(httptest.NewRecorder(), request.WithContext(ctx))
	if len(store.entries) != 1 || store.entries[0].Action != "PAYMENT_REVERSAL" || store.entries[0].EntityID != "12" {
		t.Fatalf("unexpected reversal audit entry: %+v", store.entries)
	}
}

func TestAuditNamesDummyPaymentTransitions(t *testing.T) {
	tests := []struct{ path, action string }{
		{"/payments/12/simulate", "PAYMENT_SIMULATION"},
		{"/payments/12/cancel", "PAYMENT_CANCEL"},
	}
	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			store := &auditStoreStub{}
			next := Audit(store, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			request := httptest.NewRequest(http.MethodPost, test.path, nil)
			ctx := context.WithValue(request.Context(), claimsKey, auth.Claims{Subject: "7"})
			next(httptest.NewRecorder(), request.WithContext(ctx))
			if len(store.entries) != 1 || store.entries[0].Action != test.action || store.entries[0].EntityID != "12" {
				t.Fatalf("unexpected payment audit entry: %+v", store.entries)
			}
		})
	}
}

func TestAuditRecordsOnlyPaymentReadThatExpires(t *testing.T) {
	store := &auditStoreStub{}
	next := Audit(store, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Audit-Action", "PAYMENT_EXPIRY")
		w.WriteHeader(http.StatusOK)
	})
	request := httptest.NewRequest(http.MethodGet, "/payments/12", nil)
	ctx := context.WithValue(request.Context(), claimsKey, auth.Claims{Subject: "7"})
	next(httptest.NewRecorder(), request.WithContext(ctx))
	if len(store.entries) != 1 || store.entries[0].Action != "PAYMENT_EXPIRY" {
		t.Fatalf("unexpected expiry audit entry: %+v", store.entries)
	}

	store.entries = nil
	next = Audit(store, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	next(httptest.NewRecorder(), request.WithContext(ctx))
	if len(store.entries) != 0 {
		t.Fatalf("ordinary payment polling must not be audited: %+v", store.entries)
	}
}
