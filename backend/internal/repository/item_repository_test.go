package repository

import (
	"strings"
	"testing"

	"go-pos-playground/internal/pkg/listquery"
)

func TestItemQueryParts(t *testing.T) {
	query := listquery.Params{
		Search: "kopi",
		Sort:   "stock",
		Order:  "desc",
		Values: map[string]string{
			"supplier_id": "7",
			"min_stock":   "2",
			"max_stock":   "20",
		},
	}
	where, order, args, err := itemQueryParts(query)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"i.deleted_at IS NULL", "i.name ILIKE", "i.supplier_id=$2", "i.stock>=$3", "i.stock<=$4"} {
		if !strings.Contains(where, fragment) {
			t.Fatalf("where %q does not contain %q", where, fragment)
		}
	}
	if order != " ORDER BY i.stock desc, i.id desc" {
		t.Fatalf("unexpected order: %q", order)
	}
	if len(args) != 4 || args[0] != "kopi" || args[1] != int64(7) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestItemQueryPartsRejectsInvalidStockRange(t *testing.T) {
	_, _, _, err := itemQueryParts(listquery.Params{
		Sort:   "id",
		Order:  "asc",
		Values: map[string]string{"min_stock": "10", "max_stock": "2"},
	})
	if err == nil {
		t.Fatal("expected invalid stock range error")
	}
}

func TestItemQueryPartsRejectsUnmappedSort(t *testing.T) {
	_, _, _, err := itemQueryParts(listquery.Params{
		Sort:   "deleted_at",
		Order:  "asc",
		Values: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected invalid sort error")
	}
}
