package service

import (
	"context"
	models "ecommerce/internal/domain"
	repository "ecommerce/internal/repository/postgres"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req CreateProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error)
	ListProducts(ctx context.Context, filter models.ListFilter) (*ProductListResponse, error)
	SearchProducts(ctx context.Context, query string, limit, offset int) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	CheckAvailability(ctx context.Context, id uuid.UUID, quantity int) (bool, error)
	ReserveStock(ctx context.Context, id uuid.UUID, quantity int) error
	ReleaseStock(ctx context.Context, id uuid.UUID, quantity int) error
}

type service struct {
	productRepo repository.ProductRepository
}

type CreateProductRequest struct {
	Name        string     `json:"name" validate:"required,min=3,max=255"`
	Description string     `json:"description" validate:"max=1000"`
	Price       float64    `json:"price" validate:"required,gt=0"`
	Stock       int        `json:"stock" validate:"required,gte=0"`
	CategoryID  *uuid.UUID `json:"category_id" validate:"omitempty,uuid"`
	ImageURL    string     `json:"image_url" validate:"omitempty,url"`
}

type UpdateProductRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Price       *float64   `json:"price"`
	Stock       *int       `json:"stock"`
	CategoryID  *uuid.UUID `json:"category_id"`
	ImageURL    *string    `json:"image_url"`
}

type ProductListResponse struct {
	Products []*models.Product `json:"products"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

func NewService(productRepo repository.ProductRepository) ProductService {
	return &service{productRepo: productRepo}
}

func (s *service) CreateProduct(ctx context.Context, req CreateProductRequest) (*models.Product, error) {
	product := &models.Product{
		Name:       strings.TrimSpace(req.Name),
		Price:      req.Price,
		Stock:      req.Stock,
		CategoryID: req.CategoryID,
	}

	if description := strings.TrimSpace(req.Description); description != "" {
		product.Description = &description
	}

	if imageURL := strings.TrimSpace(req.ImageURL); imageURL != "" {
		product.ImageURL = &imageURL
	}

	if err := validateProduct(product); err != nil {
		return nil, err
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *service) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrProductNotFound
		}

		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (s *service) ListProducts(ctx context.Context, filter models.ListFilter) (*ProductListResponse, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.OrderBy = strings.ToLower(strings.TrimSpace(filter.OrderBy))

	if filter.MinPrice != nil && filter.MaxPrice != nil && *filter.MinPrice > *filter.MaxPrice {
		*filter.MinPrice, *filter.MaxPrice = *filter.MaxPrice, *filter.MinPrice
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	products, err := s.productRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return &ProductListResponse{
		Products: products,
		Total:    len(products),
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}, nil
}

func (s *service) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	products, err := s.productRepo.Search(ctx, strings.TrimSpace(query), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

func (s *service) ListCategories(ctx context.Context) ([]*models.Category, error) {
	categories, err := s.productRepo.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

func (s *service) UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*models.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrProductNotFound
		}

		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if req.Name != nil {
		product.Name = strings.TrimSpace(*req.Name)
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			product.Description = nil
		} else {
			product.Description = &description
		}
	}

	if req.Price != nil {
		product.Price = *req.Price
	}

	if req.Stock != nil {
		product.Stock = *req.Stock
	}

	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}

	if req.ImageURL != nil {
		imageURL := strings.TrimSpace(*req.ImageURL)
		if imageURL == "" {
			product.ImageURL = nil
		} else {
			product.ImageURL = &imageURL
		}
	}

	if err := validateProduct(product); err != nil {
		return nil, err
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		if isNotFoundError(err) {
			return nil, ErrProductNotFound
		}

		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (s *service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		if isNotFoundError(err) {
			return ErrProductNotFound
		}

		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s *service) CheckAvailability(ctx context.Context, id uuid.UUID, quantity int) (bool, error) {
	product, err := s.GetProduct(ctx, id)
	if err != nil {
		return false, err
	}

	return product.Stock >= quantity, nil
}

func (s *service) ReserveStock(ctx context.Context, id uuid.UUID, quantity int) error {
	product, err := s.GetProduct(ctx, id)
	if err != nil {
		return err
	}

	product.Stock -= quantity
	if product.Stock < 0 {
		return ErrInvalidStock
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	return nil
}

func (s *service) ReleaseStock(ctx context.Context, id uuid.UUID, quantity int) error {
	product, err := s.GetProduct(ctx, id)
	if err != nil {
		return err
	}

	product.Stock += quantity
	if err := s.productRepo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to release stock: %w", err)
	}

	return nil
}

func validateProduct(product *models.Product) error {
	if product == nil {
		return ErrProductNotFound
	}

	if strings.TrimSpace(product.Name) == "" {
		return ErrInvalidProductName
	}

	if product.Price < 0 {
		return ErrInvalidPrice
	}

	if product.Stock < 0 {
		return ErrInvalidStock
	}

	return nil
}

func isNotFoundError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}
