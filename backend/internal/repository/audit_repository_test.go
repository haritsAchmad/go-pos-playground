package repository

import (
	"strings"
	"testing"

	"go-pos-playground/internal/pkg/listquery"
)

func TestAuditQueryParts(t *testing.T) {
	where, order, args, err := auditQueryParts(listquery.Params{
		Search: "admin",
		Sort:   "created_at",
		Order:  "desc",
		Values: map[string]string{
			"user_id":     "7",
			"action":      "UPDATE",
			"entity_type": "items",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"a.user_name ILIKE", "a.action=$2", "a.entity_type=$3", "a.user_id=$4"} {
		if !strings.Contains(where, fragment) {
			t.Fatalf("where %q does not contain %q", where, fragment)
		}
	}
	if order != " ORDER BY a.created_at desc, a.id desc" || len(args) != 4 {
		t.Fatalf("unexpected result: order=%q args=%#v", order, args)
	}
}

func TestAuditQueryPartsRejectsUnsafeSort(t *testing.T) {
	_, _, _, err := auditQueryParts(listquery.Params{
		Sort: "user_agent", Order: "asc", Values: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected invalid sort error")
	}
}
