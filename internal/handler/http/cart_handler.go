package http

import (
	"encoding/json"
	"net/http"

	cartService "ecommerce/internal/service/cart"

	"github.com/go-chi/chi/v5"
)

type CartHandler struct {
	cartService cartService.CartService
}

type UpdateCartItemQuantityRequest struct {
	Quantity int `json:"quantity"`
}

func NewCartHandler(service cartService.CartService) *CartHandler {
	return &CartHandler{cartService: service}
}

func (h *CartHandler) RegisterRoutes(r chi.Router) {
	r.Get("/cart", h.GetCart)
	r.Delete("/cart", h.ClearCart)
	r.Post("/cart/items", h.AddItem)
	r.Put("/cart/items/{id}", h.UpdateItemQuantity)
	r.Delete("/cart/items/{id}", h.RemoveItem)
}

// GetCart godoc
// @Summary Get cart
// @Description View the authenticated user's cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} service.CartResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	cart, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		handleCartError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, cart)
}

// AddItem godoc
// @Summary Add item to cart
// @Description Add a product to the authenticated user's cart
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.AddItemRequest true "Cart item payload"
// @Success 201 {object} models.CartItem
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cart/items [post]
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req cartService.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item, err := h.cartService.AddItem(r.Context(), userID, req)
	if err != nil {
		handleCartError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// UpdateItemQuantity godoc
// @Summary Update cart item quantity
// @Description Update quantity for a cart item owned by the authenticated user
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cart item UUID"
// @Param request body UpdateCartItemQuantityRequest true "New quantity"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cart/items/{id} [put]
func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	itemID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.cartService.UpdateQuantity(r.Context(), userID, itemID, req.Quantity); err != nil {
		handleCartError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Cart item updated successfully",
	})
}

// RemoveItem godoc
// @Summary Remove cart item
// @Description Remove a cart item from the authenticated user's cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cart item UUID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	itemID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.cartService.RemoveItem(r.Context(), userID, itemID); err != nil {
		handleCartError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Cart item removed successfully",
	})
}

// ClearCart godoc
// @Summary Clear cart
// @Description Remove all items from the authenticated user's cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cart [delete]
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.cartService.ClearCart(r.Context(), userID); err != nil {
		handleCartError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Cart cleared successfully",
	})
}

func handleCartError(w http.ResponseWriter, err error) {
	switch err {
	case cartService.ErrInvalidCartItem,
		cartService.ErrInvalidQuantity:
		respondError(w, http.StatusBadRequest, err.Error())
	case cartService.ErrCartItemNotFound:
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "Internal server error")
	}
}
