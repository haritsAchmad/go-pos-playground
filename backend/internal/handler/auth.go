package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-pos-playground/internal/auth"
	"go-pos-playground/internal/middleware"
	"go-pos-playground/internal/pkg/listquery"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo          *repository.AuthRepository
	tokens        *auth.Manager
	refreshTokens *auth.RefreshManager
}

const refreshCookieName = "pos_refresh_token"

func NewAuthHandler(repo *repository.AuthRepository, tokens *auth.Manager, refreshTokens *auth.RefreshManager) *AuthHandler {
	return &AuthHandler{repo: repo, tokens: tokens, refreshTokens: refreshTokens}
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: refreshCookieName, Value: value, Path: "/",
		MaxAge: maxAge, HttpOnly: true,
		Secure:   r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"),
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}
	u, err := h.repo.FindByEmail(r.Context(), req.Email)
	if errors.Is(err, repository.ErrUserNotFound) || (err == nil && bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil) {
		response.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to authenticate")
		return
	}
	if !u.Active {
		response.Error(w, http.StatusForbidden, "account is inactive")
		return
	}
	token, expiresAt, err := h.tokens.Issue(u.ID, u.Name, u.Email, u.Role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create access token")
		return
	}
	refreshToken, refreshHash, refreshExpiresAt, err := h.refreshTokens.Issue()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	if err := h.repo.CreateSession(r.Context(), u.ID, refreshHash, refreshExpiresAt); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	h.setRefreshCookie(w, r, refreshToken, h.refreshTokens.MaxAge())
	response.Success(w, http.StatusOK, "login successful", map[string]any{"access_token": token, "token_type": "Bearer", "expires_at": expiresAt, "user": u})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "refresh session required")
		return
	}
	nextToken, nextHash, refreshExpiresAt, err := h.refreshTokens.Issue()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to rotate session")
		return
	}
	user, err := h.repo.RotateSession(r.Context(), auth.HashRefreshToken(cookie.Value), nextHash, refreshExpiresAt)
	if errors.Is(err, repository.ErrInvalidSession) {
		h.setRefreshCookie(w, r, "", -1)
		response.Error(w, http.StatusUnauthorized, "refresh session expired or invalid")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to rotate session")
		return
	}
	token, expiresAt, err := h.tokens.Issue(user.ID, user.Name, user.Email, user.Role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to refresh access token")
		return
	}
	h.setRefreshCookie(w, r, nextToken, h.refreshTokens.MaxAge())
	response.Success(w, http.StatusOK, "access token refreshed", map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		if err := h.repo.RevokeSession(r.Context(), auth.HashRefreshToken(cookie.Value)); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to end session")
			return
		}
	}
	h.setRefreshCookie(w, r, "", -1)
	response.Success(w, http.StatusOK, "logout successful", nil)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	c, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	response.Success(w, http.StatusOK, "profile fetched successfully", map[string]any{"id": c.Subject, "name": c.Name, "email": c.Email, "role": c.Role})
}

func (h *AuthHandler) Users(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query, err := listquery.Parse(r.URL.Query(), listquery.Config{
			DefaultSort: "id",
			Sorts: map[string]bool{
				"id": true, "name": true, "email": true, "role": true, "active": true,
			},
			Filters: map[string]bool{"role": true, "active": true},
		})
		if err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if role := query.Values["role"]; role != "" && role != "admin" && role != "cashier" && role != "viewer" {
			response.Error(w, http.StatusBadRequest, "role must be admin, cashier, or viewer")
			return
		}
		if active := query.Values["active"]; active != "" && active != "true" && active != "false" {
			response.Error(w, http.StatusBadRequest, "active must be true or false")
			return
		}
		params, paginated, ok := paginationParams(w, r)
		if !ok {
			return
		}
		if paginated {
			users, err := h.repo.ListUsersPageQuery(r.Context(), params, query)
			if err != nil {
				response.Error(w, 500, "failed to get users")
				return
			}
			response.Success(w, 200, "users fetched successfully", users)
			return
		}
		users, err := h.repo.ListUsersQuery(r.Context(), query)
		if err != nil {
			response.Error(w, 500, "failed to get users")
			return
		}
		response.Success(w, 200, "users fetched successfully", users)
		return
	}
	if r.Method != http.MethodPost {
		response.Error(w, 405, "method not allowed")
		return
	}
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req) != nil || len(strings.TrimSpace(req.Name)) < 2 || !strings.Contains(req.Email, "@") || len(req.Password) < 8 || (req.Role != "admin" && req.Role != "cashier" && req.Role != "viewer") {
		response.Error(w, 400, "name, valid email, password (min. 8 characters), and valid role are required")
		return
	}
	u, err := h.repo.CreateUser(r.Context(), req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		response.Error(w, 400, "user could not be created; ensure email is unique")
		return
	}
	response.Success(w, 201, "user created successfully", u)
}

func (h *AuthHandler) UserDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/users/"), 10, 64)
	if err != nil || id < 1 {
		response.Error(w, 400, "invalid user id")
		return
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())
	selfID, _ := strconv.ParseInt(claims.Subject, 10, 64)
	if r.Method == http.MethodDelete {
		if id == selfID {
			response.Error(w, 400, "you cannot delete your own account")
			return
		}
		if err := h.repo.DeleteUser(r.Context(), id); errors.Is(err, repository.ErrUserNotFound) {
			response.Error(w, 404, "user not found")
			return
		} else if err != nil {
			response.Error(w, 500, "failed to delete user")
			return
		}
		response.Success(w, 200, "user deleted successfully", nil)
		return
	}
	if r.Method != http.MethodPut {
		response.Error(w, 405, "method not allowed")
		return
	}
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Active   bool   `json:"active"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req) != nil || len(strings.TrimSpace(req.Name)) < 2 || !strings.Contains(req.Email, "@") || (req.Password != "" && len(req.Password) < 8) || (req.Role != "admin" && req.Role != "cashier" && req.Role != "viewer") {
		response.Error(w, 400, "name, valid email, optional password (min. 8 characters), and valid role are required")
		return
	}
	current, err := h.repo.FindUserByID(r.Context(), id)
	if errors.Is(err, repository.ErrUserNotFound) {
		response.Error(w, 404, "user not found")
		return
	} else if err != nil {
		response.Error(w, 500, "failed to get user")
		return
	}
	if id == selfID && req.Role != current.Role {
		response.Error(w, 400, "you cannot change your own role")
		return
	}
	if id == selfID && !req.Active {
		response.Error(w, 400, "you cannot deactivate your own account")
		return
	}
	u, err := h.repo.UpdateUser(r.Context(), id, req.Name, req.Email, req.Password, req.Role, req.Active)
	if err != nil {
		response.Error(w, 400, "user could not be updated; ensure email is unique")
		return
	}
	response.Success(w, 200, "user updated successfully", u)
}
