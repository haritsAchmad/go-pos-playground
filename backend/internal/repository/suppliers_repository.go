package repository

import (
	"context"
	"errors"
	"fmt"

	dto "go-pos-playground/internal/dto/suppliers"
	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/pagination"

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
	return r.find(ctx, "", nil)
}

func (r *SupplierRepository) FindPage(ctx context.Context, params pagination.Params) (pagination.Result[entity.Suppliers], error) {
	var total int64
	if err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.suppliers WHERE deleted_at IS NULL`, r.schema)).Scan(&total); err != nil {
		return pagination.Result[entity.Suppliers]{}, err
	}
	suppliers, err := r.find(ctx, " LIMIT $1 OFFSET $2", []any{params.PerPage, params.Offset()})
	if err != nil {
		return pagination.Result[entity.Suppliers]{}, err
	}
	return pagination.NewResult(suppliers, params, total), nil
}

func (r *SupplierRepository) find(ctx context.Context, suffix string, args []any) ([]entity.Suppliers, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			code,
			name,
			phone,
			address,
			created_at,
			updated_at
		FROM %s.suppliers
		WHERE deleted_at IS NULL
		ORDER BY id ASC%s
	`, r.schema, suffix)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers := make([]entity.Suppliers, 0)

	for rows.Next() {
		var supplier entity.Suppliers

		if err := rows.Scan(
			&supplier.ID,
			&supplier.Code,
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
			code,
			name,
			phone,
			address
		)
		VALUES
		(
			$1,
			$2,
			$3,
			$4
		)
	`, r.schema)

	_, err := r.db.Exec(
		ctx,
		query,
		req.Code,
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
			code,
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
		&supplier.Code,
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
			code = $1,
			name = $2,
			phone = $3,
			address = $4,
			updated_at = NOW()
		WHERE
			id = $5
			AND deleted_at IS NULL
	`, r.schema)

	commandTag, err := r.db.Exec(
		ctx,
		query,
		req.Code,
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
