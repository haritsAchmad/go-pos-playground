package handler

import (
	"net/http"

	"go-pos-playground/internal/pkg/pagination"
	"go-pos-playground/internal/pkg/response"
)

func paginationParams(w http.ResponseWriter, r *http.Request) (pagination.Params, bool, bool) {
	params, enabled, err := pagination.Parse(r.URL.Query())
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return pagination.Params{}, enabled, false
	}
	return params, enabled, true
}
