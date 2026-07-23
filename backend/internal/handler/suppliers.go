package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	dto "go-pos-playground/internal/dto/suppliers"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"
)

type SupplierHandler struct {
	supplierRepo *repository.SupplierRepository
}

func NewSupplierHandler(supplierRepo *repository.SupplierRepository) *SupplierHandler {
	return &SupplierHandler{
		supplierRepo: supplierRepo,
	}
}

func (h *SupplierHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	params, paginated, ok := paginationParams(w, r)
	if !ok {
		return
	}
	if paginated {
		suppliers, err := h.supplierRepo.FindPage(r.Context(), params)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to get suppliers")
			return
		}
		response.Success(w, http.StatusOK, "suppliers fetched successfully", suppliers)
		return
	}
	suppliers, err := h.supplierRepo.FindAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get suppliers")
		return
	}

	response.Success(w, http.StatusOK, "suppliers fetched successfully", suppliers)
}

func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSupplierRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = validate.Struct(req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "kode, nama, alamat, dan nomor telepon 8-20 digit wajib diisi")
		return
	}

	err = h.supplierRepo.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create supplier")
		return
	}

	response.Success(w, http.StatusCreated, "supplier created successfully", nil)
}

func (h *SupplierHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid supplier id")
		return
	}

	supplier, err := h.supplierRepo.FindByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get supplier")
		return
	}

	if supplier == nil {
		response.Error(w, http.StatusNotFound, "supplier not found")
		return
	}

	response.Success(w, http.StatusOK, "success", supplier)
}

func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid supplier id")
		return
	}

	var req dto.UpdateSupplierRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = validate.Struct(req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "kode, nama, alamat, dan nomor telepon 8-20 digit wajib diisi")
		return
	}

	err = h.supplierRepo.Update(r.Context(), id, req)
	if errors.Is(err, repository.ErrSupplierNotFound) {
		response.Error(w, http.StatusNotFound, "supplier not found")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update supplier")
		return
	}

	response.Success(w, http.StatusOK, "supplier updated successfully", nil)
}

func (h *SupplierHandler) Delete(
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
			"invalid supplier id",
		)

		return
	}

	err = h.supplierRepo.Delete(
		r.Context(),
		id,
	)

	if errors.Is(
		err,
		repository.ErrSupplierNotFound,
	) {

		response.Error(
			w,
			http.StatusNotFound,
			"supplier not found",
		)

		return
	}

	if err != nil {

		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to delete supplier",
		)

		return
	}

	response.Success(
		w,
		http.StatusOK,
		"supplier deleted successfully",
		nil,
	)
}
