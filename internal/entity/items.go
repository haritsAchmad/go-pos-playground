package entity

import "time"

type Items struct {
	ID          int       `json:"id"`
	SupplierID  int       `json:"supplier_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
