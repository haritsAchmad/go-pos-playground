package entity

import "time"

type Items struct {
	ID              int       `json:"id"`
	SupplierID      *int      `json:"supplier_id"`
	SKU             string    `json:"sku"`
	CategoryID      *int      `json:"category_id"`
	CategoryName    *string   `json:"category_name"`
	BrandID         *int      `json:"brand_id"`
	BrandName       *string   `json:"brand_name"`
	UnitID          *int      `json:"unit_id"`
	UnitName        *string   `json:"unit_name"`
	BaseUnitID      *int      `json:"base_unit_id"`
	BaseUnitName    *string   `json:"base_unit_name"`
	UnitsPerPackage int       `json:"units_per_package"`
	AllowRetail     bool      `json:"allow_retail"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Stock           int       `json:"stock"`
	ReservedStock   int       `json:"reserved_stock"`
	AvailableStock  int       `json:"available_stock"`
	Price           int64     `json:"price"`
	Cost            int64     `json:"cost"`
	RetailPrice     int64     `json:"retail_price"`
	RetailCost      int64     `json:"retail_cost"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
