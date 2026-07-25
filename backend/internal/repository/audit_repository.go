package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/pagination"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	db     *pgxpool.Pool
	schema string
}

func NewAuditRepository(db *pgxpool.Pool, schema string) *AuditRepository {
	return &AuditRepository{db: db, schema: pgx.Identifier{schema}.Sanitize()}
}

func (r *AuditRepository) Record(ctx context.Context, entry entity.AuditEntry) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s.audit_logs
			(user_id,user_name,user_email,action,entity_type,entity_id,method,path,status_code,ip_address,user_agent)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, r.schema), entry.UserID, entry.UserName, entry.UserEmail, entry.Action, entry.EntityType, entry.EntityID,
		entry.Method, entry.Path, entry.StatusCode, entry.IPAddress, entry.UserAgent)
	return err
}

func (r *AuditRepository) List(ctx context.Context, query listquery.Params) ([]entity.AuditLog, error) {
	where, order, args, err := auditQueryParts(query)
	if err != nil {
		return nil, err
	}
	return r.list(ctx, where, order, "", args)
}

func (r *AuditRepository) ListPage(ctx context.Context, params pagination.Params, query listquery.Params) (pagination.Result[entity.AuditLog], error) {
	where, order, args, err := auditQueryParts(query)
	if err != nil {
		return pagination.Result[entity.AuditLog]{}, err
	}
	var total int64
	if err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.audit_logs a%s
	`, r.schema, where), args...).Scan(&total); err != nil {
		return pagination.Result[entity.AuditLog]{}, err
	}
	paging := fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.PerPage, params.Offset())
	items, err := r.list(ctx, where, order, paging, args)
	if err != nil {
		return pagination.Result[entity.AuditLog]{}, err
	}
	return pagination.NewResult(items, params, total), nil
}

func auditQueryParts(query listquery.Params) (string, string, []any, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if query.Search != "" {
		args = append(args, query.Search)
		position := len(args)
		clauses = append(clauses, fmt.Sprintf(
			"(a.user_name ILIKE '%%' || $%d || '%%' OR a.user_email ILIKE '%%' || $%d || '%%' OR a.path ILIKE '%%' || $%d || '%%' OR a.entity_type ILIKE '%%' || $%d || '%%')",
			position, position, position, position,
		))
	}
	if action := query.Values["action"]; action != "" {
		add("a.action=$%d", action)
	}
	if entityType := query.Values["entity_type"]; entityType != "" {
		add("a.entity_type=$%d", entityType)
	}
	if userID, set, err := query.PositiveInt("user_id"); err != nil {
		return "", "", nil, err
	} else if set {
		add("a.user_id=$%d", userID)
	}
	sortColumns := map[string]string{
		"id": "a.id", "created_at": "a.created_at", "user_name": "a.user_name",
		"action": "a.action", "entity_type": "a.entity_type", "status_code": "a.status_code",
	}
	column, ok := sortColumns[query.Sort]
	if !ok || (query.Order != "asc" && query.Order != "desc") {
		return "", "", nil, errors.New("invalid audit sorting")
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	return where, " ORDER BY " + column + " " + query.Order + ", a.id " + query.Order, args, nil
}

func (r *AuditRepository) list(ctx context.Context, where, order, paging string, args []any) ([]entity.AuditLog, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT a.id,a.user_id,a.user_name,a.user_email,a.action,a.entity_type,a.entity_id,
			a.method,a.path,a.status_code,a.ip_address,a.user_agent,a.created_at
		FROM %s.audit_logs a
		%s%s%s
	`, r.schema, where, order, paging), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.AuditLog, 0)
	for rows.Next() {
		var value entity.AuditLog
		if err := rows.Scan(
			&value.ID, &value.UserID, &value.UserName, &value.UserEmail,
			&value.Action, &value.EntityType, &value.EntityID, &value.Method,
			&value.Path, &value.StatusCode, &value.IPAddress, &value.UserAgent, &value.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
