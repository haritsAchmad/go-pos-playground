package repository

import (
	"strings"
	"testing"

	"go-pos-playground/internal/pkg/listquery"
)

func TestSupplierQueryParts(t *testing.T) {
	where, order, args, err := supplierQueryParts(listquery.Params{
		Search: "makmur", Sort: "name", Order: "desc", Values: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(where, "s.name ILIKE") || !strings.Contains(where, "s.address ILIKE") {
		t.Fatalf("unexpected where: %q", where)
	}
	if order != " ORDER BY s.name desc, s.id desc" || len(args) != 1 || args[0] != "makmur" {
		t.Fatalf("unexpected result: order=%q args=%#v", order, args)
	}
}

func TestSupplierQueryPartsRejectsUnmappedSort(t *testing.T) {
	_, _, _, err := supplierQueryParts(listquery.Params{
		Sort: "password", Order: "asc", Values: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected invalid sort error")
	}
}
