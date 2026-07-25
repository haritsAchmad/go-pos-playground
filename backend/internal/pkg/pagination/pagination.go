package pagination

import (
	"errors"
	"net/url"
	"strconv"
)

const (
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Params struct {
	Page    int
	PerPage int
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type Result[T any] struct {
	Items []T  `json:"items"`
	Meta  Meta `json:"meta"`
}

func NewResult[T any](items []T, params Params, total int64) Result[T] {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(params.PerPage) - 1) / int64(params.PerPage))
	}
	return Result[T]{
		Items: items,
		Meta: Meta{
			Page:       params.Page,
			PerPage:    params.PerPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// Parse is opt-in: without page or per_page, handlers keep their legacy array response.
func Parse(values url.Values) (Params, bool, error) {
	if !values.Has("page") && !values.Has("per_page") {
		return Params{}, false, nil
	}
	page, err := positiveInt(values.Get("page"), 1)
	if err != nil {
		return Params{}, true, errors.New("page must be a positive integer")
	}
	perPage, err := positiveInt(values.Get("per_page"), DefaultPerPage)
	if err != nil || perPage > MaxPerPage {
		return Params{}, true, errors.New("per_page must be a positive integer up to 100")
	}
	return Params{Page: page, PerPage: perPage}, true, nil
}

func positiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, errors.New("invalid positive integer")
	}
	return value, nil
}
