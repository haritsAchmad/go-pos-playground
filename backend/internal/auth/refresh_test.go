package auth

import "testing"

func TestRefreshManagerIssuesUniqueHashedTokens(t *testing.T) {
	manager, err := NewRefreshManager("3")
	if err != nil {
		t.Fatal(err)
	}
	first, firstHash, expiresAt, err := manager.Issue()
	if err != nil {
		t.Fatal(err)
	}
	second, secondHash, _, err := manager.Issue()
	if err != nil {
		t.Fatal(err)
	}
	if first == firstHash || first == second || firstHash == secondHash {
		t.Fatal("refresh tokens must be unique and stored only as hashes")
	}
	if HashRefreshToken(first) != firstHash || manager.MaxAge() != 3*24*60*60 || expiresAt.IsZero() {
		t.Fatal("unexpected refresh token metadata")
	}
}

func TestRefreshManagerRejectsInvalidExpiry(t *testing.T) {
	if _, err := NewRefreshManager("0"); err == nil {
		t.Fatal("expected invalid expiry error")
	}
}
