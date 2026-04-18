package repository

import (
	"context"
	"encoding/json"
	"fmt"

	database "ecommerce/internal/db"
	models "ecommerce/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error)
	ListAll(ctx context.Context, limit, offset int) ([]*models.Order, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type orderRepo struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (o *orderRepo) Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
	tx, err := o.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin order transaction: %w", err)
	}
	defer tx.Rollback()

	shippingAddress, err := json.Marshal(order.ShippingAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal shipping address: %w", err)
	}

	orderQuery := `
		INSERT INTO orders (id, user_id, status, total_amount, shipping_address, payment_method)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	order.ID = uuid.New()
	err = tx.QueryRowContext(
		ctx, orderQuery,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalAmount,
		shippingAddress,
		order.PaymentMethod,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	for _, item := range items {
		item.ID = uuid.New()
		item.OrderID = order.ID

		err = tx.QueryRowContext(
			ctx, itemQuery,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.Price,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit order transaction: %w", err)
	}

	return nil
}

func (o *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	var shippingAddress []byte

	query := `
		SELECT id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	err := o.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalAmount,
		&shippingAddress,
		&order.PaymentMethod,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	if err := json.Unmarshal(shippingAddress, &order.ShippingAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
	}

	return &order, nil
}

func (o *orderRepo) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	var items []*models.OrderItem
	err := o.db.SelectContext(ctx, &items, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return items, nil
}

func (o *orderRepo) ListAll(ctx context.Context, limit int, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := o.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var shippingAddress []byte

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&shippingAddress,
			&order.PaymentMethod,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if err := json.Unmarshal(shippingAddress, &order.ShippingAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}

	return orders, nil
}

func (o *orderRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := o.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders by user ID: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var shippingAddress []byte

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&shippingAddress,
			&order.PaymentMethod,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if err := json.Unmarshal(shippingAddress, &order.ShippingAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}

	return orders, nil
}

func (o *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := o.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (o *orderRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	return o.UpdateStatus(ctx, id, "cancelled")
}
