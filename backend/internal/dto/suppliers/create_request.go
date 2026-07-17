package dto

type CreateSupplierRequest struct {
	Code string `json:"code" validate:"required,min=2,max=50"`

	Name string `json:"name" validate:"required,min=3,max=100"`

	Phone string `json:"phone" validate:"required,min=8,max=20,numeric"`

	Address string `json:"address" validate:"required"`
}
