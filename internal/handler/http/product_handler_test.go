package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	models "ecommerce/internal/domain"
	productService "ecommerce/internal/service/product"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type stubProductService struct {
	listProductsFn   func(ctx context.Context, filter models.ListFilter) (*productService.ProductListResponse, error)
	getProductFn     func(ctx context.Context, id uuid.UUID) (*models.Product, error)
	searchProductsFn func(ctx context.Context, query string, limit, offset int) ([]*models.Product, error)
	createProductFn  func(ctx context.Context, req productService.CreateProductRequest) (*models.Product, error)
	updateProductFn  func(ctx context.Context, id uuid.UUID, req productService.UpdateProductRequest) (*models.Product, error)
	deleteProductFn  func(ctx context.Context, id uuid.UUID) error
	categoriesFn     func(ctx context.Context) ([]*models.Category, error)
}

func (s *stubProductService) CreateProduct(ctx context.Context, req productService.CreateProductRequest) (*models.Product, error) {
	if s.createProductFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.createProductFn(ctx, req)
}

func (s *stubProductService) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if s.getProductFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getProductFn(ctx, id)
}

func (s *stubProductService) ListProducts(ctx context.Context, filter models.ListFilter) (*productService.ProductListResponse, error) {
	if s.listProductsFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.listProductsFn(ctx, filter)
}

func (s *stubProductService) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
	if s.searchProductsFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.searchProductsFn(ctx, query, limit, offset)
}

func (s *stubProductService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	if s.categoriesFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.categoriesFn(ctx)
}

func (s *stubProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req productService.UpdateProductRequest) (*models.Product, error) {
	if s.updateProductFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.updateProductFn(ctx, id, req)
}

func (s *stubProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if s.deleteProductFn == nil {
		return errors.New("unexpected call")
	}
	return s.deleteProductFn(ctx, id)
}

func (s *stubProductService) CheckAvailability(ctx context.Context, id uuid.UUID, quantity int) (bool, error) {
	return false, errors.New("unexpected call")
}

func (s *stubProductService) ReserveStock(ctx context.Context, id uuid.UUID, quantity int) error {
	return errors.New("unexpected call")
}

func (s *stubProductService) ReleaseStock(ctx context.Context, id uuid.UUID, quantity int) error {
	return errors.New("unexpected call")
}

func TestProductHandlerListProducts(t *testing.T) {
	categoryID := uuid.New()
	service := &stubProductService{
		listProductsFn: func(ctx context.Context, filter models.ListFilter) (*productService.ProductListResponse, error) {
			if filter.CategoryID == nil || *filter.CategoryID != categoryID {
				t.Fatalf("unexpected category filter: %+v", filter.CategoryID)
			}
			if filter.MinPrice == nil || *filter.MinPrice != 10.5 {
				t.Fatalf("unexpected min price: %+v", filter.MinPrice)
			}
			if filter.MaxPrice == nil || *filter.MaxPrice != 99.9 {
				t.Fatalf("unexpected max price: %+v", filter.MaxPrice)
			}
			if filter.InStock == nil || !*filter.InStock {
				t.Fatalf("unexpected in_stock: %+v", filter.InStock)
			}
			if filter.Limit != 12 || filter.Offset != 4 || filter.OrderBy != "price_desc" || filter.Search != "book" {
				t.Fatalf("unexpected filter: %+v", filter)
			}

			return &productService.ProductListResponse{
				Products: []*models.Product{{ID: uuid.New(), Name: "Clean Architecture"}},
				Total:    1,
				Limit:    filter.Limit,
				Offset:   filter.Offset,
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products?category_id="+categoryID.String()+"&min_price=10.5&max_price=99.9&in_stock=true&limit=12&offset=4&order_by=price_desc&search=book", nil)
	rec := httptest.NewRecorder()

	NewProductHandler(service).ListProducts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestProductHandlerGetProductInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/products/not-a-uuid", nil)
	req = withURLParam(req, "id", "not-a-uuid")
	rec := httptest.NewRecorder()

	NewProductHandler(&stubProductService{}).GetProduct(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestProductHandlerGetProductNotFound(t *testing.T) {
	productID := uuid.New()
	service := &stubProductService{
		getProductFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			if id != productID {
				t.Fatalf("unexpected product id: %s", id)
			}
			return nil, productService.ErrProductNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products/"+productID.String(), nil)
	req = withURLParam(req, "id", productID.String())
	rec := httptest.NewRecorder()

	NewProductHandler(service).GetProduct(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestProductHandlerSearchProducts(t *testing.T) {
	service := &stubProductService{
		searchProductsFn: func(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
			if query != "lamp" || limit != 5 || offset != 10 {
				t.Fatalf("unexpected search args: %q %d %d", query, limit, offset)
			}

			return []*models.Product{{ID: uuid.New(), Name: "Desk Lamp"}}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products/search?q=lamp&limit=5&offset=10", nil)
	rec := httptest.NewRecorder()

	NewProductHandler(service).SearchProducts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestProductHandlerListCategories(t *testing.T) {
	service := &stubProductService{
		categoriesFn: func(ctx context.Context) ([]*models.Category, error) {
			now := time.Now()
			return []*models.Category{{ID: uuid.New(), Name: "Books", CreatedAt: now, UpdatedAt: now}}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec := httptest.NewRecorder()

	NewProductHandler(service).ListCategories(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response []models.Category
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 || response[0].Name != "Books" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestProductHandlerListProductsInvalidLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/products?limit=abc", nil)
	rec := httptest.NewRecorder()

	NewProductHandler(&stubProductService{}).ListProducts(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}
