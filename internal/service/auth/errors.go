package service

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidFirstName    = errors.New("invalid first name")
	ErrInvalidLastName     = errors.New("invalid last name")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredToken        = errors.New("token expired")
)
