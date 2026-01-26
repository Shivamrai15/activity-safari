package models

import "github.com/golang-jwt/jwt/v5"

type TokenPayload struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}