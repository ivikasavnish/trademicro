package models

import (
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Username       string `gorm:"unique" json:"username"`
	Password       string `json:"password,omitempty" gorm:"column:password"`
	HashedPassword string `json:"-" gorm:"column:hashed_password"`
	Email          string `json:"email"`
	Role           string `json:"role"`
}

// Claims for JWT authentication
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
