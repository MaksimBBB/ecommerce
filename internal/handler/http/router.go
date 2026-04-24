package http

import (
	authService "ecommerce/internal/service/auth"
	cartService "ecommerce/internal/service/cart"
	orderService "ecommerce/internal/service/order"
	productService "ecommerce/internal/service/product"
	userService "ecommerce/internal/service/user"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RouterConfig struct {
	AuthService    authService.AuthService
	UserService    userService.UserService
	ProductService productService.ProductService
	CartService    cartService.CartService
	OrderService   orderService.OrderService
}

func NewRouter(config RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(CORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		authHandler := NewAuthHandler(config.AuthService)
		authHandler.RegisterRoutes(r)

		productHandler := NewProductHandler(config.ProductService)
		productHandler.RegisterRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(config.AuthService))
			userHandler := NewUserHandler(config.UserService)
			userHandler.RegisterRoutes(r)

			cartHandler := NewCartHandler(config.CartService)
			cartHandler.RegisterRoutes(r)

			orderHandler := NewOrderHandler(config.OrderService)
			orderHandler.RegisterRoutes(r)
		})

		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(config.AuthService))
			r.Use(RequireAdmin)

			adminHandler := NewAdminHandler(
				config.UserService,
				config.OrderService,
				config.ProductService,
			)

			adminHandler.RegisterRoutes(r)
		})
	})

	return r
}
