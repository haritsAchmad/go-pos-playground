package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	dto "go-pos-playground/internal/dto/items"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"

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
	params, paginated, ok := paginationParams(w, r)
	if !ok {
		return
	}
	if paginated {
		items, err := h.itemRepo.FindPage(r.Context(), params)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to get items")
			return
		}
		response.Success(w, http.StatusOK, "items fetched successfully", items)
		return
	}
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
		response.Error(w, http.StatusBadRequest, "kode/SKU dan nama wajib diisi; stok serta harga tidak boleh negatif")
		return
	}

	err = h.itemRepo.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "barang gagal disimpan; pastikan kode/SKU unik dan data pilihan masih tersedia")
		return
	}

	response.Success(w, http.StatusCreated, "item created successfully", nil)
}

func (h *ItemHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.itemRepo.FindByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	if item == nil {
		response.Error(w, http.StatusNotFound, "item not found")
		return
	}

	response.Success(w, http.StatusOK, "success", item)
}

func getIDFromPath(path string) (int, error) {
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[2] == "" {
		return 0, strconv.ErrSyntax
	}

	return strconv.Atoi(parts[2])
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var req dto.UpdateItemRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = validate.Struct(req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "kode/SKU dan nama wajib diisi; stok serta harga tidak boleh negatif")
		return
	}

	err = h.itemRepo.Update(r.Context(), id, req)
	if errors.Is(err, repository.ErrItemNotFound) {
		response.Error(w, http.StatusNotFound, "item not found")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update item")
		return
	}

	response.Success(w, http.StatusOK, "item updated successfully", nil)
}

func (h *ItemHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {

	id, err := getIDFromPath(
		r.URL.Path,
	)

	if err != nil {

		response.Error(
			w,
			http.StatusBadRequest,
			"invalid item id",
		)

		return
	}

	err = h.itemRepo.Delete(
		r.Context(),
		id,
	)

	if errors.Is(
		err,
		repository.ErrItemNotFound,
	) {

		response.Error(
			w,
			http.StatusNotFound,
			"item not found",
		)

		return
	}

	if err != nil {

		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to delete item",
		)

		return
	}

	response.Success(
		w,
		http.StatusOK,
		"item deleted successfully",
		nil,
	)
}
