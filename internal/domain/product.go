package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description *string    `db:"description" json:"description"`
	Price       float64    `db:"price" json:"price"`
	Stock       int        `db:"stock" json:"stock"`
	CategoryID  *uuid.UUID `db:"category_id" json:"category_id,omitempty"`
	ImageURL    *string    `db:"image_url" json:"image_url,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type ListFilter struct {
	CategoryID *uuid.UUID
	MinPrice   *float64
	MaxPrice   *float64
	Search     string
	InStock    *bool
	Limit      int
	Offset     int
	OrderBy    string
}
