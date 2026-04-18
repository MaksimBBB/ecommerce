package service

import "errors"

var (
	ErrInvalidProduct     = errors.New("invalid product")
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidPrice       = errors.New("invalid product price")
	ErrInvalidStock       = errors.New("invalid product stock")
)
