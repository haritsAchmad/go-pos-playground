package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"time"
)

const DefaultRefreshDays = 7

type RefreshManager struct {
	ttl time.Duration
}

func NewRefreshManager(expiryDays string) (*RefreshManager, error) {
	days := DefaultRefreshDays
	if expiryDays != "" {
		value, err := strconv.Atoi(expiryDays)
		if err != nil || value < 1 {
			return nil, errors.New("REFRESH_TOKEN_EXPIRY_DAYS must be a positive integer")
		}
		days = value
	}
	return &RefreshManager{ttl: time.Duration(days) * 24 * time.Hour}, nil
}

func (m *RefreshManager) Issue() (raw, hash string, expiresAt time.Time, err error) {
	value := make([]byte, 32)
	if _, err = rand.Read(value); err != nil {
		return "", "", time.Time{}, err
	}
	raw = base64.RawURLEncoding.EncodeToString(value)
	return raw, HashRefreshToken(raw), time.Now().Add(m.ttl), nil
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (m *RefreshManager) MaxAge() int {
	return int(m.ttl.Seconds())
}
