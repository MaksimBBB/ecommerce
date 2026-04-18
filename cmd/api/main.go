package main

import (
	"context"
	"ecommerce/internal/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	httpHandler "ecommerce/internal/handler/http"
	postgres "ecommerce/internal/repository/postgres"
	authService "ecommerce/internal/service/auth"
	userService "ecommerce/internal/service/user"
)

// @title Ecommerce API
// @version 1.0
// @host localhost:8080
// @BasePath/api/v1

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables: %v", err)
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "ecommerce_db")
	dbSslMode := getEnv("DB_SSLMODE", "disable")
	serverPort := getEnv("SERVER_PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")

	dbconfig := db.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSslMode,
	}

	database, err := db.NewDB(dbconfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database established")

	log.Println("Database connection established")
	userRepo := postgres.NewUserRepository(database)

	authSvc := authService.NewService(userRepo, jwtSecret)
	userSvc := userService.NewService(userRepo)

	router := httpHandler.NewRouter(httpHandler.RouterConfig{
		AuthService: authSvc,
		UserService: userSvc,
	})

	server := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on http://localhost:%s", serverPort)
		log.Printf("API documentation available at http://localhost:%s/api/v1", serverPort)
		log.Printf("Health check: http://localhost%s/health", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue

}
