package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct { Subject string `json:"sub"`; Name string `json:"name"`; Email string `json:"email"`; Role string `json:"role"`; Issuer string `json:"iss"`; IssuedAt int64 `json:"iat"`; ExpiresAt int64 `json:"exp"` }

type Manager struct { secret []byte; issuer string; ttl time.Duration }

func NewManager(secret, issuer, expiryMinutes string) (*Manager, error) {
	if len(secret) < 32 { return nil, errors.New("JWT_SECRET must contain at least 32 characters") }
	if issuer == "" { issuer = "go-pos-playground" }
	minutes := 480
	if expiryMinutes != "" { v, err := strconv.Atoi(expiryMinutes); if err != nil || v <= 0 { return nil, errors.New("JWT_EXPIRY_MINUTES must be a positive integer") }; minutes = v }
	return &Manager{[]byte(secret), issuer, time.Duration(minutes)*time.Minute}, nil
}

func (m *Manager) Issue(id int64, name, email, role string) (string, int64, error) {
	now := time.Now(); exp := now.Add(m.ttl).Unix()
	header, _ := json.Marshal(map[string]string{"alg":"HS256","typ":"JWT"})
	payload, err := json.Marshal(Claims{fmt.Sprint(id),name,email,role,m.issuer,now.Unix(),exp}); if err != nil { return "",0,err }
	unsigned := encode(header)+"."+encode(payload)
	return unsigned+"."+m.sign(unsigned), exp, nil
}

func (m *Manager) Parse(token string) (Claims, error) {
	var claims Claims
	parts := strings.Split(token, "."); if len(parts) != 3 { return claims, errors.New("invalid token") }
	unsigned := parts[0]+"."+parts[1]
	if !hmac.Equal([]byte(m.sign(unsigned)), []byte(parts[2])) { return claims, errors.New("invalid token signature") }
	payload, err := base64.RawURLEncoding.DecodeString(parts[1]); if err != nil { return claims, errors.New("invalid token") }
	if json.Unmarshal(payload, &claims) != nil || claims.Issuer != m.issuer || claims.Subject == "" || time.Now().Unix() >= claims.ExpiresAt { return Claims{}, errors.New("token expired or invalid") }
	return claims, nil
}

func encode(v []byte) string { return base64.RawURLEncoding.EncodeToString(v) }
func (m *Manager) sign(v string) string { h := hmac.New(sha256.New,m.secret); h.Write([]byte(v)); return encode(h.Sum(nil)) }
