package service

import (
	"context"
	models "ecommerce/internal/domain"
	repository "ecommerce/internal/repository/postgres"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type cartRepoMock struct {
	addItemFn        func(ctx context.Context, item *models.CartItem) error
	getByUserIDFn    func(ctx context.Context, userID uuid.UUID) ([]*models.CartItemWithProduct, error)
	updateQuantityFn func(ctx context.Context, id uuid.UUID, quantity int) error
	removeItemFn     func(ctx context.Context, id uuid.UUID) error
	clearFn          func(ctx context.Context, userID uuid.UUID) error
	getItemFn        func(ctx context.Context, userID, productID uuid.UUID) (*models.CartItem, error)
}

func (m *cartRepoMock) AddItem(ctx context.Context, item *models.CartItem) error {
	if m.addItemFn != nil {
		return m.addItemFn(ctx, item)
	}

	return nil
}

func (m *cartRepoMock) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.CartItemWithProduct, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, userID)
	}

	return nil, nil
}

func (m *cartRepoMock) UpdateQuantity(ctx context.Context, id uuid.UUID, quantity int) error {
	if m.updateQuantityFn != nil {
		return m.updateQuantityFn(ctx, id, quantity)
	}

	return nil
}

func (m *cartRepoMock) RemoveItem(ctx context.Context, id uuid.UUID) error {
	if m.removeItemFn != nil {
		return m.removeItemFn(ctx, id)
	}

	return nil
}

func (m *cartRepoMock) Clear(ctx context.Context, userID uuid.UUID) error {
	if m.clearFn != nil {
		return m.clearFn(ctx, userID)
	}

	return nil
}

func (m *cartRepoMock) GetItem(ctx context.Context, userID, productID uuid.UUID) (*models.CartItem, error) {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, userID, productID)
	}

	return nil, nil
}

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

var (
	_ repository.CartRepository    = (*cartRepoMock)(nil)
	_ repository.ProductRepository = (*productRepoMock)(nil)
)

func TestAddItemRejectsInsufficientStock(t *testing.T) {
	productID := uuid.New()
	svc := NewService(&cartRepoMock{
		getItemFn: func(ctx context.Context, userID, requestedProductID uuid.UUID) (*models.CartItem, error) {
			return nil, errors.New("cart item not found")
		},
	}, &productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return &models.Product{ID: productID, Stock: 1}, nil
		},
	})

	_, err := svc.AddItem(context.Background(), uuid.New(), AddItemRequest{
		ProductID: productID,
		Quantity:  2,
	})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestAddItemUpdatesExistingQuantityWithinStock(t *testing.T) {
	productID := uuid.New()
	itemID := uuid.New()
	updated := false

	svc := NewService(&cartRepoMock{
		getItemFn: func(ctx context.Context, userID, requestedProductID uuid.UUID) (*models.CartItem, error) {
			return &models.CartItem{
				ID:        itemID,
				UserID:    userID,
				ProductID: requestedProductID,
				Quantity:  2,
			}, nil
		},
		updateQuantityFn: func(ctx context.Context, id uuid.UUID, quantity int) error {
			updated = true
			if id != itemID {
				t.Fatalf("expected item ID %s, got %s", itemID, id)
			}

			if quantity != 4 {
				t.Fatalf("expected quantity 4, got %d", quantity)
			}

			return nil
		},
	}, &productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return &models.Product{ID: productID, Stock: 5}, nil
		},
	})

	item, err := svc.AddItem(context.Background(), uuid.New(), AddItemRequest{
		ProductID: productID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !updated {
		t.Fatalf("expected existing cart item to be updated")
	}

	if item.Quantity != 4 {
		t.Fatalf("expected response quantity 4, got %d", item.Quantity)
	}
}

func TestUpdateQuantityRejectsInsufficientStock(t *testing.T) {
	itemID := uuid.New()
	productID := uuid.New()
	svc := NewService(&cartRepoMock{
		getByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*models.CartItemWithProduct, error) {
			return []*models.CartItemWithProduct{
				{
					CartItem: models.CartItem{
						ID:        itemID,
						UserID:    userID,
						ProductID: productID,
						Quantity:  1,
					},
					ProductStock: 2,
				},
			}, nil
		},
	}, &productRepoMock{})

	err := svc.UpdateQuantity(context.Background(), uuid.New(), itemID, 3)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}
