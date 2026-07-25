package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go-pos-playground/internal/auth"
	"go-pos-playground/internal/repository"
	"go-pos-playground/internal/pkg/response"
)

type contextKey string
const claimsKey contextKey = "auth_claims"

func Authenticate(tokens *auth.Manager, users *repository.AuthRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") { response.Error(w,http.StatusUnauthorized,"authentication required"); return }
		claims, err := tokens.Parse(parts[1])
		if err != nil { response.Error(w,http.StatusUnauthorized,"invalid or expired access token"); return }
		userID, err := strconv.ParseInt(claims.Subject,10,64)
		if err != nil { response.Error(w,http.StatusUnauthorized,"invalid access token"); return }
		user, err := users.FindUserByID(r.Context(),userID)
		if err != nil || !user.Active { response.Error(w,http.StatusUnauthorized,"account is unavailable or inactive"); return }
		claims.Name,claims.Email,claims.Role=user.Name,user.Email,user.Role
		next(w, r.WithContext(context.WithValue(r.Context(),claimsKey,claims)))
	}
}

func Authorize(next http.HandlerFunc, roles ...string) http.HandlerFunc {
	allowed := map[string]bool{}; for _, role := range roles { allowed[role] = true }
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok { response.Error(w,http.StatusUnauthorized,"authentication required"); return }
		if !allowed[claims.Role] { response.Error(w,http.StatusForbidden,"you do not have permission to perform this action"); return }
		next(w,r)
	}
}

func ClaimsFromContext(ctx context.Context) (auth.Claims, bool) { c, ok := ctx.Value(claimsKey).(auth.Claims); return c,ok }
