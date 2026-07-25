package entity

type User struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	Active       bool   `json:"active"`
}
