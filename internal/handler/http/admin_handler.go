package http

import (
	"encoding/json"
	"net/http"
	"strings"

	models "ecommerce/internal/domain"
	orderService "ecommerce/internal/service/order"
	productService "ecommerce/internal/service/product"
	userService "ecommerce/internal/service/user"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	productService productService.ProductService
	orderService   orderService.OrderService
	userService    userService.UserService
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

type StatisticsResponse struct {
	Products       int            `json:"products"`
	Orders         int            `json:"orders"`
	Users          int            `json:"users"`
	Revenue        float64        `json:"revenue"`
	OrdersByStatus map[string]int `json:"orders_by_status"`
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(
	userSvc userService.UserService,
	orderSvc orderService.OrderService,
	productSvc productService.ProductService,
) *AdminHandler {
	return &AdminHandler{
		productService: productSvc,
		orderService:   orderSvc,
		userService:    userSvc,
	}
}

// RegisterRoutes registers all admin routes.
func (h *AdminHandler) RegisterRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Post("/products", h.CreateProduct)
		r.Put("/products/{id}", h.UpdateProduct)
		r.Delete("/products/{id}", h.DeleteProduct)

		r.Get("/orders", h.ListAllOrders)
		r.Put("/orders/{id}/status", h.UpdateOrderStatus)

		r.Get("/users", h.ListUsers)
		r.Get("/statistics", h.GetStatistics)
	})
}

// CreateProduct godoc
// @Summary Create product
// @Description Create a new product as admin
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateProductRequest true "Product payload"
// @Success 201 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/products [post]
func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req productService.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	product, err := h.productService.CreateProduct(r.Context(), req)
	if err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update a product as admin
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Param request body service.UpdateProductRequest true "Updated product payload"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/products/{id} [put]
func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req productService.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	product, err := h.productService.UpdateProduct(r.Context(), productID, req)
	if err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product as admin
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/products/{id} [delete]
func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.DeleteProduct(r.Context(), productID); err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// ListAllOrders godoc
// @Summary List all orders
// @Description Get paginated list of all orders as admin
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Success 200 {object} OrderListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/orders [get]
func (h *AdminHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
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

	orders, err := h.orderService.ListAllOrders(r.Context(), limit, offset)
	if err != nil {
		handleOrderError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"orders": orders,
		"total":  len(orders),
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update an order status as admin
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Param request body UpdateOrderStatusRequest true "Order status payload"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/orders/{id}/status [put]
func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.orderService.UpdateStatus(r.Context(), orderID, req.Status); err != nil {
		handleOrderError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Order status updated successfully",
	})
}

// ListUsers godoc
// @Summary List users
// @Description Get paginated list of users as admin
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by name"
// @Param role query string false "Filter by role"
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Success 200 {object} service.UserListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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

	filter := userService.UserFilter{
		Search: strings.TrimSpace(r.URL.Query().Get("search")),
		Limit:  limit,
		Offset: offset,
	}

	if role := strings.TrimSpace(r.URL.Query().Get("role")); role != "" {
		filter.Role = &role
	}

	users, err := h.userService.List(r.Context(), filter)
	if err != nil {
		handleAdminUserError(w, err)
		return
	}

	if len(users) == 0 {
		respondJSON(w, http.StatusOK, userService.UserListResponse{
			Users:  []*models.User{},
			Total:  0,
			Limit:  limit,
			Offset: offset,
		})
		return
	}

	respondJSON(w, http.StatusOK, users[0])
}

// GetStatistics godoc
// @Summary Get statistics
// @Description Get dashboard statistics as admin
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} StatisticsResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/statistics [get]
func (h *AdminHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	productResp, err := h.productService.ListProducts(r.Context(), models.ListFilter{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		handleProductError(w, err)
		return
	}

	orders, err := h.orderService.ListAllOrders(r.Context(), 100, 0)
	if err != nil {
		handleOrderError(w, err)
		return
	}

	users, err := h.userService.List(r.Context(), userService.UserFilter{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		handleAdminUserError(w, err)
		return
	}

	productsCount := 0
	if productResp != nil {
		productsCount = len(productResp.Products)
	}

	usersCount := 0
	if len(users) > 0 {
		usersCount = users[0].Total
	}

	revenue := 0.0
	ordersByStatus := map[string]int{}
	for _, order := range orders {
		if order == nil {
			continue
		}

		revenue += order.TotalAmount
		ordersByStatus[order.Status]++
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"products":         productsCount,
		"orders":           len(orders),
		"users":            usersCount,
		"revenue":          revenue,
		"orders_by_status": ordersByStatus,
	})
}

func handleAdminUserError(w http.ResponseWriter, err error) {
	switch err {
	case userService.ErrUserNotFound:
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "Internal server error")
	}
	return
}
