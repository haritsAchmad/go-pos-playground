package entity

import "time"

type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserEmail  string    `json:"user_email"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditEntry struct {
	UserID     int64
	UserName   string
	UserEmail  string
	Action     string
	EntityType string
	EntityID   string
	Method     string
	Path       string
	StatusCode int
	IPAddress  string
	UserAgent  string
}
