package handler

import (
	"encoding/json"
	"net/http"

	"go-inventory-playground/internal/repository"
)

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
