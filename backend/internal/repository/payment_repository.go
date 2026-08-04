package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-pos-playground/internal/entity"

	"github.com/jackc/pgx/v5"
)

func (r *CooperativeRepository) SetDummyPaymentStatus(ctx context.Context, paymentID int64, desired string) (entity.Payment, error) {
	if desired != "PAID" && desired != "FAILED" && desired != "EXPIRED" {
		return entity.Payment{}, errors.New("payment status must be PAID, FAILED, or EXPIRED")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return entity.Payment{}, err
	}
	defer tx.Rollback(ctx)

	var payment entity.Payment
	var invoice string
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT p.id,p.transaction_id,p.provider,p.external_reference,p.status,p.amount,
			p.paid_at,p.expires_at,p.created_at,t.invoice_no
		FROM %s.payments p JOIN %s.transactions t ON t.id=p.transaction_id
		WHERE p.id=$1 FOR UPDATE OF p,t
	`, r.schema, r.schema), paymentID).Scan(
		&payment.ID, &payment.TransactionID, &payment.Provider, &payment.ExternalReference,
		&payment.Status, &payment.Amount, &payment.PaidAt, &payment.ExpiresAt,
		&payment.CreatedAt, &invoice,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Payment{}, errors.New("payment not found")
	}
	if err != nil {
		return entity.Payment{}, err
	}
	if payment.Provider != "DUMMY" {
		return entity.Payment{}, errors.New("payment is not handled by dummy provider")
	}
	if payment.Status == desired {
		return payment, nil
	}
	if payment.Status != "PENDING" {
		return entity.Payment{}, fmt.Errorf("payment is already %s", payment.Status)
	}
	if desired == "PAID" && time.Now().After(payment.ExpiresAt) {
		return entity.Payment{}, errors.New("expired payment cannot be paid")
	}

	reservationStatus := map[string]string{"PAID": "FULFILLED", "FAILED": "RELEASED", "EXPIRED": "EXPIRED"}[desired]
	if desired == "PAID" {
		rows, err := tx.Query(ctx, fmt.Sprintf(`SELECT item_id,quantity FROM %s.stock_reservations WHERE payment_id=$1 AND status='ACTIVE' ORDER BY item_id`, r.schema), payment.ID)
		if err != nil {
			return entity.Payment{}, err
		}
		type reservation struct {
			itemID   int64
			quantity int
		}
		var reservations []reservation
		for rows.Next() {
			var reservation reservation
			if err := rows.Scan(&reservation.itemID, &reservation.quantity); err != nil {
				rows.Close()
				return entity.Payment{}, err
			}
			reservations = append(reservations, reservation)
		}
		rows.Close()
		for _, reservation := range reservations {
			var stock int
			if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT stock FROM %s.items WHERE id=$1 FOR UPDATE`, r.schema), reservation.itemID).Scan(&stock); err != nil {
				return entity.Payment{}, err
			}
			if stock < reservation.quantity {
				return entity.Payment{}, errors.New("reserved stock is no longer available")
			}
			if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.items SET stock=stock-$1,updated_at=NOW() WHERE id=$2`, r.schema), reservation.quantity, reservation.itemID); err != nil {
				return entity.Payment{}, err
			}
			if err := r.recordStockMovement(ctx, tx, reservation.itemID, "SALE", stock, -reservation.quantity, invoice); err != nil {
				return entity.Payment{}, err
			}
		}
		paidAt := time.Now()
		payment.PaidAt = &paidAt
		if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.transactions SET payment_status='PAID',paid_amount=grand_total,amount_received=grand_total WHERE id=$1`, r.schema), payment.TransactionID); err != nil {
			return entity.Payment{}, err
		}
	} else {
		reason := "Pembayaran dummy " + desired
		if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.transactions SET status='VOID',void_reason=$2,voided_at=NOW() WHERE id=$1`, r.schema), payment.TransactionID, reason); err != nil {
			return entity.Payment{}, err
		}
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.stock_reservations SET status=$2 WHERE payment_id=$1 AND status='ACTIVE'`, r.schema), payment.ID, reservationStatus); err != nil {
		return entity.Payment{}, err
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.payments SET status=$2,paid_at=$3,updated_at=NOW() WHERE id=$1`, r.schema), payment.ID, desired, payment.PaidAt); err != nil {
		return entity.Payment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return entity.Payment{}, err
	}
	payment.Status = desired
	return payment, nil
}
