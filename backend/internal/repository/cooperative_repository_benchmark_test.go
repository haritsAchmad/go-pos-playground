package repository

import (
	"fmt"
	"testing"

	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/pagination"

	"github.com/jackc/pgx/v5"
)

// BenchmarkCreateSaleIntegration measures the complete repository write path:
// transaction creation, row locking, stock mutation, and stock movement audit.
func BenchmarkCreateSaleIntegration(b *testing.B) {
	f := newTransactionFixture(b)
	q := pgx.Identifier{f.schema}.Sanitize()
	if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`UPDATE %s.items SET stock=$1 WHERE id=$2`, q), b.N+100, f.itemID); err != nil {
		b.Fatalf("prepare benchmark stock: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 1, 1000, false)); err != nil {
			b.Fatalf("create sale: %v", err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "sales/s")
}

// BenchmarkTransactionsPageIntegration measures a representative cashier
// history read against a non-empty transaction table.
func BenchmarkTransactionsPageIntegration(b *testing.B) {
	f := newTransactionFixture(b)
	q := pgx.Identifier{f.schema}.Sanitize()
	const rows = 1000
	if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`
		INSERT INTO %s.transactions(
			invoice_no,transaction_type,customer_id,payment_method_id,
			payment_status,grand_total,paid_amount,amount_received
		)
		SELECT 'BENCH-' || n,'SALE',$1,$2,'PAID',1000,1000,1000
		FROM generate_series(1,$3) AS n
	`, q), f.customerID, f.cashID, rows); err != nil {
		b.Fatalf("seed benchmark transactions: %v", err)
	}
	if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`
		INSERT INTO %s.transaction_items(
			transaction_id,item_id,unit_id,unit_name,quantity,
			conversion_factor,base_quantity,unit_price,subtotal
		)
		SELECT id,$1,$2,'Pcs',1,1,1,1000,1000
		FROM %s.transactions
	`, q, q), f.itemID, f.pcsUnitID); err != nil {
		b.Fatalf("seed benchmark transaction items: %v", err)
	}

	params := pagination.Params{Page: 1, PerPage: 25}
	query := listquery.Params{Sort: "transaction_date", Order: "desc", Values: map[string]string{}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := f.repository.TransactionsPageQuery(f.ctx, "SALE", params, query)
		if err != nil {
			b.Fatalf("list transaction page: %v", err)
		}
		if len(result.Items) != 25 || result.Meta.Total != rows {
			b.Fatalf("unexpected page size/total: %d/%d", len(result.Items), result.Meta.Total)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "pages/s")
}
