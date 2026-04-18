package service

import "errors"

var (
	ErrCartItemNotFound = errors.New("cart item not found")
	ErrInvalidCartItem  = errors.New("invalid cart item")
	ErrInvalidQuantity  = errors.New("invalid quantity")
)
