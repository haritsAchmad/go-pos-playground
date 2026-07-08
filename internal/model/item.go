package model

import "time"

type Item struct {
	ID          int
	Name        string
	Description string
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
