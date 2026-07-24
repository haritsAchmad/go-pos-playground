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

func TestTransactionQueryParts(t *testing.T) {
	where, order, args, err := transactionQueryParts("SALE", listquery.Params{
		Search: "PJL-001",
		Sort:   "grand_total",
		Order:  "desc",
		Values: map[string]string{
			"payment_status": "PARTIAL",
			"status":         "ACTIVE",
			"date_from":      "2026-07-01",
			"date_to":        "2026-07-31",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"t.transaction_type=$1", "t.invoice_no ILIKE", "t.payment_status=$3", "t.status=$4", "::date>=$5::date", "::date<=$6::date"} {
		if !strings.Contains(where, fragment) {
			t.Fatalf("where %q does not contain %q", where, fragment)
		}
	}
	if order != " ORDER BY t.grand_total desc, t.id desc" || len(args) != 6 {
		t.Fatalf("unexpected result: order=%q args=%#v", order, args)
	}
}

func TestTransactionQueryPartsRejectsInvalidDateRange(t *testing.T) {
	_, _, _, err := transactionQueryParts("", listquery.Params{
		Sort: "transaction_date", Order: "desc",
		Values: map[string]string{"date_from": "2026-08-01", "date_to": "2026-07-01"},
	})
	if err == nil {
		t.Fatal("expected invalid date range error")
	}
}

func TestDebtQueryParts(t *testing.T) {
	where, order, args, err := debtQueryParts(listquery.Params{
		Search: "budi",
		Sort:   "remaining_amount",
		Order:  "asc",
		Values: map[string]string{"status": "OPEN", "min_remaining": "1000", "max_remaining": "50000"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"c.name ILIKE", "d.status=$2", "d.remaining_amount>=$3", "d.remaining_amount<=$4"} {
		if !strings.Contains(where, fragment) {
			t.Fatalf("where %q does not contain %q", where, fragment)
		}
	}
	if order != " ORDER BY d.remaining_amount asc, d.id asc" || len(args) != 4 {
		t.Fatalf("unexpected result: order=%q args=%#v", order, args)
	}
}

func TestDebtQueryPartsRejectsInvalidRange(t *testing.T) {
	_, _, _, err := debtQueryParts(listquery.Params{
		Sort: "created_at", Order: "desc",
		Values: map[string]string{"min_remaining": "50", "max_remaining": "10"},
	})
	if err == nil {
		t.Fatal("expected invalid remaining amount range")
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
