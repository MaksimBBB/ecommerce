package service

import "errors"

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrInvalidOrder           = errors.New("invalid order")
	ErrInvalidOrderItem       = errors.New("invalid order item")
	ErrInvalidOrderStatus     = errors.New("invalid order status")
	ErrEmptyOrderItems        = errors.New("order must contain at least one item")
	ErrInvalidShippingAddress = errors.New("invalid shipping address")
	ErrInvalidPaymentMethod   = errors.New("invalid payment method")
)
