package pagination

import (
	"net/url"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		enabled bool
		want    Params
		wantErr bool
	}{
		{name: "legacy request", query: "", enabled: false},
		{name: "defaults per page", query: "page=2", enabled: true, want: Params{Page: 2, PerPage: 20}},
		{name: "defaults page", query: "per_page=50", enabled: true, want: Params{Page: 1, PerPage: 50}},
		{name: "custom values", query: "page=3&per_page=25", enabled: true, want: Params{Page: 3, PerPage: 25}},
		{name: "invalid page", query: "page=0", enabled: true, wantErr: true},
		{name: "invalid per page", query: "per_page=101", enabled: true, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values, _ := url.ParseQuery(test.query)
			got, enabled, err := Parse(values)
			if (err != nil) != test.wantErr || enabled != test.enabled || (!test.wantErr && got != test.want) {
				t.Fatalf("Parse() = %+v, %v, %v; want %+v, %v, error=%v", got, enabled, err, test.want, test.enabled, test.wantErr)
			}
		})
	}
}

func TestNewResult(t *testing.T) {
	result := NewResult([]int{1, 2}, Params{Page: 3, PerPage: 2}, 5)
	if result.Meta.TotalPages != 3 || result.Meta.Total != 5 {
		t.Fatalf("meta = %+v, want total=5 and total_pages=3", result.Meta)
	}
}
