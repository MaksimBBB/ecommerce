package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	database "ecommerce/internal/db"
	models "ecommerce/internal/domain"

	"github.com/google/uuid"
)

type CartRepository interface {
	AddItem(ctx context.Context, item *models.CartItem) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.CartItemWithProduct, error)
	UpdateQuantity(ctx context.Context, id uuid.UUID, quantity int) error
	RemoveItem(ctx context.Context, id uuid.UUID) error
	Clear(ctx context.Context, userID uuid.UUID) error
	GetItem(ctx context.Context, userID, productID uuid.UUID) (*models.CartItem, error)
}

type cartRepo struct {
	db *database.DB
}

func NewCartRepository(db *database.DB) CartRepository {
	return &cartRepo{db: db}
}

func (c *cartRepo) AddItem(ctx context.Context, item *models.CartItem) error {
	query := `
		INSERT INTO cart_items (id, user_id, product_id, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	item.ID = uuid.New()
	return c.db.QueryRowContext(
		ctx, query,
		item.ID,
		item.UserID,
		item.ProductID,
		item.Quantity,
	).Scan(&item.ID, &item.CreatedAt)
}

func (c *cartRepo) Clear(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := c.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

func (c *cartRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.CartItemWithProduct, error) {
	query := `
		SELECT
			ci.id,
			ci.user_id,
			ci.product_id,
			ci.quantity,
			ci.created_at,
			p.name AS product_name,
			p.price AS product_price,
			p.stock AS product_stock
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.user_id = $1
		ORDER BY ci.created_at DESC
	`

	var items []*models.CartItemWithProduct
	err := c.db.SelectContext(ctx, &items, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items by user ID: %w", err)
	}

	return items, nil
}

func (c *cartRepo) GetItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*models.CartItem, error) {
	var item models.CartItem
	query := `
		SELECT id, user_id, product_id, quantity, created_at
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2
	`

	err := c.db.GetContext(ctx, &item, query, userID, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cart item not found")
		}

		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	return &item, nil
}

func (c *cartRepo) RemoveItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE id = $1`
	result, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}

func (c *cartRepo) UpdateQuantity(ctx context.Context, id uuid.UUID, quantity int) error {
	query := `UPDATE cart_items SET quantity = $1 WHERE id = $2`
	result, err := c.db.ExecContext(ctx, query, quantity, id)
	if err != nil {
		return fmt.Errorf("failed to update cart item quantity: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}
