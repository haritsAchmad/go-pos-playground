package seed

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Options struct {
	Items, Customers, Suppliers, Purchases, Sales, Months int
	DebtRate                                              float64
	RandomSeed                                            int64
}

type Result struct{ Items, Customers, Suppliers, Purchases, Sales, Debts int }
type item struct {
	id, supplierID, cost, price int64
	stock                       int
}

func Generate(ctx context.Context, db *pgxpool.Pool, schema string, o Options) (Result, error) {
	var result Result
	if o.Items < 1 || o.Customers < 1 || o.Suppliers < 1 || o.Months < 1 || o.DebtRate < 0 || o.DebtRate > 1 {
		return result, fmt.Errorf("counts/months must be positive and debt-rate must be between 0 and 1")
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	s := pgx.Identifier{schema}.Sanitize()
	random := rand.New(rand.NewSource(o.RandomSeed))
	loc, _ := time.LoadLocation("Asia/Jakarta")
	var categoryID, brandID, unitID, cashID, debtID int64
	if err = queryID(ctx, tx, fmt.Sprintf(`SELECT id FROM %s.categories ORDER BY id LIMIT 1`, s), &categoryID); err != nil {
		return result, err
	}
	if err = queryID(ctx, tx, fmt.Sprintf(`SELECT id FROM %s.brands ORDER BY id LIMIT 1`, s), &brandID); err != nil {
		return result, err
	}
	if err = queryID(ctx, tx, fmt.Sprintf(`SELECT id FROM %s.units ORDER BY id LIMIT 1`, s), &unitID); err != nil {
		return result, err
	}
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s.payment_methods WHERE name='Tunai'`, s)).Scan(&cashID); err != nil {
		return result, err
	}
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s.payment_methods WHERE name='Piutang'`, s)).Scan(&debtID); err != nil {
		return result, err
	}

	supplierIDs := make([]int64, o.Suppliers)
	for i := range supplierIDs {
		code := fmt.Sprintf("SEED-SUP-%03d", i+1)
		err := tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.suppliers(code,name,phone,address) VALUES($1,$2,$3,$4) ON CONFLICT(code) DO UPDATE SET deleted_at=NULL RETURNING id`, s), code, fmt.Sprintf("Supplier Demo %03d", i+1), fmt.Sprintf("0812%08d", i+1), "Alamat supplier data demo").Scan(&supplierIDs[i])
		if err != nil {
			return result, err
		}
		result.Suppliers++
	}
	customerIDs := make([]int64, o.Customers)
	for i := range customerIDs {
		code := fmt.Sprintf("SEED-CUS-%03d", i+1)
		err := tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.customers(code,name,customer_type,phone,address) VALUES($1,$2,'MEMBER',$3,$4) ON CONFLICT(code) DO UPDATE SET deleted_at=NULL RETURNING id`, s), code, fmt.Sprintf("Pelanggan Demo %03d", i+1), fmt.Sprintf("0821%08d", i+1), "Alamat pelanggan data demo").Scan(&customerIDs[i])
		if err != nil {
			return result, err
		}
		result.Customers++
	}
	items := make([]item, o.Items)
	for i := range items {
		sku := fmt.Sprintf("SEED-BRG-%03d", i+1)
		supplierID := supplierIDs[i%len(supplierIDs)]
		cost := int64(5000 + random.Intn(46)*1000)
		price := cost + int64(2000+random.Intn(20)*1000)
		err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT id,stock FROM %s.items WHERE sku=$1 AND deleted_at IS NULL`, s), sku).Scan(&items[i].id, &items[i].stock)
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.items(sku,name,description,supplier_id,category_id,brand_id,unit_id,stock,cost,price) VALUES($1,$2,$3,$4,$5,$6,$7,0,$8,$9) RETURNING id`, s), sku, fmt.Sprintf("Barang Demo %03d", i+1), "Data hasil seed generator", supplierID, categoryID, brandID, unitID, cost, price).Scan(&items[i].id)
		}
		if err != nil {
			return result, err
		}
		items[i].supplierID, items[i].cost, items[i].price = supplierID, cost, price
		result.Items++
	}

	start := time.Now().In(loc).AddDate(0, -o.Months, 0)
	for i := 0; i < o.Purchases; i++ {
		it := &items[random.Intn(len(items))]
		qty := 5 + random.Intn(16)
		date := randomDate(random, start, time.Now().In(loc))
		total := int64(qty) * it.cost
		var transactionID int64
		invoice := fmt.Sprintf("SEED-PBL-%d-%04d", o.RandomSeed, i+1)
		err := tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.transactions(invoice_no,transaction_type,transaction_date,supplier_id,payment_method_id,payment_status,grand_total,paid_amount,amount_received,notes,created_at) VALUES($1,'PURCHASE',$2,$3,$4,'PAID',$5,$5,$5,'Pembelian data demo',$2) ON CONFLICT(invoice_no) DO NOTHING RETURNING id`, s), invoice, date, it.supplierID, cashID, total).Scan(&transactionID)
		if err == pgx.ErrNoRows {
			continue
		}
		if err != nil {
			return result, err
		}
		if _, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.transaction_items(transaction_id,item_id,quantity,unit_price,subtotal) VALUES($1,$2,$3,$4,$5)`, s), transactionID, it.id, qty, it.cost, total); err != nil {
			return result, err
		}
		if _, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.items SET stock=stock+$1,updated_at=NOW() WHERE id=$2`, s), qty, it.id); err != nil {
			return result, err
		}
		it.stock += qty
		result.Purchases++
	}
	for i := 0; i < o.Sales; i++ {
		candidates := []int{}
		for j := range items {
			if items[j].stock > 0 {
				candidates = append(candidates, j)
			}
		}
		if len(candidates) == 0 {
			break
		}
		it := &items[candidates[random.Intn(len(candidates))]]
		qty := 1 + random.Intn(min(3, it.stock))
		date := randomDate(random, start, time.Now().In(loc))
		total := int64(qty) * it.price
		isDebt := random.Float64() < o.DebtRate
		paid := total
		methodID := cashID
		if isDebt {
			paid = total / 2
			methodID = debtID
		}
		status := "PAID"
		if paid < total {
			status = "PARTIAL"
		}
		var transactionID int64
		invoice := fmt.Sprintf("SEED-PJL-%d-%04d", o.RandomSeed, i+1)
		err := tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.transactions(invoice_no,transaction_type,transaction_date,customer_id,payment_method_id,payment_status,grand_total,paid_amount,amount_received,notes,created_at) VALUES($1,'SALE',$2,$3,$4,$5,$6,$7,$7,'Penjualan data demo',$2) ON CONFLICT(invoice_no) DO NOTHING RETURNING id`, s), invoice, date, customerIDs[random.Intn(len(customerIDs))], methodID, status, total, paid).Scan(&transactionID)
		if err == pgx.ErrNoRows {
			continue
		}
		if err != nil {
			return result, err
		}
		if _, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.transaction_items(transaction_id,item_id,quantity,unit_price,subtotal) VALUES($1,$2,$3,$4,$5)`, s), transactionID, it.id, qty, it.price, total); err != nil {
			return result, err
		}
		if _, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.items SET stock=stock-$1,updated_at=NOW() WHERE id=$2`, s), qty, it.id); err != nil {
			return result, err
		}
		it.stock -= qty
		if isDebt {
			if _, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.debts(transaction_id,customer_id,original_amount,remaining_amount,status,created_at,updated_at) SELECT id,customer_id,$2,$2,'OPEN',$3,$3 FROM %s.transactions WHERE id=$1`, s, s), transactionID, total-paid, date); err != nil {
				return result, err
			}
			result.Debts++
		}
		result.Sales++
	}
	if err := tx.Commit(ctx); err != nil {
		return result, err
	}
	return result, nil
}

func queryID(ctx context.Context, tx pgx.Tx, q string, id *int64) error {
	return tx.QueryRow(ctx, q).Scan(id)
}
func randomDate(r *rand.Rand, start, end time.Time) time.Time {
	span := end.Sub(start)
	return start.Add(time.Duration(r.Int63n(int64(span))))
}
