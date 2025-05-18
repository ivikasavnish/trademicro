package services

import (
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/models"
)

// UserService defines the interface for user-related operations
type UserService interface {
	GetUsers() ([]models.User, error)
	GetUserByUsername(username string) (models.User, error)
	GetUserIDByUsername(username string) (uint, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUser(id uint, user models.User) (models.User, error)
	IsUserAdmin(userID uint) (bool, error)
}

// userService implements the UserService interface
type userService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) UserService {
	return &userService{
		db: db,
	}
}

// GetUsers returns all users
func (s *userService) GetUsers() ([]models.User, error) {
	var users []models.User
	result := s.db.Select("id, username, email, role").Find(&users) // Exclude password field
	return users, result.Error
}

// GetUserByUsername returns a user by username
func (s *userService) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	result := s.db.Where("username = ?", username).First(&user)
	return user, result.Error
}

// CreateUser creates a new user
func (s *userService) CreateUser(user models.User) (models.User, error) {
	result := s.db.Create(&user)
	return user, result.Error
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(id uint, user models.User) (models.User, error) {
	var existingUser models.User
	if err := s.db.First(&existingUser, id).Error; err != nil {
		return models.User{}, err
	}

	// Update allowed fields
	existingUser.Email = user.Email
	existingUser.Role = user.Role

	if user.HashedPassword != "" {
		existingUser.HashedPassword = user.HashedPassword
	}

	result := s.db.Save(&existingUser)
	return existingUser, result.Error
}

// GetUserIDByUsername retrieves a user's ID by their username
func (s *userService) GetUserIDByUsername(username string) (uint, error) {
	var user models.User
	result := s.db.Select("id").Where("username = ?", username).First(&user)
	if result.Error != nil {
		return 0, result.Error
	}
	return user.ID, nil
}

// IsUserAdmin checks if a user has admin role
func (s *userService) IsUserAdmin(userID uint) (bool, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false, err
	}

	return user.Role == "admin", nil
}
