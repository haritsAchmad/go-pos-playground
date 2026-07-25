package handler

import (
	"net/http"

	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"
)

type AuditHandler struct {
	repo *repository.AuditRepository
}

func NewAuditHandler(repo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	query, err := listquery.Parse(r.URL.Query(), listquery.Config{
		DefaultSort: "created_at",
		Sorts: map[string]bool{
			"id": true, "created_at": true, "user_name": true,
			"action": true, "entity_type": true, "status_code": true,
		},
		Filters: map[string]bool{"user_id": true, "action": true, "entity_type": true},
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, _, err := query.PositiveInt("user_id"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	params, paginated, ok := paginationParams(w, r)
	if !ok {
		return
	}
	if paginated {
		result, err := h.repo.ListPage(r.Context(), params, query)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to get audit logs")
			return
		}
		response.Success(w, http.StatusOK, "audit logs fetched", result)
		return
	}
	result, err := h.repo.List(r.Context(), query)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get audit logs")
		return
	}
	response.Success(w, http.StatusOK, "audit logs fetched", result)
}
