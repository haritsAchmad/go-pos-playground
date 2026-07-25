package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dto "go-pos-playground/internal/dto/items"
	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/pagination"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository struct {
	db     *pgxpool.Pool
	schema string
}

var ErrItemNotFound = errors.New("item not found")

func NewItemRepository(db *pgxpool.Pool, schema string) *ItemRepository {
	return &ItemRepository{
		db:     db,
		schema: schema,
	}
}

func (r *ItemRepository) FindAll(ctx context.Context) ([]entity.Items, error) {
	return r.FindAllQuery(ctx, defaultItemQuery())
}

func (r *ItemRepository) FindPage(ctx context.Context, params pagination.Params) (pagination.Result[entity.Items], error) {
	return r.FindPageQuery(ctx, params, defaultItemQuery())
}

func defaultItemQuery() listquery.Params {
	return listquery.Params{Sort: "id", Order: "asc", Values: map[string]string{}}
}

func (r *ItemRepository) FindAllQuery(ctx context.Context, query listquery.Params) ([]entity.Items, error) {
	where, order, args, err := itemQueryParts(query)
	if err != nil {
		return nil, err
	}
	return r.find(ctx, where, order, "", args)
}

func (r *ItemRepository) FindPageQuery(ctx context.Context, params pagination.Params, query listquery.Params) (pagination.Result[entity.Items], error) {
	where, order, args, err := itemQueryParts(query)
	if err != nil {
		return pagination.Result[entity.Items]{}, err
	}
	var total int64
	if err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.items i%s`, r.schema, where), args...).Scan(&total); err != nil {
		return pagination.Result[entity.Items]{}, err
	}
	paging := fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.PerPage, params.Offset())
	items, err := r.find(ctx, where, order, paging, args)
	if err != nil {
		return pagination.Result[entity.Items]{}, err
	}
	return pagination.NewResult(items, params, total), nil
}

func itemQueryParts(query listquery.Params) (string, string, []any, error) {
	clauses := []string{"i.deleted_at IS NULL"}
	args := make([]any, 0, 7)
	var minStock, maxStock int64
	var hasMinStock, hasMaxStock bool
	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if query.Search != "" {
		args = append(args, query.Search)
		position := len(args)
		clauses = append(clauses, fmt.Sprintf("(i.name ILIKE '%%' || $%d || '%%' OR COALESCE(i.sku,'') ILIKE '%%' || $%d || '%%' OR COALESCE(i.description,'') ILIKE '%%' || $%d || '%%')", position, position, position))
	}
	for _, key := range []string{"supplier_id", "category_id", "brand_id", "unit_id"} {
		if value, set, err := query.PositiveInt(key); err != nil {
			return "", "", nil, err
		} else if set {
			add("i."+key+"=$%d", value)
		}
	}
	if value, set, err := query.NonNegativeInt("min_stock"); err != nil {
		return "", "", nil, err
	} else if set {
		minStock, hasMinStock = value, true
		add("i.stock>=$%d", value)
	}
	if value, set, err := query.NonNegativeInt("max_stock"); err != nil {
		return "", "", nil, err
	} else if set {
		maxStock, hasMaxStock = value, true
		add("i.stock<=$%d", value)
	}
	if hasMinStock && hasMaxStock && minStock > maxStock {
		return "", "", nil, errors.New("min_stock must not exceed max_stock")
	}
	sortColumns := map[string]string{
		"id": "i.id", "sku": "i.sku", "name": "i.name", "stock": "i.stock",
		"price": "i.price", "cost": "i.cost", "created_at": "i.created_at", "updated_at": "i.updated_at",
	}
	column, ok := sortColumns[query.Sort]
	if !ok || (query.Order != "asc" && query.Order != "desc") {
		return "", "", nil, errors.New("invalid item sorting")
	}
	return " WHERE " + strings.Join(clauses, " AND "), " ORDER BY " + column + " " + query.Order + ", i.id " + query.Order, args, nil
}

func (r *ItemRepository) find(ctx context.Context, where, order, paging string, args []any) ([]entity.Items, error) {
	query := fmt.Sprintf(`
		SELECT i.id,i.supplier_id,COALESCE(i.sku,''),i.category_id,c.name,i.brand_id,b.name,i.unit_id,u.name,i.name,i.description,i.stock,i.price,i.cost,i.created_at,i.updated_at
		FROM %s.items i LEFT JOIN %s.categories c ON c.id=i.category_id LEFT JOIN %s.brands b ON b.id=i.brand_id LEFT JOIN %s.units u ON u.id=i.unit_id
		%s%s%s
	`, r.schema, r.schema, r.schema, r.schema, where, order, paging)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.Items, 0)

	for rows.Next() {
		var item entity.Items

		if err := rows.Scan(
			&item.ID,
			&item.SupplierID,
			&item.SKU, &item.CategoryID, &item.CategoryName, &item.BrandID, &item.BrandName, &item.UnitID, &item.UnitName,
			&item.Name,
			&item.Description,
			&item.Stock,
			&item.Price,
			&item.Cost,
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
			supplier_id,
			name,
			description,
			stock,
			price
			,cost,sku,category_id,brand_id,unit_id
		)
		VALUES
		(
			$1,
			$2,
			$3,
			$4,
			$5,$6,$7,$8,$9,$10
		)
	`, r.schema)

	_, err := r.db.Exec(
		ctx,
		query,
		req.SupplierID,
		req.Name,
		req.Description,
		req.Stock,
		req.Price,
		req.Cost, req.SKU, req.CategoryID, req.BrandID, req.UnitID,
	)

	return err
}

func (r *ItemRepository) FindByID(ctx context.Context, id int) (*entity.Items, error) {
	query := fmt.Sprintf(`
		SELECT i.id,i.supplier_id,COALESCE(i.sku,''),i.category_id,c.name,i.brand_id,b.name,i.unit_id,u.name,i.name,i.description,i.stock,i.price,i.cost,i.created_at,i.updated_at
		FROM %s.items i LEFT JOIN %s.categories c ON c.id=i.category_id LEFT JOIN %s.brands b ON b.id=i.brand_id LEFT JOIN %s.units u ON u.id=i.unit_id
		WHERE i.id=$1 AND i.deleted_at IS NULL
	`, r.schema, r.schema, r.schema, r.schema)

	var item entity.Items

	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.SupplierID,
		&item.SKU, &item.CategoryID, &item.CategoryName, &item.BrandID, &item.BrandName, &item.UnitID, &item.UnitName,
		&item.Name,
		&item.Description,
		&item.Stock,
		&item.Price,
		&item.Cost,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *ItemRepository) Update(
	ctx context.Context,
	id int,
	req dto.UpdateItemRequest,
) error {
	query := fmt.Sprintf(`
		UPDATE %s.items
		SET
			supplier_id = $1,
			name = $2,
			description = $3,
			stock = $4,
			price = $5,
			cost = $6, sku = $7, category_id = $8, brand_id = $9, unit_id = $10,
			updated_at = NOW()
		WHERE
			id = $11
			AND deleted_at IS NULL
	`, r.schema)

	commandTag, err := r.db.Exec(
		ctx,
		query,
		req.SupplierID,
		req.Name,
		req.Description,
		req.Stock,
		req.Price,
		req.Cost, req.SKU, req.CategoryID, req.BrandID, req.UnitID,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrItemNotFound
	}

	return nil
}

func (r *ItemRepository) Delete(
	ctx context.Context,
	id int,
) error {

	query := fmt.Sprintf(`
		UPDATE %s.items
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
		return ErrItemNotFound
	}

	return nil
}
