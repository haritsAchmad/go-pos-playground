package auth

import "testing"

func TestIssueAndParse(t *testing.T) {
	m, err := NewManager("12345678901234567890123456789012", "test", "5")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := m.Issue(7, 19, "Kasir", "kasir@example.com", "cashier")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "7" || claims.SessionID != 19 || claims.Role != "cashier" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestIssueRequiresSession(t *testing.T) {
	m, _ := NewManager("12345678901234567890123456789012", "test", "5")
	if _, _, err := m.Issue(7, 0, "Kasir", "kasir@example.com", "cashier"); err == nil {
		t.Fatal("expected session ID error")
	}
}

func TestRejectsWeakSecret(t *testing.T) {
	if _, err := NewManager("short", "test", "5"); err == nil {
		t.Fatal("expected weak secret error")
	}
}
