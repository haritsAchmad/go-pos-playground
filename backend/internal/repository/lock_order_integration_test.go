package repository

import (
	"fmt"
	"testing"
	"time"

	"go-pos-playground/internal/entity"

	"github.com/jackc/pgx/v5"
)

// Transactions that touch the same items in a different cart order must not
// deadlock. Without a common lock order, A (first, second) and B (second,
// first) can each hold one item and wait for the other; PostgreSQL then
// aborts one of them with "deadlock detected".
//
// The interleaving is forced instead of hoped for: a third connection holds
// the first item, A queues behind it, B (old code) locks the second item and
// queues too, and releasing the holder lets A take the first item and ask
// for the second one that B holds.
func TestStockChangesLockItemsInIDOrderIntegration(t *testing.T) {
	type operation func(f *transactionFixture, first, second int64) (func() error, func() error)

	sale := func(f *transactionFixture, items ...int64) entity.CreateTransactionRequest {
		request := f.request("SALE", 1, int64(len(items))*1000, false)
		request.Items = nil
		for _, item := range items {
			request.Items = append(request.Items, entity.TransactionLine{ItemID: item, Quantity: 1})
		}
		return request
	}
	create := func(t *testing.T, f *transactionFixture, items ...int64) int64 {
		t.Helper()
		transaction, err := f.repository.CreateTransaction(f.ctx, sale(f, items...))
		if err != nil {
			t.Fatalf("create sale %v: %v", items, err)
		}
		return transaction.ID
	}

	for name, op := range map[string]operation{
		"checkout": func(f *transactionFixture, first, second int64) (func() error, func() error) {
			return func() error { _, err := f.repository.CreateTransaction(f.ctx, sale(f, first, second)); return err },
				func() error { _, err := f.repository.CreateTransaction(f.ctx, sale(f, second, first)); return err }
		},
		"void": func(f *transactionFixture, first, second int64) (func() error, func() error) {
			a, b := create(t, f, first, second), create(t, f, second, first)
			return func() error { return f.repository.VoidTransaction(f.ctx, a, "uji urutan kunci") },
				func() error { return f.repository.VoidTransaction(f.ctx, b, "uji urutan kunci") }
		},
		"update": func(f *transactionFixture, first, second int64) (func() error, func() error) {
			a, b := create(t, f, first, second), create(t, f, second, first)
			return func() error { _, err := f.repository.UpdateTransaction(f.ctx, a, sale(f, first, second)); return err },
				func() error { _, err := f.repository.UpdateTransaction(f.ctx, b, sale(f, second, first)); return err }
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := newTransactionFixture(t)
			q := pgx.Identifier{f.schema}.Sanitize()
			first := f.itemID
			var second int64
			if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`INSERT INTO %s.items(supplier_id,sku,name,unit_id,base_unit_id,units_per_package,stock,price,cost,retail_price,retail_cost) VALUES($1,'ITEM-TEST-2','Barang Test 2',$2,$2,1,10,1000,500,1000,500) RETURNING id`, q), f.supplierID, f.pcsUnitID).Scan(&second); err != nil {
				t.Fatal(err)
			}
			if second <= first {
				t.Fatalf("expected the second item to have the larger ID: %d, %d", first, second)
			}
			runA, runB := op(f, first, second)

			holder, err := f.db.Begin(f.ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer holder.Rollback(f.ctx)
			if _, err := holder.Exec(f.ctx, fmt.Sprintf(`SELECT id FROM %s.items WHERE id=$1 FOR UPDATE`, q), first); err != nil {
				t.Fatal(err)
			}

			results := make(chan error, 2)
			go func() { results <- runA() }()
			waitForLockWaiters(t, f, 1)
			go func() { results <- runB() }()
			waitForLockWaiters(t, f, 2)
			if err := holder.Rollback(f.ctx); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				select {
				case err := <-results:
					if err != nil {
						t.Errorf("concurrent %s failed: %v", name, err)
					}
				case <-time.After(30 * time.Second):
					t.Fatalf("concurrent %s did not finish", name)
				}
			}
		})
	}
}

// waitForLockWaiters waits until n sessions working on this fixture's schema
// are blocked on a row lock.
func waitForLockWaiters(t *testing.T, f *transactionFixture, n int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		if err := f.db.QueryRow(f.ctx, `SELECT count(*) FROM pg_stat_activity WHERE wait_event_type='Lock' AND datname=current_database() AND query LIKE '%' || $1 || '%'`, f.schema).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting >= n {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected %d sessions waiting for a lock", n)
}
