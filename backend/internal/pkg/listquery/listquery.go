package listquery

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const MaxSearchLength = 100

type Params struct {
	Search string
	Sort   string
	Order  string
	Values map[string]string
}

type Config struct {
	DefaultSort string
	Sorts       map[string]bool
	Filters     map[string]bool
}

func Parse(values url.Values, config Config) (Params, error) {
	params := Params{
		Search: strings.TrimSpace(values.Get("search")),
		Sort:   strings.TrimSpace(values.Get("sort")),
		Order:  strings.ToLower(strings.TrimSpace(values.Get("order"))),
		Values: make(map[string]string),
	}
	if len(params.Search) > MaxSearchLength {
		return Params{}, errors.New("search must not exceed 100 characters")
	}
	if params.Sort == "" {
		params.Sort = config.DefaultSort
	}
	if !config.Sorts[params.Sort] {
		return Params{}, errors.New("invalid sort field")
	}
	if params.Order == "" {
		params.Order = "asc"
	}
	if params.Order != "asc" && params.Order != "desc" {
		return Params{}, errors.New("order must be asc or desc")
	}
	for key := range config.Filters {
		value := strings.TrimSpace(values.Get(key))
		if value != "" {
			params.Values[key] = value
		}
	}
	return params, nil
}

func (p Params) PositiveInt(key string) (int64, bool, error) {
	raw, ok := p.Values[key]
	if !ok {
		return 0, false, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 1 {
		return 0, true, errors.New(key + " must be a positive integer")
	}
	return value, true, nil
}

func (p Params) NonNegativeInt(key string) (int64, bool, error) {
	raw, ok := p.Values[key]
	if !ok {
		return 0, false, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, true, errors.New(key + " must be a non-negative integer")
	}
	return value, true, nil
}
