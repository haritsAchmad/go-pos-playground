package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-pos-playground/internal/entity"
	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/pagination"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")

type AuthRepository struct {
	db     *pgxpool.Pool
	schema string
}

func NewAuthRepository(db *pgxpool.Pool, schema string) *AuthRepository {
	return &AuthRepository{db: db, schema: pgx.Identifier{schema}.Sanitize()}
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	q := fmt.Sprintf(`SELECT id,name,email,password_hash,role,active FROM %s.users WHERE LOWER(email)=LOWER($1)`, r.schema)
	var u entity.User
	err := r.db.QueryRow(ctx, q, strings.TrimSpace(email)).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (r *AuthRepository) SeedAdmin(ctx context.Context, name, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	var count int
	if err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.users`, r.schema)).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.users(name,email,password_hash,role) VALUES($1,$2,$3,'admin')`, r.schema), name, strings.ToLower(strings.TrimSpace(email)), string(hash))
	return err
}

func (r *AuthRepository) ListUsers(ctx context.Context) ([]entity.User, error) {
	return r.ListUsersQuery(ctx, defaultUserQuery())
}

func (r *AuthRepository) ListUsersPage(ctx context.Context, params pagination.Params) (pagination.Result[entity.User], error) {
	return r.ListUsersPageQuery(ctx, params, defaultUserQuery())
}

func defaultUserQuery() listquery.Params {
	return listquery.Params{Sort: "id", Order: "asc", Values: map[string]string{}}
}

func (r *AuthRepository) ListUsersQuery(ctx context.Context, query listquery.Params) ([]entity.User, error) {
	where, order, args, err := userQueryParts(query)
	if err != nil {
		return nil, err
	}
	return r.listUsers(ctx, where, order, "", args)
}

func (r *AuthRepository) ListUsersPageQuery(ctx context.Context, params pagination.Params, query listquery.Params) (pagination.Result[entity.User], error) {
	where, order, args, err := userQueryParts(query)
	if err != nil {
		return pagination.Result[entity.User]{}, err
	}
	var total int64
	if err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.users u%s`, r.schema, where), args...).Scan(&total); err != nil {
		return pagination.Result[entity.User]{}, err
	}
	paging := fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.PerPage, params.Offset())
	users, err := r.listUsers(ctx, where, order, paging, args)
	if err != nil {
		return pagination.Result[entity.User]{}, err
	}
	return pagination.NewResult(users, params, total), nil
}

func userQueryParts(query listquery.Params) (string, string, []any, error) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if query.Search != "" {
		args = append(args, query.Search)
		clauses = append(clauses, "(u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')")
	}
	if role := query.Values["role"]; role != "" {
		if role != "admin" && role != "cashier" && role != "viewer" {
			return "", "", nil, errors.New("invalid user role filter")
		}
		args = append(args, role)
		clauses = append(clauses, fmt.Sprintf("u.role=$%d", len(args)))
	}
	if active := query.Values["active"]; active != "" {
		if active != "true" && active != "false" {
			return "", "", nil, errors.New("invalid active filter")
		}
		args = append(args, active == "true")
		clauses = append(clauses, fmt.Sprintf("u.active=$%d", len(args)))
	}
	sortColumns := map[string]string{
		"id": "u.id", "name": "u.name", "email": "u.email", "role": "u.role", "active": "u.active",
	}
	column, ok := sortColumns[query.Sort]
	if !ok || (query.Order != "asc" && query.Order != "desc") {
		return "", "", nil, errors.New("invalid user sorting")
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	return where, " ORDER BY " + column + " " + query.Order + ", u.id " + query.Order, args, nil
}

func (r *AuthRepository) listUsers(ctx context.Context, where, order, paging string, args []any) ([]entity.User, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT u.id,u.name,u.email,u.role,u.active FROM %s.users u%s%s%s`, r.schema, where, order, paging), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []entity.User{}
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Active); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *AuthRepository) CreateUser(ctx context.Context, name, email, password, role string) (*entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var u entity.User
	err = r.db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.users(name,email,password_hash,role) VALUES($1,$2,$3,$4) RETURNING id,name,email,role,active`, r.schema), strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(email)), string(hash), role).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Active)
	return &u, err
}

func (r *AuthRepository) FindUserByID(ctx context.Context, id int64) (*entity.User, error) {
	var u entity.User
	err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT id,name,email,role,active FROM %s.users WHERE id=$1`, r.schema), id).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (r *AuthRepository) UpdateUser(ctx context.Context, id int64, name, email, password, role string, active bool) (*entity.User, error) {
	var hash any = nil
	if password != "" {
		value, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hash = string(value)
	}
	var u entity.User
	err := r.db.QueryRow(ctx, fmt.Sprintf(`UPDATE %s.users SET name=$2,email=LOWER($3),role=$4,active=$5,password_hash=COALESCE($6,password_hash),updated_at=NOW() WHERE id=$1 RETURNING id,name,email,role,active`, r.schema), id, strings.TrimSpace(name), strings.TrimSpace(email), role, active, hash).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (r *AuthRepository) DeleteUser(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.users WHERE id=$1`, r.schema), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
