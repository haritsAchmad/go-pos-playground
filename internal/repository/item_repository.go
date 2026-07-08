package repository

import (
	"context"
	"fmt"

	dto "go-inventory-playground/internal/dto/items"
	"go-inventory-playground/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository struct {
	db     *pgxpool.Pool
	schema string
}

func NewItemRepository(db *pgxpool.Pool, schema string) *ItemRepository {
	return &ItemRepository{
		db:     db,
		schema: schema,
	}
}

func (r *ItemRepository) FindAll(ctx context.Context) ([]entity.Item, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			description,
			stock,
			created_at,
			updated_at
		FROM %s.items
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`, r.schema)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.Item, 0)

	for rows.Next() {
		var item entity.Item

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Stock,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ItemRepository) Create(
	ctx context.Context,
	req dto.CreateItemRequest,
) error {

	query := fmt.Sprintf(`
		INSERT INTO %s.items
		(
			name,
			description,
			stock
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
		req.Description,
		req.Stock,
	)

	return err
}
