package service

import (
	"context"
	models "ecommerce/internal/domain"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type productRepoMock struct {
	createFn         func(ctx context.Context, product *models.Product) error
	getByIDFn        func(ctx context.Context, id uuid.UUID) (*models.Product, error)
	listFn           func(ctx context.Context, filter models.ListFilter) ([]*models.Product, error)
	searchFn         func(ctx context.Context, query string, limit, offset int) ([]*models.Product, error)
	listCategoriesFn func(ctx context.Context) ([]*models.Category, error)
	updateFn         func(ctx context.Context, product *models.Product) error
	deleteFn         func(ctx context.Context, id uuid.UUID) error
}

func (m *productRepoMock) Create(ctx context.Context, product *models.Product) error {
	if m.createFn != nil {
		return m.createFn(ctx, product)
	}

	return nil
}

func (m *productRepoMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (m *productRepoMock) List(ctx context.Context, filter models.ListFilter) ([]*models.Product, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}

	return nil, nil
}

func (m *productRepoMock) Search(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, limit, offset)
	}

	return nil, nil
}

func (m *productRepoMock) ListCategories(ctx context.Context) ([]*models.Category, error) {
	if m.listCategoriesFn != nil {
		return m.listCategoriesFn(ctx)
	}

	return nil, nil
}

func (m *productRepoMock) Update(ctx context.Context, product *models.Product) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, product)
	}

	return nil
}

func (m *productRepoMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
}

func TestCreateProductValidatesName(t *testing.T) {
	svc := NewService(&productRepoMock{})

	_, err := svc.CreateProduct(context.Background(), CreateProductRequest{
		Name:  "",
		Price: 10,
		Stock: 1,
	})
	if !errors.Is(err, ErrInvalidProductName) {
		t.Fatalf("expected ErrInvalidProductName, got %v", err)
	}
}

func TestGetProductMapsNotFound(t *testing.T) {
	svc := NewService(&productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return nil, errors.New("product not found")
		},
	})

	_, err := svc.GetProduct(context.Background(), uuid.New())
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestListProductsAppliesDefaultPagination(t *testing.T) {
	svc := NewService(&productRepoMock{
		listFn: func(ctx context.Context, filter models.ListFilter) ([]*models.Product, error) {
			if filter.Limit != 20 {
				t.Fatalf("expected default limit 20, got %d", filter.Limit)
			}

			if filter.Offset != 0 {
				t.Fatalf("expected default offset 0, got %d", filter.Offset)
			}

			return []*models.Product{}, nil
		},
	})

	_, err := svc.ListProducts(context.Background(), models.ListFilter{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSearchProductsNormalizesPagination(t *testing.T) {
	svc := NewService(&productRepoMock{
		searchFn: func(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
			if query != "phone" {
				t.Fatalf("expected trimmed query, got %q", query)
			}

			if limit != 20 {
				t.Fatalf("expected default limit 20, got %d", limit)
			}

			if offset != 0 {
				t.Fatalf("expected default offset 0, got %d", offset)
			}

			return []*models.Product{}, nil
		},
	})

	_, err := svc.SearchProducts(context.Background(), "  phone  ", 0, -5)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDeleteProductMapsNotFound(t *testing.T) {
	svc := NewService(&productRepoMock{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("product not found")
		},
	})

	err := svc.DeleteProduct(context.Background(), uuid.New())
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestListProductsNormalizesFilterValues(t *testing.T) {
	minPrice := 100.0
	maxPrice := 10.0
	svc := NewService(&productRepoMock{
		listFn: func(ctx context.Context, filter models.ListFilter) ([]*models.Product, error) {
			if filter.Search != "phone" {
				t.Fatalf("expected trimmed search, got %q", filter.Search)
			}

			if filter.OrderBy != "price asc" {
				t.Fatalf("expected normalized order by, got %q", filter.OrderBy)
			}

			if filter.MinPrice == nil || *filter.MinPrice != 10.0 {
				t.Fatalf("expected swapped min price 10, got %+v", filter.MinPrice)
			}

			if filter.MaxPrice == nil || *filter.MaxPrice != 100.0 {
				t.Fatalf("expected swapped max price 100, got %+v", filter.MaxPrice)
			}

			return []*models.Product{}, nil
		},
	})

	_, err := svc.ListProducts(context.Background(), models.ListFilter{
		Search:   "  phone  ",
		OrderBy:  "  PRICE ASC ",
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUpdateProductAppliesChanges(t *testing.T) {
	price := 1499.99
	stock := 7
	name := "Updated phone"
	productID := uuid.New()

	svc := NewService(&productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return &models.Product{
				ID:    productID,
				Name:  "Old phone",
				Price: 999.99,
				Stock: 3,
			}, nil
		},
		updateFn: func(ctx context.Context, product *models.Product) error {
			if product.Name != name {
				t.Fatalf("expected updated name %q, got %q", name, product.Name)
			}

			if product.Price != price {
				t.Fatalf("expected updated price %v, got %v", price, product.Price)
			}

			if product.Stock != stock {
				t.Fatalf("expected updated stock %d, got %d", stock, product.Stock)
			}

			return nil
		},
	})

	product, err := svc.UpdateProduct(context.Background(), productID, UpdateProductRequest{
		Name:  &name,
		Price: &price,
		Stock: &stock,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if product.Name != name {
		t.Fatalf("expected response name %q, got %q", name, product.Name)
	}
}

func TestUpdateProductMapsNotFound(t *testing.T) {
	svc := NewService(&productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return nil, errors.New("product not found")
		},
	})

	_, err := svc.UpdateProduct(context.Background(), uuid.New(), UpdateProductRequest{})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestUpdateProductValidatesFields(t *testing.T) {
	invalidName := "   "
	svc := NewService(&productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return &models.Product{ID: uuid.New(), Name: "Phone", Price: 100, Stock: 5}, nil
		},
	})

	_, err := svc.UpdateProduct(context.Background(), uuid.New(), UpdateProductRequest{
		Name: &invalidName,
	})
	if !errors.Is(err, ErrInvalidProductName) {
		t.Fatalf("expected ErrInvalidProductName, got %v", err)
	}
}
