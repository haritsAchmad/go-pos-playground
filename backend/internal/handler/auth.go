package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-pos-playground/internal/auth"
	"go-pos-playground/internal/middleware"
	"go-pos-playground/internal/pkg/response"
	"go-pos-playground/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct { repo *repository.AuthRepository; tokens *auth.Manager }
func NewAuthHandler(repo *repository.AuthRepository, tokens *auth.Manager) *AuthHandler { return &AuthHandler{repo,tokens} }

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct { Email string `json:"email"`; Password string `json:"password"` }
	if err := json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&req); err != nil || strings.TrimSpace(req.Email)=="" || req.Password=="" { response.Error(w,http.StatusBadRequest,"email and password are required"); return }
	u, err := h.repo.FindByEmail(r.Context(),req.Email)
	if errors.Is(err,repository.ErrUserNotFound) || (err == nil && bcrypt.CompareHashAndPassword([]byte(u.PasswordHash),[]byte(req.Password)) != nil) { response.Error(w,http.StatusUnauthorized,"invalid email or password"); return }
	if err != nil { response.Error(w,http.StatusInternalServerError,"failed to authenticate"); return }
	if !u.Active { response.Error(w,http.StatusForbidden,"account is inactive"); return }
	token, expiresAt, err := h.tokens.Issue(u.ID,u.Name,u.Email,u.Role); if err != nil { response.Error(w,http.StatusInternalServerError,"failed to create access token"); return }
	response.Success(w,http.StatusOK,"login successful",map[string]any{"access_token":token,"token_type":"Bearer","expires_at":expiresAt,"user":u})
}

// Refresh extends an active session by issuing a fresh JWT for the authenticated user.
// The authentication middleware has already rejected expired, deleted, or inactive accounts.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid access token")
		return
	}
	token, expiresAt, err := h.tokens.Issue(userID, claims.Name, claims.Email, claims.Role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to refresh access token")
		return
	}
	response.Success(w, http.StatusOK, "access token refreshed", map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	c, ok := middleware.ClaimsFromContext(r.Context()); if !ok { response.Error(w,http.StatusUnauthorized,"authentication required"); return }
	response.Success(w,http.StatusOK,"profile fetched successfully",map[string]any{"id":c.Subject,"name":c.Name,"email":c.Email,"role":c.Role})
}

func (h *AuthHandler) Users(w http.ResponseWriter,r *http.Request){
	if r.Method==http.MethodGet{users,err:=h.repo.ListUsers(r.Context());if err!=nil{response.Error(w,500,"failed to get users");return};response.Success(w,200,"users fetched successfully",users);return}
	if r.Method!=http.MethodPost{response.Error(w,405,"method not allowed");return}
	var req struct{Name string `json:"name"`;Email string `json:"email"`;Password string `json:"password"`;Role string `json:"role"`}
	if json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&req)!=nil||len(strings.TrimSpace(req.Name))<2||!strings.Contains(req.Email,"@")||len(req.Password)<8||(req.Role!="admin"&&req.Role!="cashier"&&req.Role!="viewer"){response.Error(w,400,"name, valid email, password (min. 8 characters), and valid role are required");return}
	u,err:=h.repo.CreateUser(r.Context(),req.Name,req.Email,req.Password,req.Role);if err!=nil{response.Error(w,400,"user could not be created; ensure email is unique");return};response.Success(w,201,"user created successfully",u)
}

func (h *AuthHandler) UserDetail(w http.ResponseWriter,r *http.Request){
	id,err:=strconv.ParseInt(strings.TrimPrefix(r.URL.Path,"/users/"),10,64);if err!=nil||id<1{response.Error(w,400,"invalid user id");return}
	claims,_:=middleware.ClaimsFromContext(r.Context());selfID,_:=strconv.ParseInt(claims.Subject,10,64)
	if r.Method==http.MethodDelete{if id==selfID{response.Error(w,400,"you cannot delete your own account");return};if err:=h.repo.DeleteUser(r.Context(),id);errors.Is(err,repository.ErrUserNotFound){response.Error(w,404,"user not found");return}else if err!=nil{response.Error(w,500,"failed to delete user");return};response.Success(w,200,"user deleted successfully",nil);return}
	if r.Method!=http.MethodPut{response.Error(w,405,"method not allowed");return}
	var req struct{Name string `json:"name"`;Email string `json:"email"`;Password string `json:"password"`;Role string `json:"role"`;Active bool `json:"active"`}
	if json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&req)!=nil||len(strings.TrimSpace(req.Name))<2||!strings.Contains(req.Email,"@")||(req.Password!=""&&len(req.Password)<8)||(req.Role!="admin"&&req.Role!="cashier"&&req.Role!="viewer"){response.Error(w,400,"name, valid email, optional password (min. 8 characters), and valid role are required");return}
	current,err:=h.repo.FindUserByID(r.Context(),id);if errors.Is(err,repository.ErrUserNotFound){response.Error(w,404,"user not found");return}else if err!=nil{response.Error(w,500,"failed to get user");return}
	if id==selfID&&req.Role!=current.Role{response.Error(w,400,"you cannot change your own role");return}
	if id==selfID&&!req.Active{response.Error(w,400,"you cannot deactivate your own account");return}
	u,err:=h.repo.UpdateUser(r.Context(),id,req.Name,req.Email,req.Password,req.Role,req.Active);if err!=nil{response.Error(w,400,"user could not be updated; ensure email is unique");return};response.Success(w,200,"user updated successfully",u)
}
