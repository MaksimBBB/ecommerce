package repository

import (
	"context"
	"fmt"
	"strings"

	database "ecommerce/internal/db"
	models "ecommerce/internal/domain"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	List(ctx context.Context, filter models.ListFilter) ([]*models.Product, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Product, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type productRepo struct {
	db *database.DB
}

func NewProductRepository(db *database.DB) ProductRepository {
	return &productRepo{db: db}
}

func (p *productRepo) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (id, name, description, price, stock, category_id, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	product.ID = uuid.New()
	return p.db.QueryRowContext(
		ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.ImageURL,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (p *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (p *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT id, name, description, price, stock, category_id, image_url, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	err := p.db.GetContext(ctx, &product, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by ID: %w", err)
	}

	return &product, nil
}

func (p *productRepo) List(ctx context.Context, filter models.ListFilter) ([]*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category_id, image_url, created_at, updated_at
		FROM products
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.MinPrice != nil {
		query += fmt.Sprintf(" AND price >= $%d", argIndex)
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		query += fmt.Sprintf(" AND price <= $%d", argIndex)
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.InStock != nil && *filter.InStock {
		query += " AND stock > 0"
	}

	orderBy := "created_at DESC"
	switch strings.ToLower(filter.OrderBy) {
	case "price asc":
		orderBy = "price ASC"
	case "price desc":
		orderBy = "price DESC"
	case "name asc":
		orderBy = "name ASC"
	case "name desc":
		orderBy = "name DESC"
	case "created_at asc":
		orderBy = "created_at ASC"
	case "created_at desc", "":
		orderBy = "created_at DESC"
	}

	query += " ORDER BY " + orderBy

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	var products []*models.Product
	err := p.db.SelectContext(ctx, &products, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

func (p *productRepo) Search(ctx context.Context, query string, limit int, offset int) ([]*models.Product, error) {
	filter := models.ListFilter{
		Search: query,
		Limit:  limit,
		Offset: offset,
	}

	return p.List(ctx, filter)
}

func (p *productRepo) ListCategories(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY name ASC
	`

	var categories []*models.Category
	err := p.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

func (p *productRepo) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, category_id = $5, image_url = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	return p.db.QueryRowContext(
		ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.ImageURL,
		product.ID,
	).Scan(&product.UpdatedAt)
}
