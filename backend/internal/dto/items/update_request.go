package dto

type UpdateItemRequest struct {
	SupplierID      *int   `json:"supplier_id"`
	SKU             string `json:"sku" validate:"required,min=2,max=50"`
	CategoryID      *int   `json:"category_id"`
	BrandID         *int   `json:"brand_id"`
	UnitID          *int   `json:"unit_id"`
	BaseUnitID      *int   `json:"base_unit_id"`
	UnitsPerPackage int    `json:"units_per_package" validate:"gte=1"`
	AllowRetail     bool   `json:"allow_retail"`

	Name string `json:"name" validate:"required,min=3,max=100"`

	Description string `json:"description" validate:"max=500"`

	Stock int `json:"stock" validate:"gte=0"`

	Price       int64 `json:"price" validate:"gte=0"`
	Cost        int64 `json:"cost" validate:"gte=0"`
	RetailPrice int64 `json:"retail_price" validate:"gte=0"`
	RetailCost  int64 `json:"retail_cost" validate:"gte=0"`
}
