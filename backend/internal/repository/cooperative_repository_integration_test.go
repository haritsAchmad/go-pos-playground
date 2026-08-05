package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"go-pos-playground/internal/config"
	"go-pos-playground/internal/database"
	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/pagination"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func TestAuditLogIdentitySnapshotIntegration(t *testing.T) {
	f := newTransactionFixture(t)
	q := pgx.Identifier{f.schema}.Sanitize()
	var userID int64
	if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`
		INSERT INTO %s.users(name,email,password_hash,role)
		VALUES('Audit User','audit@example.com','unused','cashier')
		RETURNING id
	`, q)).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	legacyStatements := []struct {
		query string
		args  []any
	}{
		{fmt.Sprintf(`ALTER TABLE %s.audit_logs DROP COLUMN user_name`, q), nil},
		{fmt.Sprintf(`ALTER TABLE %s.audit_logs DROP COLUMN user_email`, q), nil},
		{fmt.Sprintf(`ALTER TABLE %s.audit_logs ADD CONSTRAINT audit_logs_user_id_fkey FOREIGN KEY(user_id) REFERENCES %s.users(id) ON DELETE RESTRICT`, q, q), nil},
		{fmt.Sprintf(`INSERT INTO %s.audit_logs(user_id,action,entity_type,entity_id,method,path,status_code) VALUES($1,'UPDATE','items','42','PUT','/items/42',200)`, q), []any{userID}},
	}
	for _, statement := range legacyStatements {
		if _, err := f.db.Exec(f.ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	_ = godotenv.Load("../../.env")
	if err := database.Migrate(f.ctx, f.db, f.schema); err != nil {
		t.Fatalf("upgrade legacy audit schema: %v", err)
	}
	repo := NewAuditRepository(f.db, f.schema)
	if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`DELETE FROM %s.users WHERE id=$1`, q), userID); err != nil {
		t.Fatalf("audit history must not prevent user deletion after migration: %v", err)
	}
	result, err := repo.ListPage(f.ctx, pagination.Params{Page: 1, PerPage: 10}, listquery.Params{
		Search: "audit@example.com", Sort: "created_at", Order: "desc", Values: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Meta.Total != 1 || len(result.Items) != 1 || result.Items[0].UserName != "Audit User" {
		t.Fatalf("unexpected audit result after user deletion: %+v", result)
	}
}

func TestSessionRotationIntegration(t *testing.T) {
	f := newTransactionFixture(t)
	q := pgx.Identifier{f.schema}.Sanitize()
	var userID int64
	if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`
		INSERT INTO %s.users(name,email,password_hash,role)
		VALUES('Session User','session@example.com','unused','cashier')
		RETURNING id
	`, q)).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	repo := NewAuthRepository(f.db, f.schema)
	oldSessionID, err := repo.CreateSession(f.ctx, userID, "old-hash", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if active, err := repo.IsSessionActive(f.ctx, oldSessionID, userID); err != nil || !active {
		t.Fatalf("new session active=%v err=%v", active, err)
	}
	user, sessionID, err := repo.RotateSession(f.ctx, "old-hash", "new-hash", time.Now().Add(2*time.Hour))
	if err != nil || user.ID != userID {
		t.Fatalf("RotateSession() user=%+v err=%v", user, err)
	}
	if sessionID < 1 {
		t.Fatalf("RotateSession() sessionID=%d", sessionID)
	}
	if active, err := repo.IsSessionActive(f.ctx, oldSessionID, userID); err != nil || active {
		t.Fatalf("rotated old session active=%v err=%v", active, err)
	}
	if active, err := repo.IsSessionActive(f.ctx, sessionID, userID); err != nil || !active {
		t.Fatalf("rotated new session active=%v err=%v", active, err)
	}
	if _, _, err := repo.RotateSession(f.ctx, "old-hash", "reused-hash", time.Now().Add(time.Hour)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("reused refresh token error = %v, want ErrInvalidSession", err)
	}
	if err := repo.RevokeSession(f.ctx, "new-hash"); err != nil {
		t.Fatal(err)
	}
	if active, err := repo.IsSessionActive(f.ctx, sessionID, userID); err != nil || active {
		t.Fatalf("revoked session active=%v err=%v", active, err)
	}
	var activeSessions int
	if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.auth_sessions
		WHERE user_id=$1 AND revoked_at IS NULL
	`, q), userID).Scan(&activeSessions); err != nil {
		t.Fatal(err)
	}
	if activeSessions != 0 {
		t.Fatalf("active sessions = %d, want 0", activeSessions)
	}
}

const integrationTestSchemaPrefix = "go_pos_test_"

type transactionFixture struct {
	ctx        context.Context
	db         *pgxpool.Pool
	repository *CooperativeRepository
	schema     string
	customerID int64
	supplierID int64
	cashID     int64
	debtID     int64
	qrisID     int64
	itemID     int64
	pcsUnitID  int64
	boxUnitID  int64
}

func newTransactionFixture(t testing.TB) *transactionFixture {
	t.Helper()
	if os.Getenv("GO_POS_INTEGRATION_TESTS") != "1" {
		t.Skip("set GO_POS_INTEGRATION_TESTS=1 to run PostgreSQL integration tests")
	}
	_ = godotenv.Load("../../.env")
	cfg := config.Config{
		DBHost: os.Getenv("DB_HOST"), DBPort: os.Getenv("DB_PORT"),
		DBUser: os.Getenv("DB_USER"), DBPassword: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"), DBSSLMode: os.Getenv("DB_SSLMODE"),
	}
	if cfg.DBName != "playground" && cfg.DBName != "pos_playground" {
		t.Fatalf("refusing integration tests on database %q; expected playground or pos_playground", cfg.DBName)
	}

	ctx := context.Background()
	db, err := database.New(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to integration database: %v", err)
	}
	schema := fmt.Sprintf("%s%d", integrationTestSchemaPrefix, time.Now().UnixNano())
	if err := database.Migrate(ctx, db, schema); err != nil {
		db.Close()
		t.Fatalf("migrate isolated test schema: %v", err)
	}
	t.Cleanup(func() {
		defer db.Close()
		if !strings.HasPrefix(schema, integrationTestSchemaPrefix) {
			t.Errorf("refusing to clean unexpected schema %q", schema)
			return
		}
		if _, err := db.Exec(ctx, fmt.Sprintf("DROP SCHEMA %s CASCADE", pgx.Identifier{schema}.Sanitize())); err != nil {
			t.Errorf("drop isolated test schema: %v", err)
		}
	})

	f := &transactionFixture{
		ctx: ctx, db: db, repository: NewCooperativeRepository(db, schema), schema: schema,
	}
	q := pgx.Identifier{schema}.Sanitize()
	mustScan := func(query string, args []any, destination ...any) {
		t.Helper()
		if err := db.QueryRow(ctx, query, args...).Scan(destination...); err != nil {
			t.Fatalf("prepare integration fixture: %v", err)
		}
	}
	mustScan(fmt.Sprintf(`SELECT id FROM %s.customers WHERE code='UMUM'`, q), nil, &f.customerID)
	mustScan(fmt.Sprintf(`SELECT id FROM %s.payment_methods WHERE name='Tunai'`, q), nil, &f.cashID)
	mustScan(fmt.Sprintf(`SELECT id FROM %s.payment_methods WHERE name='Piutang'`, q), nil, &f.debtID)
	mustScan(fmt.Sprintf(`SELECT id FROM %s.payment_methods WHERE name='QRIS Dummy'`, q), nil, &f.qrisID)
	mustScan(fmt.Sprintf(`SELECT id FROM %s.units WHERE name='Pcs'`, q), nil, &f.pcsUnitID)
	mustScan(fmt.Sprintf(`SELECT id FROM %s.units WHERE name='Box'`, q), nil, &f.boxUnitID)
	mustScan(fmt.Sprintf(`INSERT INTO %s.suppliers(code,name,phone,address) VALUES('SUP-TEST','Supplier Test','','') RETURNING id`, q), nil, &f.supplierID)
	mustScan(fmt.Sprintf(`INSERT INTO %s.items(supplier_id,sku,name,unit_id,base_unit_id,units_per_package,stock,price,cost,retail_price,retail_cost) VALUES($1,'ITEM-TEST','Barang Test',$2,$2,1,10,1000,500,1000,500) RETURNING id`, q), []any{f.supplierID, f.pcsUnitID}, &f.itemID)
	return f
}

func (f *transactionFixture) request(kind string, quantity int, paid int64, debt bool) entity.CreateTransactionRequest {
	methodID := f.cashID
	if debt {
		methodID = f.debtID
	}
	request := entity.CreateTransactionRequest{
		TransactionType: kind, PaymentMethodID: &methodID, PaidAmount: paid,
		Items: []entity.TransactionLine{{ItemID: f.itemID, Quantity: quantity}},
	}
	if kind == "SALE" {
		request.CustomerID = &f.customerID
	} else {
		request.SupplierID = &f.supplierID
	}
	return request
}

func (f *transactionFixture) stock(t testing.TB) int {
	t.Helper()
	var stock int
	if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT stock FROM %s.items WHERE id=$1`, pgx.Identifier{f.schema}.Sanitize()), f.itemID).Scan(&stock); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	return stock
}

func TestTransactionStockIntegration(t *testing.T) {
	t.Run("package and retail sale share base stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		q := pgx.Identifier{f.schema}.Sanitize()
		if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`UPDATE %s.items SET unit_id=$1,base_unit_id=$2,units_per_package=12,allow_retail=TRUE,stock=24,price=10000,cost=8000,retail_price=1000,retail_cost=700 WHERE id=$3`, q), f.boxUnitID, f.pcsUnitID, f.itemID); err != nil {
			t.Fatal(err)
		}
		request := f.request("SALE", 1, 11000, false)
		request.Items[0].UnitID = &f.boxUnitID
		request.Items = append(request.Items, entity.TransactionLine{ItemID: f.itemID, UnitID: &f.pcsUnitID, Quantity: 1})
		if _, err := f.repository.CreateTransaction(f.ctx, request); err != nil {
			t.Fatalf("create mixed-unit sale: %v", err)
		}
		if got := f.stock(t); got != 11 {
			t.Fatalf("stock = %d, want 11 base units", got)
		}
	})

	t.Run("sale decreases stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 3000, false)); err != nil {
			t.Fatalf("create sale: %v", err)
		}
		if got := f.stock(t); got != 7 {
			t.Fatalf("stock = %d, want 7", got)
		}
	})
	t.Run("purchase increases stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("PURCHASE", 4, 2000, false)); err != nil {
			t.Fatalf("create purchase: %v", err)
		}
		if got := f.stock(t); got != 14 {
			t.Fatalf("stock = %d, want 14", got)
		}
	})
	t.Run("insufficient sale rolls back transaction and stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 11, 11000, false)); err == nil {
			t.Fatal("expected insufficient-stock error")
		}
		if got := f.stock(t); got != 10 {
			t.Fatalf("stock = %d after rollback, want 10", got)
		}
		var count int
		_ = f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.transactions`, pgx.Identifier{f.schema}.Sanitize())).Scan(&count)
		if count != 0 {
			t.Fatalf("transaction count = %d after rollback, want 0", count)
		}
	})
}

func TestTransactionUpdateAndVoidIntegration(t *testing.T) {
	t.Run("sale update recalculates stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		created, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 2, 2000, false))
		if err != nil {
			t.Fatalf("create sale: %v", err)
		}
		if _, err := f.repository.UpdateTransaction(f.ctx, created.ID, f.request("SALE", 5, 5000, false)); err != nil {
			t.Fatalf("update sale: %v", err)
		}
		if got := f.stock(t); got != 5 {
			t.Fatalf("stock = %d, want 5", got)
		}
	})
	t.Run("failed update restores original transaction and stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		created, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 2, 2000, false))
		if err != nil {
			t.Fatalf("create sale: %v", err)
		}
		if _, err := f.repository.UpdateTransaction(f.ctx, created.ID, f.request("SALE", 11, 11000, false)); err == nil {
			t.Fatal("expected insufficient-stock error")
		}
		if got := f.stock(t); got != 8 {
			t.Fatalf("stock = %d after failed update, want 8", got)
		}
		var quantity int
		_ = f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT quantity FROM %s.transaction_items WHERE transaction_id=$1`, pgx.Identifier{f.schema}.Sanitize()), created.ID).Scan(&quantity)
		if quantity != 2 {
			t.Fatalf("stored quantity = %d after failed update, want 2", quantity)
		}
	})
	t.Run("void restores sale stock", func(t *testing.T) {
		f := newTransactionFixture(t)
		created, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 3000, false))
		if err != nil {
			t.Fatalf("create sale: %v", err)
		}
		if err := f.repository.VoidTransaction(f.ctx, created.ID, "test void"); err != nil {
			t.Fatalf("void sale: %v", err)
		}
		if got := f.stock(t); got != 10 {
			t.Fatalf("stock = %d, want 10", got)
		}
	})
	t.Run("void purchase is blocked when its stock was consumed", func(t *testing.T) {
		f := newTransactionFixture(t)
		purchase, err := f.repository.CreateTransaction(f.ctx, f.request("PURCHASE", 4, 2000, false))
		if err != nil {
			t.Fatalf("create purchase: %v", err)
		}
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 12, 12000, false)); err != nil {
			t.Fatalf("consume purchased stock: %v", err)
		}
		if err := f.repository.VoidTransaction(f.ctx, purchase.ID, "test void"); err == nil {
			t.Fatal("expected void to be rejected")
		}
		if got := f.stock(t); got != 2 {
			t.Fatalf("stock = %d after rejected void, want 2", got)
		}
	})
}

func TestDebtPaymentIntegration(t *testing.T) {
	t.Run("partial and final payments stay consistent", func(t *testing.T) {
		f := newTransactionFixture(t)
		transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 1000, true))
		if err != nil {
			t.Fatalf("create credit sale: %v", err)
		}
		var debtID, remaining int64
		var status string
		q := pgx.Identifier{f.schema}.Sanitize()
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT id,remaining_amount,status FROM %s.debts WHERE transaction_id=$1`, q), transaction.ID).Scan(&debtID, &remaining, &status); err != nil {
			t.Fatalf("read debt: %v", err)
		}
		if remaining != 2000 || status != "OPEN" {
			t.Fatalf("initial debt = %d/%s, want 2000/OPEN", remaining, status)
		}
		if err := f.repository.PayDebt(f.ctx, debtID, 500, "partial"); err != nil {
			t.Fatalf("partial payment: %v", err)
		}
		if err := f.repository.PayDebt(f.ctx, debtID, 1500, "final"); err != nil {
			t.Fatalf("final payment: %v", err)
		}
		var paid int64
		var paymentStatus, debtStatus string
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT d.remaining_amount,d.status,t.paid_amount,t.payment_status FROM %s.debts d JOIN %s.transactions t ON t.id=d.transaction_id WHERE d.id=$1`, q, q), debtID).Scan(&remaining, &debtStatus, &paid, &paymentStatus); err != nil {
			t.Fatalf("read paid debt: %v", err)
		}
		if remaining != 0 || debtStatus != "PAID" || paid != 3000 || paymentStatus != "PAID" {
			t.Fatalf("final state debt=%d/%s transaction=%d/%s", remaining, debtStatus, paid, paymentStatus)
		}
	})
	t.Run("overpayment rolls back", func(t *testing.T) {
		f := newTransactionFixture(t)
		transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 1000, true))
		if err != nil {
			t.Fatalf("create credit sale: %v", err)
		}
		q := pgx.Identifier{f.schema}.Sanitize()
		var debtID int64
		_ = f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT id FROM %s.debts WHERE transaction_id=$1`, q), transaction.ID).Scan(&debtID)
		if err := f.repository.PayDebt(f.ctx, debtID, 2001, "too much"); err == nil {
			t.Fatal("expected overpayment error")
		}
		var remaining, payments int64
		_ = f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT remaining_amount FROM %s.debts WHERE id=$1`, q), debtID).Scan(&remaining)
		_ = f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.debt_payments WHERE debt_id=$1`, q), debtID).Scan(&payments)
		if remaining != 2000 || payments != 0 {
			t.Fatalf("state after overpayment: remaining=%d payments=%d", remaining, payments)
		}
	})
	t.Run("reversal restores balances once and keeps its trace", func(t *testing.T) {
		f := newTransactionFixture(t)
		transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 1000, true))
		if err != nil {
			t.Fatalf("create credit sale: %v", err)
		}
		q := pgx.Identifier{f.schema}.Sanitize()
		var debtID int64
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT id FROM %s.debts WHERE transaction_id=$1`, q), transaction.ID).Scan(&debtID); err != nil {
			t.Fatalf("read debt: %v", err)
		}
		if err := f.repository.PayDebt(f.ctx, debtID, 500, "wrong amount"); err != nil {
			t.Fatalf("pay debt: %v", err)
		}
		var paymentID int64
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT id FROM %s.debt_payments WHERE debt_id=$1`, q), debtID).Scan(&paymentID); err != nil {
			t.Fatalf("read payment: %v", err)
		}
		if err := f.repository.ReverseDebtPayment(f.ctx, debtID, paymentID, 7, "Admin Test", "nominal salah input"); err != nil {
			t.Fatalf("reverse payment: %v", err)
		}

		var remaining, paid int64
		var debtStatus, paymentStatus string
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT d.remaining_amount,d.status,t.paid_amount,t.payment_status FROM %s.debts d JOIN %s.transactions t ON t.id=d.transaction_id WHERE d.id=$1`, q, q), debtID).Scan(&remaining, &debtStatus, &paid, &paymentStatus); err != nil {
			t.Fatalf("read reversed balances: %v", err)
		}
		if remaining != 2000 || debtStatus != "OPEN" || paid != 1000 || paymentStatus != "PARTIAL" {
			t.Fatalf("reversed state debt=%d/%s transaction=%d/%s", remaining, debtStatus, paid, paymentStatus)
		}
		history, err := f.repository.DebtPayments(f.ctx, debtID)
		if err != nil {
			t.Fatalf("read history: %v", err)
		}
		if len(history) != 1 || history[0].ReversedAt == nil || history[0].ReversalReason != "nominal salah input" || history[0].ReversedByName != "Admin Test" {
			t.Fatalf("unexpected reversal trace: %+v", history)
		}
		if err := f.repository.ReverseDebtPayment(f.ctx, debtID, paymentID, 7, "Admin Test", "coba reversal lagi"); err == nil {
			t.Fatal("expected duplicate reversal to be rejected")
		}
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT remaining_amount FROM %s.debts WHERE id=$1`, q), debtID).Scan(&remaining); err != nil {
			t.Fatalf("read balance after duplicate reversal: %v", err)
		}
		if remaining != 2000 {
			t.Fatalf("remaining after duplicate reversal = %d, want 2000", remaining)
		}
	})
	t.Run("payment prevents editing and voiding credit sale", func(t *testing.T) {
		f := newTransactionFixture(t)
		transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 1000, true))
		if err != nil {
			t.Fatalf("create credit sale: %v", err)
		}
		q := pgx.Identifier{f.schema}.Sanitize()
		var debtID int64
		_ = f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT id FROM %s.debts WHERE transaction_id=$1`, q), transaction.ID).Scan(&debtID)
		if err := f.repository.PayDebt(f.ctx, debtID, 500, "partial"); err != nil {
			t.Fatalf("pay debt: %v", err)
		}
		if _, err := f.repository.UpdateTransaction(f.ctx, transaction.ID, f.request("SALE", 2, 2000, false)); err == nil {
			t.Fatal("expected update to be rejected after debt payment")
		}
		if err := f.repository.VoidTransaction(f.ctx, transaction.ID, "test"); err == nil {
			t.Fatal("expected void to be rejected after debt payment")
		}
		if got := f.stock(t); got != 7 {
			t.Fatalf("stock = %d after rejected operations, want 7", got)
		}
	})
}

func TestBulkOperationalActionsIntegration(t *testing.T) {
	t.Run("settlement creates individual payment history", func(t *testing.T) {
		f := newTransactionFixture(t)
		transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 1000, true))
		if err != nil {
			t.Fatalf("create credit sale: %v", err)
		}
		q := pgx.Identifier{f.schema}.Sanitize()
		var debtID int64
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT id FROM %s.debts WHERE transaction_id=$1`, q), transaction.ID).Scan(&debtID); err != nil {
			t.Fatalf("read debt: %v", err)
		}
		result := f.repository.BulkSettleDebts(f.ctx, []int64{debtID})
		if len(result.Results) != 1 || !result.Results[0].Success {
			t.Fatalf("bulk settlement result: %+v", result)
		}
		var remaining, payments int64
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT remaining_amount,(SELECT COUNT(*) FROM %s.debt_payments WHERE debt_id=$1) FROM %s.debts WHERE id=$1`, q, q), debtID).Scan(&remaining, &payments); err != nil {
			t.Fatalf("read settlement state: %v", err)
		}
		if remaining != 0 || payments != 1 {
			t.Fatalf("remaining=%d payments=%d, want 0/1", remaining, payments)
		}
	})
	t.Run("stock reset creates traceable movement", func(t *testing.T) {
		f := newTransactionFixture(t)
		result := f.repository.BulkResetStock(f.ctx, []int64{f.itemID})
		if len(result.Results) != 1 || !result.Results[0].Success {
			t.Fatalf("bulk reset result: %+v", result)
		}
		q := pgx.Identifier{f.schema}.Sanitize()
		var stock, before, change, after int
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT i.stock,m.quantity_before,m.quantity_change,m.quantity_after FROM %s.items i JOIN %s.stock_movements m ON m.item_id=i.id WHERE i.id=$1`, q, q), f.itemID).Scan(&stock, &before, &change, &after); err != nil {
			t.Fatalf("read stock movement: %v", err)
		}
		if stock != 0 || before != 10 || change != -10 || after != 0 {
			t.Fatalf("stock movement = %d/%d/%d/%d", stock, before, change, after)
		}
	})
}

func TestStockMovementTraceIntegration(t *testing.T) {
	f := newTransactionFixture(t)
	transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 3, 3000, false))
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}
	movements, err := f.repository.StockMovements(f.ctx, f.itemID)
	if err != nil {
		t.Fatalf("read sale movement: %v", err)
	}
	if len(movements) != 1 || movements[0].MovementType != "SALE" || movements[0].QuantityBefore != 10 || movements[0].QuantityChange != -3 || movements[0].QuantityAfter != 7 {
		t.Fatalf("unexpected sale movement: %+v", movements)
	}
	if err := f.repository.VoidTransaction(f.ctx, transaction.ID, "test trace"); err != nil {
		t.Fatalf("void sale: %v", err)
	}
	movements, err = f.repository.StockMovements(f.ctx, f.itemID)
	if err != nil {
		t.Fatalf("read void movement: %v", err)
	}
	if len(movements) != 2 || movements[0].MovementType != "TRANSACTION_VOID" || movements[0].QuantityBefore != 7 || movements[0].QuantityChange != 3 || movements[0].QuantityAfter != 10 {
		t.Fatalf("unexpected void movement: %+v", movements)
	}
}

func TestConcurrentCheckoutInvoiceAndStockIntegration(t *testing.T) {
	f := newTransactionFixture(t)
	q := pgx.Identifier{f.schema}.Sanitize()
	const checkouts = 25
	if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`UPDATE %s.items SET stock=$1 WHERE id=$2`, q), checkouts, f.itemID); err != nil {
		t.Fatalf("prepare checkout stock: %v", err)
	}

	start := make(chan struct{})
	results := make(chan entity.Transaction, checkouts)
	errors := make(chan error, checkouts)
	var workers sync.WaitGroup
	for i := 0; i < checkouts; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			transaction, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 1, 1000, false))
			if err != nil {
				errors <- err
				return
			}
			results <- transaction
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	close(errors)

	for err := range errors {
		t.Errorf("concurrent checkout: %v", err)
	}
	if t.Failed() {
		return
	}
	invoices := make(map[string]struct{}, checkouts)
	for transaction := range results {
		if _, duplicate := invoices[transaction.InvoiceNo]; duplicate {
			t.Errorf("duplicate invoice number %q", transaction.InvoiceNo)
		}
		invoices[transaction.InvoiceNo] = struct{}{}
	}
	if len(invoices) != checkouts {
		t.Fatalf("unique invoices = %d, want %d", len(invoices), checkouts)
	}
	if stock := f.stock(t); stock != 0 {
		t.Fatalf("stock after concurrent checkout = %d, want 0", stock)
	}
}

func TestIdempotentCheckoutIntegration(t *testing.T) {
	t.Run("same request only mutates stock once", func(t *testing.T) {
		f := newTransactionFixture(t)
		request := f.request("SALE", 2, 2000, false)
		request.IdempotencyKey = "checkout-retry-001"

		first, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatalf("first checkout: %v", err)
		}
		replay, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatalf("replayed checkout: %v", err)
		}
		if replay.ID != first.ID || replay.InvoiceNo != first.InvoiceNo || !replay.IdempotencyReplay {
			t.Fatalf("replay = %+v, want original transaction %+v marked as replay", replay, first)
		}
		if stock := f.stock(t); stock != 8 {
			t.Fatalf("stock after replay = %d, want 8", stock)
		}
	})

	t.Run("same key rejects different request", func(t *testing.T) {
		f := newTransactionFixture(t)
		first := f.request("SALE", 1, 1000, false)
		first.IdempotencyKey = "checkout-conflict-001"
		if _, err := f.repository.CreateTransaction(f.ctx, first); err != nil {
			t.Fatalf("first checkout: %v", err)
		}
		changed := f.request("SALE", 2, 2000, false)
		changed.IdempotencyKey = first.IdempotencyKey
		if _, err := f.repository.CreateTransaction(f.ctx, changed); !errors.Is(err, ErrIdempotencyConflict) {
			t.Fatalf("conflicting checkout error = %v, want ErrIdempotencyConflict", err)
		}
		if stock := f.stock(t); stock != 9 {
			t.Fatalf("stock after conflict = %d, want 9", stock)
		}
	})

	t.Run("concurrent retries create one transaction", func(t *testing.T) {
		f := newTransactionFixture(t)
		const retries = 10
		start := make(chan struct{})
		results := make(chan entity.Transaction, retries)
		errors := make(chan error, retries)
		var workers sync.WaitGroup
		for i := 0; i < retries; i++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				request := f.request("SALE", 1, 1000, false)
				request.IdempotencyKey = "checkout-concurrent-001"
				<-start
				transaction, err := f.repository.CreateTransaction(f.ctx, request)
				if err != nil {
					errors <- err
					return
				}
				results <- transaction
			}()
		}
		close(start)
		workers.Wait()
		close(results)
		close(errors)
		for err := range errors {
			t.Errorf("concurrent retry: %v", err)
		}
		var transactionID int64
		for transaction := range results {
			if transactionID == 0 {
				transactionID = transaction.ID
			}
			if transaction.ID != transactionID {
				t.Errorf("transaction ID = %d, want replay of %d", transaction.ID, transactionID)
			}
		}
		if stock := f.stock(t); stock != 9 {
			t.Fatalf("stock after concurrent retries = %d, want 9", stock)
		}
	})
}

func TestDummyAsyncPaymentIntegration(t *testing.T) {
	t.Run("pending reservation is fulfilled exactly once", func(t *testing.T) {
		f := newTransactionFixture(t)
		request := f.request("SALE", 2, 0, false)
		request.PaymentMethodID = &f.qrisID
		request.PaymentProvider = "DUMMY"
		request.IdempotencyKey = "dummy-payment-001"
		transaction, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatalf("create async checkout: %v", err)
		}
		if transaction.Payment == nil || transaction.Payment.Status != "PENDING" {
			t.Fatalf("payment = %+v, want PENDING", transaction.Payment)
		}
		history, err := f.repository.Transactions(f.ctx, "SALE")
		if err != nil || len(history) != 1 || history[0].Payment == nil || history[0].Payment.Status != "PENDING" {
			t.Fatalf("payment was not hydrated in transaction history: history=%+v err=%v", history, err)
		}
		if stock := f.stock(t); stock != 10 {
			t.Fatalf("stock while pending = %d, want 10", stock)
		}
		if _, err := f.repository.UpdateTransaction(f.ctx, transaction.ID, f.request("SALE", 1, 1000, false)); err == nil {
			t.Fatal("expected provider transaction edit to be rejected")
		}
		if err := f.repository.VoidTransaction(f.ctx, transaction.ID, "direct void"); err == nil {
			t.Fatal("expected direct provider transaction void to be rejected")
		}
		blocked := f.request("SALE", 9, 9000, false)
		if _, err := f.repository.CreateTransaction(f.ctx, blocked); err == nil {
			t.Fatal("expected reserved stock to block another sale")
		}
		payment, err := f.repository.SetDummyPaymentStatus(f.ctx, transaction.Payment.ID, "PAID")
		if err != nil || payment.Status != "PAID" {
			t.Fatalf("pay dummy payment: payment=%+v err=%v", payment, err)
		}
		if _, err := f.repository.SetDummyPaymentStatus(f.ctx, transaction.Payment.ID, "PAID"); err != nil {
			t.Fatalf("replay paid callback: %v", err)
		}
		if stock := f.stock(t); stock != 8 {
			t.Fatalf("stock after paid callback replay = %d, want 8", stock)
		}
	})

	t.Run("failed payment releases reservation", func(t *testing.T) {
		f := newTransactionFixture(t)
		request := f.request("SALE", 10, 0, false)
		request.PaymentMethodID = &f.qrisID
		request.PaymentProvider = "DUMMY"
		transaction, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.repository.SetDummyPaymentStatus(f.ctx, transaction.Payment.ID, "FAILED"); err != nil {
			t.Fatalf("fail dummy payment: %v", err)
		}
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 10, 10000, false)); err != nil {
			t.Fatalf("sale after reservation release: %v", err)
		}
		if stock := f.stock(t); stock != 0 {
			t.Fatalf("stock after released reservation sale = %d, want 0", stock)
		}
	})

	t.Run("late paid callback persists expiry and releases reservation", func(t *testing.T) {
		f := newTransactionFixture(t)
		request := f.request("SALE", 10, 0, false)
		request.PaymentMethodID = &f.qrisID
		request.PaymentProvider = "DUMMY"
		transaction, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`UPDATE %s.payments SET expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, pgx.Identifier{f.schema}.Sanitize()), transaction.Payment.ID); err != nil {
			t.Fatalf("expire payment fixture: %v", err)
		}

		payment, err := f.repository.SetDummyPaymentStatus(f.ctx, transaction.Payment.ID, "PAID")
		if !errors.Is(err, ErrPaymentExpired) || payment.Status != "EXPIRED" {
			t.Fatalf("late callback: payment=%+v err=%v", payment, err)
		}
		persisted, err := f.repository.DummyPayment(f.ctx, transaction.Payment.ID)
		if err != nil || persisted.Status != "EXPIRED" {
			t.Fatalf("persisted payment: payment=%+v err=%v", persisted, err)
		}
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 10, 10000, false)); err != nil {
			t.Fatalf("sale after automatic expiry: %v", err)
		}
	})

	t.Run("status read persists expired pending payment", func(t *testing.T) {
		f := newTransactionFixture(t)
		request := f.request("SALE", 2, 0, false)
		request.PaymentMethodID = &f.qrisID
		request.PaymentProvider = "DUMMY"
		transaction, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`UPDATE %s.payments SET expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, pgx.Identifier{f.schema}.Sanitize()), transaction.Payment.ID); err != nil {
			t.Fatalf("expire payment fixture: %v", err)
		}

		payment, err := f.repository.DummyPayment(f.ctx, transaction.Payment.ID)
		if err != nil || payment.Status != "EXPIRED" {
			t.Fatalf("status read: payment=%+v err=%v", payment, err)
		}
		var transactionStatus, reservationStatus string
		q := pgx.Identifier{f.schema}.Sanitize()
		if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT t.status,r.status FROM %s.transactions t JOIN %s.payments p ON p.transaction_id=t.id JOIN %s.stock_reservations r ON r.payment_id=p.id WHERE p.id=$1`, q, q, q), transaction.Payment.ID).Scan(&transactionStatus, &reservationStatus); err != nil {
			t.Fatal(err)
		}
		if transactionStatus != "VOID" || reservationStatus != "EXPIRED" {
			t.Fatalf("expired state transaction=%s reservation=%s", transactionStatus, reservationStatus)
		}
	})

	t.Run("concurrent paid callbacks only deduct once", func(t *testing.T) {
		f := newTransactionFixture(t)
		request := f.request("SALE", 3, 0, false)
		request.PaymentMethodID = &f.qrisID
		request.PaymentProvider = "DUMMY"
		transaction, err := f.repository.CreateTransaction(f.ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		const callbacks = 10
		start := make(chan struct{})
		errors := make(chan error, callbacks)
		var workers sync.WaitGroup
		for i := 0; i < callbacks; i++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				_, err := f.repository.SetDummyPaymentStatus(f.ctx, transaction.Payment.ID, "PAID")
				errors <- err
			}()
		}
		close(start)
		workers.Wait()
		close(errors)
		for err := range errors {
			if err != nil {
				t.Errorf("paid callback: %v", err)
			}
		}
		if stock := f.stock(t); stock != 7 {
			t.Fatalf("stock after concurrent callbacks = %d, want 7", stock)
		}
	})
}

func TestVersionedMigrationIntegration(t *testing.T) {
	f := newTransactionFixture(t)
	if err := database.Migrate(f.ctx, f.db, f.schema); err != nil {
		t.Fatalf("rerun migrations: %v", err)
	}
	q := pgx.Identifier{f.schema}.Sanitize()
	var count int
	var name string
	if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.schema_migrations`, q)).Scan(&count); err != nil {
		t.Fatalf("read migration ledger: %v", err)
	}
	if err := f.db.QueryRow(f.ctx, fmt.Sprintf(`SELECT name FROM %s.schema_migrations ORDER BY version DESC LIMIT 1`, q)).Scan(&name); err != nil {
		t.Fatalf("read latest migration: %v", err)
	}
	if count != 4 || name != "asynchronous payments and stock reservations" {
		t.Fatalf("migration ledger = %d/%q, want 4/asynchronous payments and stock reservations", count, name)
	}
}

func TestPaginationIntegration(t *testing.T) {
	t.Run("customers include stable metadata and page boundaries", func(t *testing.T) {
		f := newTransactionFixture(t)
		q := pgx.Identifier{f.schema}.Sanitize()
		for i := 1; i <= 4; i++ {
			if _, err := f.db.Exec(f.ctx, fmt.Sprintf(`INSERT INTO %s.customers(code,name,customer_type) VALUES($1,$2,'MEMBER')`, q), fmt.Sprintf("C-%02d", i), fmt.Sprintf("Customer %02d", i)); err != nil {
				t.Fatalf("insert customer %d: %v", i, err)
			}
		}
		result, err := f.repository.CustomersPage(f.ctx, pagination.Params{Page: 2, PerPage: 2})
		if err != nil {
			t.Fatalf("get customer page: %v", err)
		}
		if result.Meta.Total != 5 || result.Meta.TotalPages != 3 || len(result.Items) != 2 {
			t.Fatalf("page result = items:%d meta:%+v, want 2 items and 5 total across 3 pages", len(result.Items), result.Meta)
		}
		if result.Items[0].Code != "C-02" || result.Items[1].Code != "C-03" {
			t.Fatalf("page codes = %s, %s; want C-02, C-03", result.Items[0].Code, result.Items[1].Code)
		}
	})

	t.Run("transaction type filter is counted before pagination", func(t *testing.T) {
		f := newTransactionFixture(t)
		for i := 0; i < 3; i++ {
			if _, err := f.repository.CreateTransaction(f.ctx, f.request("SALE", 1, 1000, false)); err != nil {
				t.Fatalf("create sale %d: %v", i, err)
			}
		}
		if _, err := f.repository.CreateTransaction(f.ctx, f.request("PURCHASE", 1, 500, false)); err != nil {
			t.Fatalf("create purchase: %v", err)
		}
		result, err := f.repository.TransactionsPage(f.ctx, "SALE", pagination.Params{Page: 2, PerPage: 2})
		if err != nil {
			t.Fatalf("get transaction page: %v", err)
		}
		if result.Meta.Total != 3 || result.Meta.TotalPages != 2 || len(result.Items) != 1 {
			t.Fatalf("page result = items:%d meta:%+v, want one item on page 2 and three filtered total", len(result.Items), result.Meta)
		}
		if len(result.Items[0].Items) != 1 || result.Items[0].TransactionType != "SALE" {
			t.Fatalf("paged transaction was not fully hydrated: %+v", result.Items[0])
		}
	})
}
