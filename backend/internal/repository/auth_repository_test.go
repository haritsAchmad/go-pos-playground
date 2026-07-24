package repository

import (
	"strings"
	"testing"

	"go-pos-playground/internal/pkg/listquery"
)

func TestUserQueryParts(t *testing.T) {
	where, order, args, err := userQueryParts(listquery.Params{
		Search: "andi",
		Sort:   "email",
		Order:  "desc",
		Values: map[string]string{"role": "cashier", "active": "true"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"u.name ILIKE", "u.role=$2", "u.active=$3"} {
		if !strings.Contains(where, fragment) {
			t.Fatalf("where %q does not contain %q", where, fragment)
		}
	}
	if order != " ORDER BY u.email desc, u.id desc" || len(args) != 3 || args[2] != true {
		t.Fatalf("unexpected result: order=%q args=%#v", order, args)
	}
}

func TestUserQueryPartsRejectsInvalidFilters(t *testing.T) {
	for _, values := range []map[string]string{
		{"role": "owner"},
		{"active": "yes"},
	} {
		_, _, _, err := userQueryParts(listquery.Params{Sort: "id", Order: "asc", Values: values})
		if err == nil {
			t.Fatalf("expected error for filters %#v", values)
		}
	}
}
