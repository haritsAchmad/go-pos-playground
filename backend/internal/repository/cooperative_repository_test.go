package repository

import (
	"strings"
	"testing"

	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/listquery"
)

func TestCustomerQueryParts(t *testing.T) {
	where, order, args, err := customerQueryParts(listquery.Params{
		Search: "budi",
		Sort:   "name",
		Order:  "asc",
		Values: map[string]string{"customer_type": "MEMBER"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(where, "c.name ILIKE") || !strings.Contains(where, "c.customer_type=$2") {
		t.Fatalf("unexpected where: %q", where)
	}
	if !strings.Contains(order, "CASE WHEN c.code='UMUM'") || len(args) != 2 || args[1] != "MEMBER" {
		t.Fatalf("unexpected result: order=%q args=%#v", order, args)
	}
}

func TestCustomerQueryPartsRejectsInvalidFilter(t *testing.T) {
	_, _, _, err := customerQueryParts(listquery.Params{
		Sort: "name", Order: "asc", Values: map[string]string{"customer_type": "VIP"},
	})
	if err == nil {
		t.Fatal("expected invalid customer type error")
	}
}

func TestMergeTransactionItems(t *testing.T) {
	items, err := mergeTransactionItems([]entity.TransactionLine{
		{ItemID: 10, Quantity: 1, UnitPrice: 5000},
		{ItemID: 20, Quantity: 2, UnitPrice: 3000},
		{ItemID: 10, Quantity: 3, UnitPrice: 5000},
	})
	if err != nil {
		t.Fatalf("mergeTransactionItems returned an error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ItemID != 10 || items[0].Quantity != 4 {
		t.Fatalf("first item = %+v, want item 10 with quantity 4", items[0])
	}
}

func TestMergeTransactionItemsRejectsDifferentPrices(t *testing.T) {
	_, err := mergeTransactionItems([]entity.TransactionLine{
		{ItemID: 10, Quantity: 1, UnitPrice: 5000},
		{ItemID: 10, Quantity: 1, UnitPrice: 6000},
	})
	if err == nil {
		t.Fatal("expected an error for repeated item with different prices")
	}
}
