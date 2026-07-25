package listquery

import (
	"net/url"
	"testing"
)

func TestParse(t *testing.T) {
	config := Config{
		DefaultSort: "id",
		Sorts:       map[string]bool{"id": true, "name": true},
		Filters:     map[string]bool{"supplier_id": true},
	}
	values, _ := url.ParseQuery("search=  kopi  &sort=name&order=DESC&supplier_id=3&ignored=x")
	got, err := Parse(values, config)
	if err != nil {
		t.Fatal(err)
	}
	if got.Search != "kopi" || got.Sort != "name" || got.Order != "desc" || got.Values["supplier_id"] != "3" {
		t.Fatalf("unexpected params: %+v", got)
	}
	if _, exists := got.Values["ignored"]; exists {
		t.Fatal("unknown filter must be ignored")
	}
}

func TestParseRejectsInvalidSortAndOrder(t *testing.T) {
	config := Config{DefaultSort: "id", Sorts: map[string]bool{"id": true}}
	for _, query := range []string{"sort=password", "order=sideways"} {
		values, _ := url.ParseQuery(query)
		if _, err := Parse(values, config); err == nil {
			t.Fatalf("Parse(%q) expected an error", query)
		}
	}
}

func TestIntegerFilters(t *testing.T) {
	params := Params{Values: map[string]string{"supplier_id": "2", "min_stock": "0", "max_stock": "-1"}}
	if value, set, err := params.PositiveInt("supplier_id"); err != nil || !set || value != 2 {
		t.Fatalf("PositiveInt() = %d, %v, %v", value, set, err)
	}
	if value, set, err := params.NonNegativeInt("min_stock"); err != nil || !set || value != 0 {
		t.Fatalf("NonNegativeInt() = %d, %v, %v", value, set, err)
	}
	if _, _, err := params.NonNegativeInt("max_stock"); err == nil {
		t.Fatal("negative value must fail")
	}
}
