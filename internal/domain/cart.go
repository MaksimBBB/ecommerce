package models

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	ProductID uuid.UUID `db:"product_id" json:"product_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CartItemWithProduct struct {
	CartItem
	ProductName  string  `db:"product_name" json:"product_name"`
	ProductPrice float64 `db:"product_price" json:"product_price"`
	ProductStock int     `db:"product_stock" json:"product_stock"`
}
