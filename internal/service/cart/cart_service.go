package service

import (
	"context"
	models "ecommerce/internal/domain"
	repository "ecommerce/internal/repository/postgres"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CartService interface {
	AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*models.CartItem, error)
	GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error)
	UpdateQuantity(ctx context.Context, userID, itemID uuid.UUID, quantity int) error
	RemoveItem(ctx context.Context, userID, itemID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	cartRepo repository.CartRepository
}

func NewService(cartRepo repository.CartRepository) CartService {
	return &service{cartRepo: cartRepo}
}

type AddItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
}

type CartResponse struct {
	Items      []*models.CartItemWithProduct `json:"items"`
	TotalItems int                           `json:"total_items"`
	TotalPrice float64                       `json:"total_price"`
}

func (s *service) AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*models.CartItem, error) {
	if req.ProductID == uuid.Nil {
		return nil, ErrInvalidCartItem
	}

	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	existingItem, err := s.cartRepo.GetItem(ctx, userID, req.ProductID)
	if err == nil && existingItem != nil {
		newQuantity := existingItem.Quantity + req.Quantity
		if err := s.cartRepo.UpdateQuantity(ctx, existingItem.ID, newQuantity); err != nil {
			if isNotFoundError(err) {
				return nil, ErrCartItemNotFound
			}

			return nil, fmt.Errorf("failed to update cart item quantity: %w", err)
		}

		existingItem.Quantity = newQuantity
		return existingItem, nil
	}

	if err != nil && !isNotFoundError(err) {
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	item := &models.CartItem{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	if err := s.cartRepo.AddItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add cart item: %w", err)
	}

	return item, nil
}

func (s *service) GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error) {
	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	response := &CartResponse{
		Items: items,
	}

	for _, item := range items {
		response.TotalItems += item.Quantity
		response.TotalPrice += item.ProductPrice * float64(item.Quantity)
	}

	return response, nil
}

func (s *service) UpdateQuantity(ctx context.Context, userID, itemID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	item, err := s.findUserCartItem(ctx, userID, itemID)
	if err != nil {
		return err
	}

	if err := s.cartRepo.UpdateQuantity(ctx, item.ID, quantity); err != nil {
		if isNotFoundError(err) {
			return ErrCartItemNotFound
		}

		return fmt.Errorf("failed to update cart item quantity: %w", err)
	}

	return nil
}

func (s *service) RemoveItem(ctx context.Context, userID, itemID uuid.UUID) error {
	item, err := s.findUserCartItem(ctx, userID, itemID)
	if err != nil {
		return err
	}

	if err := s.cartRepo.RemoveItem(ctx, item.ID); err != nil {
		if isNotFoundError(err) {
			return ErrCartItemNotFound
		}

		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	return nil
}

func (s *service) ClearCart(ctx context.Context, userID uuid.UUID) error {
	if err := s.cartRepo.Clear(ctx, userID); err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

func (s *service) findUserCartItem(ctx context.Context, userID, itemID uuid.UUID) (*models.CartItemWithProduct, error) {
	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	for _, item := range items {
		if item.ID == itemID {
			return item, nil
		}
	}

	return nil, ErrCartItemNotFound
}

func isNotFoundError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}
