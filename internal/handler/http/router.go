package http

import (
	authService "ecommerce/internal/service/auth"
	userService "ecommerce/internal/service/user"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RouterConfig struct {
	AuthService authService.AuthService
	UserService userService.UserService
}

func NewRouter(config RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		authHandler := NewAuthHandler(config.AuthService)
		authHandler.RegisterRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(config.AuthService))
			userHandler := NewUserHandler(config.UserService)
			userHandler.RegisterRoutes(r)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(RequireAdmin)
	})

	return r
}
