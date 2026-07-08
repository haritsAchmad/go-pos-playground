package handler

import (
	"encoding/json"
	"net/http"

	dto "go-inventory-playground/internal/dto/items"
	"go-inventory-playground/internal/repository"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ItemHandler struct {
	itemRepo *repository.ItemRepository
}

func NewItemHandler(itemRepo *repository.ItemRepository) *ItemHandler {
	return &ItemHandler{
		itemRepo: itemRepo,
	}
}

func (h *ItemHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	items, err := h.itemRepo.FindAll(r.Context())
	if err != nil {
		http.Error(w, `{"message":"failed to get items"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(items)
}

func (h *ItemHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {

	var req dto.CreateItemRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = validate.Struct(req)
	if err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	err = h.itemRepo.Create(
		r.Context(),
		req,
	)

	if err != nil {
		http.Error(w, "failed to create item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(
		map[string]any{
			"message": "item created successfully",
		},
	)
}
