package http

import (
	"encoding/json"
	"log"
	"net/http"

	authService "ecommerce/internal/service/auth"

	"github.com/go-chi/chi/v5"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	authService authService.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(service authService.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: service,
	}
}

// RegisterRoutes registers all auth routes.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.RefreshToken)
		r.Post("/logout", h.Logout)
	})
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Registration data"
// @Success 201 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authService.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.authService.Register(r.Context(), req)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login credentials"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authService.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user (client should delete tokens)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
	})
}

// RefreshTokenRequest represents refresh token request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// SuccessResponse represents a success response.
type SuccessResponse struct {
	Message string `json:"message"`
}

// handleAuthError maps auth service errors to HTTP status codes.
func handleAuthError(w http.ResponseWriter, err error) {
	switch err {
	case authService.ErrEmailRequired,
		authService.ErrPasswordRequired,
		authService.ErrWeakPassword,
		authService.ErrInvalidFirstName,
		authService.ErrInvalidLastName:
		respondError(w, http.StatusBadRequest, err.Error())
	case authService.ErrUserAlreadyExists:
		respondError(w, http.StatusConflict, err.Error())
	case authService.ErrInvalidCredentials:
		respondError(w, http.StatusUnauthorized, err.Error())
	case authService.ErrInvalidToken,
		authService.ErrTokenExpired:
		respondError(w, http.StatusUnauthorized, err.Error())
	default:
		log.Printf("unexpected auth error: %v", err)
		respondError(w, http.StatusInternalServerError, "Internal server error")
	}
}
