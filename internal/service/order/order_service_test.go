package service

import (
	"context"
	models "ecommerce/internal/domain"
	repository "ecommerce/internal/repository/postgres"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type orderRepoMock struct {
	createFn        func(ctx context.Context, order *models.Order, items []*models.OrderItem) error
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*models.Order, error)
	getOrderItemsFn func(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error)
	listByUserIDFn  func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error)
	listAllFn       func(ctx context.Context, limit, offset int) ([]*models.Order, error)
	cancelFn        func(ctx context.Context, id uuid.UUID) error
	updateStatusFn  func(ctx context.Context, id uuid.UUID, status string) error
}

func (m *orderRepoMock) Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
	if m.createFn != nil {
		return m.createFn(ctx, order, items)
	}

	return nil
}

func (m *orderRepoMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (m *orderRepoMock) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error) {
	if m.getOrderItemsFn != nil {
		return m.getOrderItemsFn(ctx, orderID)
	}

	return nil, nil
}

func (m *orderRepoMock) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID, limit, offset)
	}

	return nil, nil
}

func (m *orderRepoMock) ListAll(ctx context.Context, limit, offset int) ([]*models.Order, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, limit, offset)
	}

	return nil, nil
}

func (m *orderRepoMock) Cancel(ctx context.Context, id uuid.UUID) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, id)
	}

	return nil
}

func (m *orderRepoMock) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}

	return nil
}

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
	_ repository.OrderRepository   = (*orderRepoMock)(nil)
	_ repository.CartRepository    = (*cartRepoMock)(nil)
	_ repository.ProductRepository = (*productRepoMock)(nil)
)

func TestCreateOrderDeductsStockAndClearsCart(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	product := &models.Product{
		ID:    productID,
		Name:  "USB-C Hub",
		Price: 49.99,
		Stock: 5,
	}

	var updatedStock int
	cleared := false
	created := false

	svc := NewService(&orderRepoMock{
		createFn: func(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
			created = true
			if len(items) != 1 {
				t.Fatalf("expected 1 order item, got %d", len(items))
			}

			if items[0].Quantity != 2 {
				t.Fatalf("expected quantity 2, got %d", items[0].Quantity)
			}

			return nil
		},
	}, &cartRepoMock{
		getByUserIDFn: func(ctx context.Context, requestedUserID uuid.UUID) ([]*models.CartItemWithProduct, error) {
			return []*models.CartItemWithProduct{
				{
					CartItem: models.CartItem{
						UserID:    requestedUserID,
						ProductID: productID,
						Quantity:  2,
					},
					ProductPrice: product.Price,
					ProductStock: product.Stock,
				},
			}, nil
		},
		clearFn: func(ctx context.Context, requestedUserID uuid.UUID) error {
			cleared = true
			return nil
		},
	}, &productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return &models.Product{
				ID:    product.ID,
				Name:  product.Name,
				Price: product.Price,
				Stock: product.Stock,
			}, nil
		},
		updateFn: func(ctx context.Context, updated *models.Product) error {
			updatedStock = updated.Stock
			return nil
		},
	})

	_, err := svc.CreateOrder(context.Background(), userID, CreateOrderRequest{
		ShippingAddress: models.ShippingAddress{
			Street:     "Main street 1",
			City:       "Kyiv",
			PostalCode: "01001",
			Country:    "Ukraine",
		},
		PaymentMethod: "card",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !created {
		t.Fatalf("expected order to be created")
	}

	if !cleared {
		t.Fatalf("expected cart to be cleared after order creation")
	}

	if updatedStock != 3 {
		t.Fatalf("expected stock to be reduced to 3, got %d", updatedStock)
	}
}

func TestCreateOrderRejectsInsufficientStock(t *testing.T) {
	productID := uuid.New()
	orderCreated := false

	svc := NewService(&orderRepoMock{
		createFn: func(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
			orderCreated = true
			return nil
		},
	}, &cartRepoMock{
		getByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*models.CartItemWithProduct, error) {
			return []*models.CartItemWithProduct{
				{
					CartItem: models.CartItem{
						UserID:    userID,
						ProductID: productID,
						Quantity:  4,
					},
					ProductPrice: 49.99,
					ProductStock: 2,
				},
			}, nil
		},
	}, &productRepoMock{})

	_, err := svc.CreateOrder(context.Background(), uuid.New(), CreateOrderRequest{
		ShippingAddress: models.ShippingAddress{
			Street:     "Main street 1",
			City:       "Kyiv",
			PostalCode: "01001",
			Country:    "Ukraine",
		},
		PaymentMethod: "cash",
	})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}

	if orderCreated {
		t.Fatalf("did not expect order creation when stock is insufficient")
	}
}

func TestCancelOrderRestoresStock(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()
	cancelled := false
	updated := false

	svc := NewService(&orderRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			return &models.Order{
				ID:     orderID,
				UserID: userID,
				Status: "pending",
			}, nil
		},
		getOrderItemsFn: func(ctx context.Context, requestedOrderID uuid.UUID) ([]*models.OrderItem, error) {
			return []*models.OrderItem{
				{
					OrderID:   requestedOrderID,
					ProductID: productID,
					Quantity:  2,
				},
			}, nil
		},
		cancelFn: func(ctx context.Context, id uuid.UUID) error {
			cancelled = true
			return nil
		},
	}, &cartRepoMock{}, &productRepoMock{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Product, error) {
			return &models.Product{
				ID:    productID,
				Name:  "USB-C Hub",
				Stock: 3,
			}, nil
		},
		updateFn: func(ctx context.Context, product *models.Product) error {
			updated = true
			if product.Stock != 5 {
				t.Fatalf("expected restored stock 5, got %d", product.Stock)
			}

			return nil
		},
	})

	err := svc.CancelOrder(context.Background(), userID, orderID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !updated {
		t.Fatalf("expected stock restoration update")
	}

	if !cancelled {
		t.Fatalf("expected order to be cancelled")
	}
}
