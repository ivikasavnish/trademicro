package services

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/models"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Authenticate(username, password string) (models.User, error)
	GenerateToken(user models.User, secretKey []byte) (string, error)
}

// authService implements the AuthService interface
type authService struct {
	db *gorm.DB
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB) AuthService {
	return &authService{
		db: db,
	}
}

// Authenticate verifies user credentials and returns the user if valid
func (s *authService) Authenticate(username, password string) (models.User, error) {
	var user models.User
	result := s.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return models.User{}, result.Error
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// GenerateToken creates a new JWT token for the user
func (s *authService) GenerateToken(user models.User, secretKey []byte) (string, error) {
	// Create a JWT token
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &models.Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
