package service

import (
	"context"
	models "ecommerce/internal/domain"
	cartrepo "ecommerce/internal/repository/postgres"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error)
	GetOrder(ctx context.Context, userID, orderID uuid.UUID) (*OrderResponse, error)
	ListOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error)
	ListAllOrders(ctx context.Context, limit, offset int) ([]*models.Order, error)
	CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error
}

type service struct {
	orderRepo   cartrepo.OrderRepository
	cartRepo    cartrepo.CartRepository
	productRepo cartrepo.ProductRepository
}

func NewService(orderRepo cartrepo.OrderRepository, cartRepo cartrepo.CartRepository, productRepo cartrepo.ProductRepository) OrderService {
	return &service{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

type CreateOrderRequest struct {
	ShippingAddress models.ShippingAddress `json:"shipping_address" validate:"required"`
	PaymentMethod   string                 `json:"payment_method" validate:"required,oneof=cash card"`
}

type OrderResponse struct {
	Order *models.Order       `json:"order"`
	Items []*models.OrderItem `json:"items"`
}

func (s *service) CreateOrder(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error) {
	if err := validateCreateOrderRequest(req); err != nil {
		return nil, err
	}

	cartItems, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, ErrEmptyOrderItems
	}

	orderItems := make([]*models.OrderItem, 0, len(cartItems))
	totalAmount := 0.0

	for _, item := range cartItems {
		if item == nil || item.Quantity <= 0 || item.ProductPrice < 0 {
			return nil, ErrInvalidOrderItem
		}

		if item.ProductStock < item.Quantity {
			return nil, ErrInsufficientStock
		}

		orderItems = append(orderItems, &models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.ProductPrice,
		})
		totalAmount += item.ProductPrice * float64(item.Quantity)
	}

	updatedProducts := make([]*models.Product, 0, len(cartItems))
	for _, item := range cartItems {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			if isNotFoundError(err) {
				return nil, ErrInvalidOrderItem
			}

			return nil, fmt.Errorf("failed to get product for order: %w", err)
		}

		if product.Stock < item.Quantity {
			return nil, ErrInsufficientStock
		}

		product.Stock -= item.Quantity
		updatedProducts = append(updatedProducts, product)
	}

	for _, product := range updatedProducts {
		if err := s.productRepo.Update(ctx, product); err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	order := &models.Order{
		UserID:          userID,
		Status:          "pending",
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   strings.TrimSpace(req.PaymentMethod),
	}

	if err := s.orderRepo.Create(ctx, order, orderItems); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if err := s.cartRepo.Clear(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to clear cart after order creation: %w", err)
	}

	return &OrderResponse{
		Order: order,
		Items: orderItems,
	}, nil
}

func (s *service) GetOrder(ctx context.Context, userID, orderID uuid.UUID) (*OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrOrderNotFound
		}

		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrOrderNotFound
		}

		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return &OrderResponse{
		Order: order,
		Items: items,
	}, nil
}

func (s *service) ListOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	limit, offset = normalizePagination(limit, offset)

	orders, err := s.orderRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user orders: %w", err)
	}

	return orders, nil
}

func (s *service) ListAllOrders(ctx context.Context, limit, offset int) ([]*models.Order, error) {
	limit, offset = normalizePagination(limit, offset)

	orders, err := s.orderRepo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, nil
}

func (s *service) CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if isNotFoundError(err) {
			return ErrOrderNotFound
		}

		return fmt.Errorf("failed to get order before cancel: %w", err)
	}

	if order.UserID != userID {
		return ErrOrderNotFound
	}

	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		if isNotFoundError(err) {
			return ErrOrderNotFound
		}

		return fmt.Errorf("failed to get order items before cancel: %w", err)
	}

	for _, item := range items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			if isNotFoundError(err) {
				return ErrInvalidOrderItem
			}

			return fmt.Errorf("failed to get product for cancel: %w", err)
		}

		product.Stock += item.Quantity
		if err := s.productRepo.Update(ctx, product); err != nil {
			return fmt.Errorf("failed to restore product stock: %w", err)
		}
	}

	if err := s.orderRepo.Cancel(ctx, orderID); err != nil {
		if isNotFoundError(err) {
			return ErrOrderNotFound
		}

		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

func (s *service) UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	status = strings.TrimSpace(status)
	if status == "" {
		return ErrInvalidOrderStatus
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		if isNotFoundError(err) {
			return ErrOrderNotFound
		}

		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func validateCreateOrderRequest(req CreateOrderRequest) error {
	if strings.TrimSpace(req.PaymentMethod) == "" {
		return ErrInvalidPaymentMethod
	}

	address := req.ShippingAddress
	if strings.TrimSpace(address.Street) == "" ||
		strings.TrimSpace(address.City) == "" ||
		strings.TrimSpace(address.PostalCode) == "" ||
		strings.TrimSpace(address.Country) == "" {
		return ErrInvalidShippingAddress
	}

	return nil
}

func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

func isNotFoundError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}
