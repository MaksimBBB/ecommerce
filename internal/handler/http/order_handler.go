package http

import (
	"encoding/json"
	"net/http"

	models "ecommerce/internal/domain"
	orderService "ecommerce/internal/service/order"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderService orderService.OrderService
}

type OrderListResponse struct {
	Orders []*models.Order `json:"orders"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	UserID string          `json:"user_id,omitempty"`
}

func NewOrderHandler(service orderService.OrderService) *OrderHandler {
	return &OrderHandler{orderService: service}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Post("/orders", h.CreateOrder)
	r.Get("/orders", h.ListOrders)
	r.Get("/orders/{id}", h.GetOrder)
	r.Put("/orders/{id}/cancel", h.CancelOrder)
}

// CreateOrder godoc
// @Summary Create order
// @Description Create a new order from the authenticated user's cart
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateOrderRequest true "Order payload"
// @Success 201 {object} service.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req orderService.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	order, err := h.orderService.CreateOrder(r.Context(), userID, req)
	if err != nil {
		handleOrderError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

// ListOrders godoc
// @Summary List my orders
// @Description Get paginated orders for the authenticated user
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Success 200 {object} OrderListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limit, err := parseOptionalIntQuery(r, "limit", 20)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	offset, err := parseOptionalIntQuery(r, "offset", 0)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	orders, err := h.orderService.ListOrders(r.Context(), userID, limit, offset)
	if err != nil {
		handleOrderError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"orders":  orders,
		"total":   len(orders),
		"limit":   limit,
		"offset":  offset,
		"user_id": userID,
	})
}

// GetOrder godoc
// @Summary Get order details
// @Description Get a single order owned by the authenticated user
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Success 200 {object} service.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	orderID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.orderService.GetOrder(r.Context(), userID, orderID)
	if err != nil {
		handleOrderError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, order)
}

// CancelOrder godoc
// @Summary Cancel order
// @Description Cancel an order owned by the authenticated user
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id}/cancel [put]
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	orderID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.orderService.CancelOrder(r.Context(), userID, orderID); err != nil {
		handleOrderError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Order cancelled successfully",
	})
}

func handleOrderError(w http.ResponseWriter, err error) {
	switch err {
	case orderService.ErrInvalidOrder,
		orderService.ErrInvalidOrderItem,
		orderService.ErrInvalidOrderStatus,
		orderService.ErrEmptyOrderItems,
		orderService.ErrInvalidShippingAddress,
		orderService.ErrInvalidPaymentMethod,
		orderService.ErrInsufficientStock:
		respondError(w, http.StatusBadRequest, err.Error())
	case orderService.ErrOrderNotFound:
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "Internal server error")
	}
}
