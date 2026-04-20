package http

import (
	userService "ecommerce/internal/service/user"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	productService productService.ProductService
	orderService   orderService.OrderService
	userService    userService.UserService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	productSvc productService.ProductService,
	orderSvc orderService.OrderService,
	userSvc userService.UserService,
) *AdminHandler {
	return &AdminHandler{
		productService: productSvc,
		orderService:   orderSvc,
		userService:    userSvc,
	}
}

// RegisterRoutes registers all admin routes
func (h *AdminHandler) RegisterRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		// All admin routes require admin role

		// Product management
		r.Post("/products", h.CreateProduct)
		r.Put("/products/{id}", h.UpdateProduct)
		r.Delete("/products/{id}", h.DeleteProduct)

		// Order management
		r.Get("/orders", h.ListAllOrders)
		r.Put("/orders/{id}/status", h.UpdateOrderStatus)

		// User management
		r.Get("/users", h.ListUsers)
		r.Delete("/users/{id}", h.DeleteUser)

		// Statistics
		r.Get("/statistics", h.GetStatistics)
	})
}

// CreateProduct godoc
// @Summary Create product (Admin)
// @Description Create a new product (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param product body service.CreateProductRequest true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/products [post]
// @Security BearerAuth

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req productService.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call service
	product, err := h.productService.CreateProduct(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, product)
}
