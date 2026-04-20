package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	models "ecommerce/internal/domain"
	repository "ecommerce/internal/repository/postgres"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
}

// RegisterRequest is the DTO for user registration
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
}

// LoginRequest is the DTO for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse contains authentication tokens
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

// Claims represents JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   string    `json:"type"`
	jwt.RegisteredClaims
}

type service struct {
	userRepo      repository.UserRepository
	jwtSecret     string
	tokenDuration time.Duration
}

// NewService creates a new auth service
func NewService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &service{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		tokenDuration: 24 * time.Hour, // 24 hours
	}
}

// register creates a new user account
func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" {
		return nil, ErrEmailRequired
	}

	if req.Password == "" {
		return nil, ErrPasswordRequired
	}

	if len(req.Password) < 8 {
		return nil, ErrWeakPassword
	}

	if len(req.FirstName) < 2 {
		return nil, ErrInvalidFirstName
	}

	if len(req.LastName) < 2 {
		return nil, ErrInvalidLastName
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FirstName:    req.FirstName,
		Surname:      req.LastName,
		Role:         "user",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.buildAuthResponse(user)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" {
		return nil, ErrEmailRequired
	}

	if req.Password == "" {
		return nil, ErrPasswordRequired
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.buildAuthResponse(user)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims, err := s.validate(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return s.buildAuthResponse(user)
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	claims, err := s.validate(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *service) buildAuthResponse(user *models.User) (*AuthResponse, error) {
	accessToken, err := s.generateToken(user, "access", s.tokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *service) generateToken(user *models.User, tokenType string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"type":    tokenType,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *service) validate(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "expired") {
			return nil, ErrTokenExpired
		}

		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userIDString, ok := mapClaims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	email, _ := mapClaims["email"].(string)
	role, _ := mapClaims["role"].(string)
	tokenType, _ := mapClaims["type"].(string)

	return &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: getExpiresAt(mapClaims),
		},
	}, nil
}

func getExpiresAt(claims jwt.MapClaims) *jwt.NumericDate {
	expValue, ok := claims["exp"]
	if !ok {
		return nil
	}

	switch v := expValue.(type) {
	case float64:
		return jwt.NewNumericDate(time.Unix(int64(v), 0))
	case int64:
		return jwt.NewNumericDate(time.Unix(v, 0))
	case json.Number:
		n, err := v.Int64()
		if err == nil {
			return jwt.NewNumericDate(time.Unix(n, 0))
		}
	}

	return nil
}
