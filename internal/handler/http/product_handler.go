package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	models "ecommerce/internal/domain"
	productService "ecommerce/internal/service/product"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type categoryProductService interface {
	productService.ProductService
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

type ProductHandler struct {
	productService productService.ProductService
	categorySrv    categoryProductService
}

type ProductSearchResponse struct {
	Products []*models.Product `json:"products"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Query    string            `json:"query"`
}

func NewProductHandler(service productService.ProductService) *ProductHandler {
	var categorySrv categoryProductService
	if srv, ok := service.(categoryProductService); ok {
		categorySrv = srv
	}

	return &ProductHandler{
		productService: service,
		categorySrv:    categorySrv,
	}
}

func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Get("/products", h.ListProducts)
	r.Get("/products/search", h.SearchProducts)
	r.Get("/products/{id}", h.GetProduct)
	r.Get("/categories", h.ListCategories)
}

// ListProducts godoc
// @Summary List products
// @Description Get a paginated list of products with optional filters
// @Tags products
// @Produce json
// @Param category_id query string false "Category UUID"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param search query string false "Search text"
// @Param in_stock query boolean false "Only show in-stock products"
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Param order_by query string false "Sort order" Enums(price_asc,price_desc,name_asc,name_desc,created_at_asc,created_at_desc)
// @Success 200 {object} service.ProductListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter, err := parseProductListFilter(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.productService.ListProducts(r.Context(), filter)
	if err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// GetProduct godoc
// @Summary Get product details
// @Description Get a single product by ID
// @Tags products
// @Produce json
// @Param id path string true "Product UUID"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.productService.GetProduct(r.Context(), productID)
	if err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, product)
}

// SearchProducts godoc
// @Summary Search products
// @Description Search products by query text
// @Tags products
// @Produce json
// @Param q query string false "Search query"
// @Param query query string false "Search query alias"
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Success 200 {object} ProductSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/search [get]
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("query"))
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

	products, err := h.productService.SearchProducts(r.Context(), query, limit, offset)
	if err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"products": products,
		"total":    len(products),
		"limit":    limit,
		"offset":   offset,
		"query":    query,
	})
}

// ListCategories godoc
// @Summary List categories
// @Description Get all product categories
// @Tags products
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/categories [get]
func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if h.categorySrv == nil {
		respondError(w, http.StatusInternalServerError, "Categories are not supported")
		return
	}

	categories, err := h.categorySrv.ListCategories(r.Context())
	if err != nil {
		handleProductError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

func parseProductListFilter(r *http.Request) (models.ListFilter, error) {
	query := r.URL.Query()
	filter := models.ListFilter{
		Search:  strings.TrimSpace(query.Get("search")),
		OrderBy: strings.TrimSpace(query.Get("order_by")),
	}

	limit, err := parseOptionalIntQuery(r, "limit", 20)
	if err != nil {
		return models.ListFilter{}, err
	}
	filter.Limit = limit

	offset, err := parseOptionalIntQuery(r, "offset", 0)
	if err != nil {
		return models.ListFilter{}, err
	}
	filter.Offset = offset

	if categoryID := strings.TrimSpace(query.Get("category_id")); categoryID != "" {
		parsed, err := uuid.Parse(categoryID)
		if err != nil {
			return models.ListFilter{}, errors.New("Invalid category_id")
		}
		filter.CategoryID = &parsed
	}

	if minPrice := strings.TrimSpace(query.Get("min_price")); minPrice != "" {
		parsed, err := strconv.ParseFloat(minPrice, 64)
		if err != nil {
			return models.ListFilter{}, errors.New("Invalid min_price")
		}
		filter.MinPrice = &parsed
	}

	if maxPrice := strings.TrimSpace(query.Get("max_price")); maxPrice != "" {
		parsed, err := strconv.ParseFloat(maxPrice, 64)
		if err != nil {
			return models.ListFilter{}, errors.New("Invalid max_price")
		}
		filter.MaxPrice = &parsed
	}

	if inStock := strings.TrimSpace(query.Get("in_stock")); inStock != "" {
		parsed, err := strconv.ParseBool(inStock)
		if err != nil {
			return models.ListFilter{}, errors.New("Invalid in_stock")
		}
		filter.InStock = &parsed
	}

	return filter, nil
}

func parseOptionalIntQuery(r *http.Request, key string, fallback int) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("Invalid " + key)
	}

	return parsed, nil
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	value := chi.URLParam(r, name)
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, errors.New("Invalid " + name)
	}

	return parsed, nil
}

func handleProductError(w http.ResponseWriter, err error) {
	switch err {
	case productService.ErrInvalidProduct,
		productService.ErrInvalidProductName,
		productService.ErrInvalidPrice,
		productService.ErrInvalidStock:
		respondError(w, http.StatusBadRequest, err.Error())
	case productService.ErrProductNotFound:
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "Internal server error")
	}
}
