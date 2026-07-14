package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-inventory-playground/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CooperativeRepository struct {
	db     *pgxpool.Pool
	schema string
}

func NewCooperativeRepository(db *pgxpool.Pool, schema string) *CooperativeRepository {
	return &CooperativeRepository{db: db, schema: pgx.Identifier{schema}.Sanitize()}
}

func (r *CooperativeRepository) Masters(ctx context.Context, table string) ([]entity.MasterData, error) {
	allowed := map[string]bool{"categories": true, "brands": true, "units": true, "payment_methods": true}
	if !allowed[table] {
		return nil, errors.New("invalid master table")
	}
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT id, name FROM %s.%s ORDER BY name`, r.schema, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.MasterData, 0)
	for rows.Next() {
		var value entity.MasterData
		if err := rows.Scan(&value.ID, &value.Name); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *CooperativeRepository) CreateMaster(ctx context.Context, table, name string) error {
	allowed := map[string]bool{"categories": true, "brands": true, "units": true, "payment_methods": true}
	if !allowed[table] {
		return errors.New("invalid master table")
	}
	_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.%s (name) VALUES ($1)`, r.schema, table), name)
	return err
}

func (r *CooperativeRepository) UpdateMaster(ctx context.Context, table string, id int64, name string) error {
	allowed := map[string]bool{"categories": true, "brands": true, "units": true, "payment_methods": true}
	if !allowed[table] {
		return errors.New("invalid master table")
	}
	tag, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE %s.%s SET name=$1 WHERE id=$2`, r.schema, table), name, id)
	if err == nil && tag.RowsAffected() == 0 {
		return errors.New("master data not found")
	}
	return err
}

func (r *CooperativeRepository) DeleteMaster(ctx context.Context, table string, id int64) error {
	allowed := map[string]bool{"categories": true, "brands": true, "units": true, "payment_methods": true}
	if !allowed[table] {
		return errors.New("invalid master table")
	}
	tag, err := r.db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.%s WHERE id=$1`, r.schema, table), id)
	if err == nil && tag.RowsAffected() == 0 {
		return errors.New("master data not found")
	}
	return err
}

func (r *CooperativeRepository) Customers(ctx context.Context) ([]entity.Customer, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT id, code, name, customer_type, phone, address, created_at FROM %s.customers WHERE deleted_at IS NULL ORDER BY CASE WHEN code='UMUM' THEN 0 ELSE 1 END, name`, r.schema))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Customer, 0)
	for rows.Next() {
		var value entity.Customer
		if err := rows.Scan(&value.ID, &value.Code, &value.Name, &value.CustomerType, &value.Phone, &value.Address, &value.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *CooperativeRepository) CreateCustomer(ctx context.Context, value entity.Customer) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customers (code,name,customer_type,phone,address) VALUES ($1,$2,$3,$4,$5)`, r.schema), value.Code, value.Name, value.CustomerType, value.Phone, value.Address)
	return err
}

func (r *CooperativeRepository) UpdateCustomer(ctx context.Context, id int64, value entity.Customer) error {
	tag, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE %s.customers SET code=$1,name=$2,phone=$3,address=$4,updated_at=NOW() WHERE id=$5 AND deleted_at IS NULL AND code<>'UMUM'`, r.schema), value.Code, value.Name, value.Phone, value.Address, id)
	if err == nil && tag.RowsAffected() == 0 {
		return errors.New("customer not found or protected")
	}
	return err
}

func (r *CooperativeRepository) DeleteCustomer(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE %s.customers SET deleted_at=NOW(),updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL AND code<>'UMUM'`, r.schema), id)
	if err == nil && tag.RowsAffected() == 0 {
		return errors.New("customer not found or protected")
	}
	return err
}

func (r *CooperativeRepository) CreateTransaction(ctx context.Context, req entity.CreateTransactionRequest) (entity.Transaction, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return entity.Transaction{}, err
	}
	defer tx.Rollback(ctx)
	if req.TransactionType == "SALE" && req.CustomerID == nil {
		return entity.Transaction{}, errors.New("customer is required for a sale")
	}
	if req.TransactionType == "PURCHASE" && req.SupplierID == nil {
		return entity.Transaction{}, errors.New("supplier is required for a purchase")
	}
	var total int64
	seenItems := make(map[int64]struct{}, len(req.Items))
	for i := range req.Items {
		if _, exists := seenItems[req.Items[i].ItemID]; exists {
			return entity.Transaction{}, fmt.Errorf("item %d appears more than once", req.Items[i].ItemID)
		}
		seenItems[req.Items[i].ItemID] = struct{}{}
		if req.Items[i].UnitPrice == 0 {
			column := "price"
			if req.TransactionType == "PURCHASE" {
				column = "cost"
			}
			if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM %s.items WHERE id=$1 AND deleted_at IS NULL`, column, r.schema), req.Items[i].ItemID).Scan(&req.Items[i].UnitPrice); err != nil {
				return entity.Transaction{}, err
			}
		}
		req.Items[i].Subtotal = int64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		total += req.Items[i].Subtotal
	}
	if total <= 0 {
		return entity.Transaction{}, errors.New("transaction total must be greater than zero")
	}
	received := req.PaidAmount
	paid := received
	if paid > total {
		paid = total
	}
	changeAmount := received - paid
	var paymentMethod string
	if req.PaymentMethodID == nil {
		return entity.Transaction{}, errors.New("metode pembayaran wajib dipilih")
	}
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT name FROM %s.payment_methods WHERE id=$1`, r.schema), req.PaymentMethodID).Scan(&paymentMethod); err != nil {
		return entity.Transaction{}, errors.New("metode pembayaran tidak valid")
	}
	if received < total && !strings.EqualFold(paymentMethod, "Piutang") {
		return entity.Transaction{}, fmt.Errorf("pembayaran kurang Rp%d; pilih metode Piutang atau masukkan pembayaran penuh", total-received)
	}
	if req.TransactionType == "PURCHASE" && received < total {
		return entity.Transaction{}, fmt.Errorf("pembayaran pembelian kurang Rp%d", total-received)
	}
	status := "PAID"
	if paid == 0 {
		status = "UNPAID"
	} else if paid < total {
		status = "PARTIAL"
	}
	invoice := fmt.Sprintf("%s-%s", map[string]string{"SALE": "JL", "PURCHASE": "BL"}[req.TransactionType], time.Now().Format("20060102-150405.000000"))
	var id int64
	err = tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.transactions (invoice_no,transaction_type,customer_id,supplier_id,payment_method_id,payment_status,grand_total,paid_amount,amount_received,change_amount,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, r.schema), invoice, req.TransactionType, req.CustomerID, req.SupplierID, req.PaymentMethodID, status, total, paid, received, changeAmount, req.Notes).Scan(&id)
	if err != nil {
		return entity.Transaction{}, err
	}
	for _, line := range req.Items {
		var stock int
		if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT stock FROM %s.items WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, r.schema), line.ItemID).Scan(&stock); err != nil {
			return entity.Transaction{}, err
		}
		change := line.Quantity
		if req.TransactionType == "SALE" {
			change = -change
			if stock+change < 0 {
				return entity.Transaction{}, fmt.Errorf("insufficient stock for item %d", line.ItemID)
			}
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.transaction_items (transaction_id,item_id,quantity,unit_price,subtotal) VALUES ($1,$2,$3,$4,$5)`, r.schema), id, line.ItemID, line.Quantity, line.UnitPrice, line.Subtotal)
		if err != nil {
			return entity.Transaction{}, err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.items SET stock=stock+$1, updated_at=NOW() WHERE id=$2`, r.schema), change, line.ItemID)
		if err != nil {
			return entity.Transaction{}, err
		}
	}
	if req.TransactionType == "SALE" && paid < total {
		_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.debts (transaction_id,customer_id,original_amount,remaining_amount) VALUES ($1,$2,$3,$4)`, r.schema), id, req.CustomerID, total-paid, total-paid)
		if err != nil {
			return entity.Transaction{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return entity.Transaction{}, err
	}
	return entity.Transaction{ID: id, InvoiceNo: invoice, TransactionType: req.TransactionType, GrandTotal: total, PaidAmount: paid, AmountReceived: received, ChangeAmount: changeAmount, PaymentStatus: status, Status: "ACTIVE"}, nil
}

func (r *CooperativeRepository) Transactions(ctx context.Context, kind string) ([]entity.Transaction, error) {
	args := []any{}
	filter := ""
	if kind == "SALE" || kind == "PURCHASE" {
		filter = "WHERE t.transaction_type=$1"
		args = append(args, kind)
	}
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT t.id,t.invoice_no,t.transaction_type,t.transaction_date,t.customer_id,c.name,t.supplier_id,s.name,t.payment_method_id,p.name,t.payment_status,t.grand_total,t.paid_amount,t.amount_received,t.change_amount,t.status,t.void_reason,t.notes FROM %s.transactions t LEFT JOIN %s.customers c ON c.id=t.customer_id LEFT JOIN %s.suppliers s ON s.id=t.supplier_id LEFT JOIN %s.payment_methods p ON p.id=t.payment_method_id %s ORDER BY t.transaction_date DESC`, r.schema, r.schema, r.schema, r.schema, filter), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Transaction, 0)
	for rows.Next() {
		var v entity.Transaction
		if err := rows.Scan(&v.ID, &v.InvoiceNo, &v.TransactionType, &v.TransactionDate, &v.CustomerID, &v.CustomerName, &v.SupplierID, &v.SupplierName, &v.PaymentMethodID, &v.PaymentMethodName, &v.PaymentStatus, &v.GrandTotal, &v.PaidAmount, &v.AmountReceived, &v.ChangeAmount, &v.Status, &v.VoidReason, &v.Notes); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *CooperativeRepository) VoidTransaction(ctx context.Context, id int64, reason string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var kind, status string
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT transaction_type,status FROM %s.transactions WHERE id=$1 FOR UPDATE`, r.schema), id).Scan(&kind, &status); err != nil {
		return err
	}
	if status == "VOID" {
		return errors.New("transaksi sudah dibatalkan")
	}
	var payments int
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.debt_payments dp JOIN %s.debts d ON d.id=dp.debt_id WHERE d.transaction_id=$1`, r.schema, r.schema), id).Scan(&payments); err != nil {
		return err
	}
	if payments > 0 {
		return errors.New("transaksi tidak dapat dibatalkan karena piutang sudah pernah dibayar")
	}
	rows, err := tx.Query(ctx, fmt.Sprintf(`SELECT item_id,quantity FROM %s.transaction_items WHERE transaction_id=$1`, r.schema), id)
	if err != nil {
		return err
	}
	type line struct {
		id  int64
		qty int
	}
	var lines []line
	for rows.Next() {
		var v line
		if err := rows.Scan(&v.id, &v.qty); err != nil {
			rows.Close()
			return err
		}
		lines = append(lines, v)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, v := range lines {
		change := -v.qty
		if kind == "SALE" {
			change = v.qty
		} else {
			var stock int
			if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT stock FROM %s.items WHERE id=$1 FOR UPDATE`, r.schema), v.id).Scan(&stock); err != nil {
				return err
			}
			if stock-v.qty < 0 {
				return errors.New("pembelian tidak dapat dibatalkan karena stoknya sudah terpakai")
			}
		}
		if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.items SET stock=stock+$1,updated_at=NOW() WHERE id=$2`, r.schema), change, v.id); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.debts WHERE transaction_id=$1`, r.schema), id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.transactions SET status='VOID',void_reason=$1,voided_at=NOW() WHERE id=$2`, r.schema), reason, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *CooperativeRepository) Debts(ctx context.Context) ([]entity.Debt, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT d.id,d.transaction_id,t.invoice_no,d.customer_id,c.name,d.original_amount,d.remaining_amount,d.status,d.created_at FROM %s.debts d JOIN %s.transactions t ON t.id=d.transaction_id JOIN %s.customers c ON c.id=d.customer_id ORDER BY d.created_at DESC`, r.schema, r.schema, r.schema))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.Debt, 0)
	for rows.Next() {
		var v entity.Debt
		if err := rows.Scan(&v.ID, &v.TransactionID, &v.InvoiceNo, &v.CustomerID, &v.CustomerName, &v.OriginalAmount, &v.RemainingAmount, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *CooperativeRepository) PayDebt(ctx context.Context, id, amount int64, notes string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var remaining, transactionID int64
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT remaining_amount,transaction_id FROM %s.debts WHERE id=$1 FOR UPDATE`, r.schema), id).Scan(&remaining, &transactionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("piutang tidak ditemukan")
		}
		return err
	}
	if amount <= 0 || amount > remaining {
		return errors.New("jumlah pembayaran harus lebih dari 0 dan tidak boleh melebihi sisa piutang")
	}
	if _, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.debt_payments(debt_id,amount,notes) VALUES($1,$2,$3)`, r.schema), id, amount, strings.TrimSpace(notes)); err != nil {
		return err
	}
	newRemaining := remaining - amount
	debtStatus := "OPEN"
	paymentStatus := "PARTIAL"
	if newRemaining == 0 {
		debtStatus = "PAID"
		paymentStatus = "PAID"
	}
	if _, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.debts SET remaining_amount=$1,status=$2,updated_at=NOW() WHERE id=$3`, r.schema), newRemaining, debtStatus, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.transactions SET paid_amount=paid_amount+$1,payment_status=$2 WHERE id=$3 AND status='ACTIVE'`, r.schema), amount, paymentStatus, transactionID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *CooperativeRepository) Dashboard(ctx context.Context, year int) (entity.Dashboard, error) {
	var v entity.Dashboard
	v.Year = year
	v.MonthlySales = make([]int64, 12)
	err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COALESCE((SELECT SUM(grand_total) FROM %s.transactions WHERE transaction_type='SALE' AND status='ACTIVE' AND transaction_date::date=CURRENT_DATE),0),COALESCE((SELECT SUM(grand_total) FROM %s.transactions WHERE transaction_type='PURCHASE' AND status='ACTIVE' AND transaction_date::date=CURRENT_DATE),0),COALESCE((SELECT SUM(remaining_amount) FROM %s.debts WHERE status='OPEN'),0),(SELECT COUNT(*) FROM %s.items WHERE deleted_at IS NULL AND stock<=5),(SELECT COUNT(*) FROM %s.items WHERE deleted_at IS NULL),(SELECT COUNT(*) FROM %s.customers WHERE deleted_at IS NULL),(SELECT COUNT(*) FROM %s.suppliers WHERE deleted_at IS NULL)`, r.schema, r.schema, r.schema, r.schema, r.schema, r.schema, r.schema)).Scan(&v.TodaySales, &v.TodayPurchases, &v.OpenDebt, &v.LowStockItems, &v.TotalItems, &v.TotalCustomers, &v.TotalSuppliers)
	if err != nil {
		return v, err
	}
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT EXTRACT(MONTH FROM transaction_date)::int,SUM(grand_total) FROM %s.transactions WHERE transaction_type='SALE' AND status='ACTIVE' AND EXTRACT(YEAR FROM transaction_date)=$1 GROUP BY 1`, r.schema), year)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	for rows.Next() {
		var month int
		var total int64
		if err := rows.Scan(&month, &total); err != nil {
			return v, err
		}
		v.MonthlySales[month-1] = total
	}
	return v, rows.Err()
}
