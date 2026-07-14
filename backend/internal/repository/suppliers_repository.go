package repository

import (
	"context"
	"errors"
	"fmt"

	dto "go-inventory-playground/internal/dto/suppliers"
	"go-inventory-playground/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SupplierRepository struct {
	db     *pgxpool.Pool
	schema string
}

var ErrSupplierNotFound = errors.New("supplier not found")

func NewSupplierRepository(db *pgxpool.Pool, schema string) *SupplierRepository {
	return &SupplierRepository{
		db:     db,
		schema: schema,
	}
}

func (r *SupplierRepository) FindAll(ctx context.Context) ([]entity.Suppliers, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			phone,
			address,
			created_at,
			updated_at
		FROM %s.suppliers
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`, r.schema)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers := make([]entity.Suppliers, 0)

	for rows.Next() {
		var supplier entity.Suppliers

		if err := rows.Scan(
			&supplier.ID,
			&supplier.Name,
			&supplier.Phone,
			&supplier.Address,
			&supplier.CreatedAt,
			&supplier.UpdatedAt,
		); err != nil {
			return nil, err
		}

		suppliers = append(suppliers, supplier)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}

func (r *SupplierRepository) Create(
	ctx context.Context,
	req dto.CreateSupplierRequest,
) error {

	query := fmt.Sprintf(`
		INSERT INTO %s.suppliers
		(
			name,
			phone,
			address
		)
		VALUES
		(
			$1,
			$2,
			$3
		)
	`, r.schema)

	_, err := r.db.Exec(
		ctx,
		query,
		req.Name,
		req.Phone,
		req.Address,
	)

	return err
}

func (r *SupplierRepository) FindByID(ctx context.Context, id int) (*entity.Suppliers, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			phone,
			address,
			created_at,
			updated_at
		FROM %s.suppliers
		WHERE id = $1
		AND deleted_at IS NULL
	`, r.schema)

	var supplier entity.Suppliers

	err := r.db.QueryRow(ctx, query, id).Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.Phone,
		&supplier.Address,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

func (r *SupplierRepository) Update(
	ctx context.Context,
	id int,
	req dto.UpdateSupplierRequest,
) error {
	query := fmt.Sprintf(`
		UPDATE %s.suppliers
		SET
			name = $1,
			phone = $2,
			address = $3,
			updated_at = NOW()
		WHERE
			id = $4
			AND deleted_at IS NULL
	`, r.schema)

	commandTag, err := r.db.Exec(
		ctx,
		query,
		req.Name,
		req.Phone,
		req.Address,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSupplierNotFound
	}

	return nil
}

func (r *SupplierRepository) Delete(
	ctx context.Context,
	id int,
) error {

	query := fmt.Sprintf(`
		UPDATE %s.suppliers
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`, r.schema)

	commandTag, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSupplierNotFound
	}

	return nil
}
