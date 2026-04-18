package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	userSrv "ecommerce/internal/service/user"
)

type UserHandler struct {
	UserSrv userSrv.UserService
}

func NewUserHandler(srv userSrv.UserService) *UserHandler {
	return &UserHandler{UserSrv: srv}
}

// @Summary Get user profile
// @Tags users
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /users [get]

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Route("/profile", func(r chi.Router) {
		r.Get("/", h.ListUser)
		r.Put("/", h.UpdateProfile)
	})

}

func (h *UserHandler) ListUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User Not Authorized")
		return
	}

	user, err := h.UserSrv.GetProfile(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)

}

// @Summary Update user profile
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /users [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User Not Authorized")
		return
	}

	var req userSrv.UpdateProfileRequest

	updReq := userSrv.UpdateProfileRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
		Role:      req.Role,
	}

	updUser, err := h.UserSrv.UpdateProfile(r.Context(), userID, updReq, false)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}
	respondJSON(w, http.StatusOK, updUser)

}
