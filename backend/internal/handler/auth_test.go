package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetRefreshCookieSecurityAttributes(t *testing.T) {
	handler := &AuthHandler{}
	request := httptest.NewRequest(http.MethodPost, "http://example.test/auth/login", nil)
	request.Header.Set("X-Forwarded-Proto", "https")
	recorder := httptest.NewRecorder()

	handler.setRefreshCookie(recorder, request, "refresh-value", 3600)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != refreshCookieName || cookie.Value != "refresh-value" || cookie.Path != "/" {
		t.Fatalf("unexpected cookie identity: %+v", cookie)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge != 3600 {
		t.Fatalf("missing refresh cookie security attributes: %+v", cookie)
	}
}
