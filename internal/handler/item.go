package handler

import (
	"encoding/json"
	"net/http"

	dto "go-inventory-playground/internal/dto/items"
	"go-inventory-playground/internal/pkg/response"
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
	items, err := h.itemRepo.FindAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get items")
		return
	}

	response.Success(w, http.StatusOK, "items fetched successfully", items)
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateItemRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = validate.Struct(req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "validation failed")
		return
	}

	err = h.itemRepo.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	response.Success(w, http.StatusCreated, "item created successfully", nil)
}
