package dto

type UpdateItemRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`

	Description string `json:"description" validate:"max=500"`

	Stock int `json:"stock" validate:"gte=0"`
}
