package entity

import "time"

type Items struct {
	ID           int       `json:"id"`
	SupplierID   *int      `json:"supplier_id"`
	SKU          string    `json:"sku"`
	CategoryID   *int      `json:"category_id"`
	CategoryName *string   `json:"category_name"`
	BrandID      *int      `json:"brand_id"`
	BrandName    *string   `json:"brand_name"`
	UnitID       *int      `json:"unit_id"`
	UnitName     *string   `json:"unit_name"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Stock        int       `json:"stock"`
	Price        int64     `json:"price"`
	Cost         int64     `json:"cost"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
