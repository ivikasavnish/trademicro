// filepath: /Users/vikasavnish/trademicro/internal/services/broker_token_service.go
package services

import (
	"time"

	"github.com/vikasavnish/trademicro/internal/models"
	"gorm.io/gorm"
)

// BrokerTokenService defines methods for managing broker tokens
type BrokerTokenService interface {
	// Broker methods
	GetAllBrokers() ([]models.Broker, error)
	GetBrokerByID(id uint) (*models.Broker, error)
	GetBrokerByCode(code string) (*models.Broker, error)
	CreateBroker(broker models.Broker) (*models.Broker, error)
	UpdateBroker(broker models.Broker) (*models.Broker, error)
	DeleteBroker(id uint) error

	// Broker token methods
	GetAllBrokerTokens() ([]models.BrokerToken, error)
	GetBrokerTokenByID(id uint) (*models.BrokerToken, error)
	GetBrokerTokensByUserID(userID uint) ([]models.BrokerToken, error)
	GetBrokerTokensByFamilyMemberID(familyMemberID uint) ([]models.BrokerToken, error)
	CreateBrokerToken(token models.BrokerToken) (*models.BrokerToken, error)
	UpdateBrokerToken(token models.BrokerToken) (*models.BrokerToken, error)
	DeleteBrokerToken(id uint) error
}

// BrokerTokenServiceImpl implements BrokerTokenService
type BrokerTokenServiceImpl struct {
	DB *gorm.DB
}

// NewBrokerTokenService creates a new broker token service
func NewBrokerTokenService(db *gorm.DB) BrokerTokenService {
	return &BrokerTokenServiceImpl{DB: db}
}

// BROKER METHODS

// GetAllBrokers returns all brokers
func (s *BrokerTokenServiceImpl) GetAllBrokers() ([]models.Broker, error) {
	var brokers []models.Broker
	result := s.DB.Find(&brokers)
	if result.Error != nil {
		return nil, result.Error
	}
	return brokers, nil
}

// GetBrokerByID returns a broker by ID
func (s *BrokerTokenServiceImpl) GetBrokerByID(id uint) (*models.Broker, error) {
	var broker models.Broker
	result := s.DB.First(&broker, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &broker, nil
}

// GetBrokerByCode returns a broker by code
func (s *BrokerTokenServiceImpl) GetBrokerByCode(code string) (*models.Broker, error) {
	var broker models.Broker
	result := s.DB.Where("code = ?", code).First(&broker)
	if result.Error != nil {
		return nil, result.Error
	}
	return &broker, nil
}

// CreateBroker creates a new broker
func (s *BrokerTokenServiceImpl) CreateBroker(broker models.Broker) (*models.Broker, error) {
	broker.CreatedAt = time.Now()
	broker.UpdatedAt = time.Now()

	result := s.DB.Create(&broker)
	if result.Error != nil {
		return nil, result.Error
	}

	return &broker, nil
}

// UpdateBroker updates an existing broker
func (s *BrokerTokenServiceImpl) UpdateBroker(broker models.Broker) (*models.Broker, error) {
	// Verify the broker exists
	var existingBroker models.Broker
	if err := s.DB.First(&existingBroker, broker.ID).Error; err != nil {
		return nil, err
	}

	broker.UpdatedAt = time.Now()

	if err := s.DB.Save(&broker).Error; err != nil {
		return nil, err
	}

	return &broker, nil
}

// DeleteBroker deletes a broker
func (s *BrokerTokenServiceImpl) DeleteBroker(id uint) error {
	return s.DB.Delete(&models.Broker{}, id).Error
}

// BROKER TOKEN METHODS

// GetAllBrokerTokens returns all broker tokens
func (s *BrokerTokenServiceImpl) GetAllBrokerTokens() ([]models.BrokerToken, error) {
	var tokens []models.BrokerToken
	result := s.DB.Preload("Broker").Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// GetBrokerTokenByID returns a broker token by ID
func (s *BrokerTokenServiceImpl) GetBrokerTokenByID(id uint) (*models.BrokerToken, error) {
	var token models.BrokerToken
	result := s.DB.Preload("Broker").First(&token, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &token, nil
}

// GetBrokerTokensByUserID returns all broker tokens for a user
func (s *BrokerTokenServiceImpl) GetBrokerTokensByUserID(userID uint) ([]models.BrokerToken, error) {
	var tokens []models.BrokerToken
	result := s.DB.Where("user_id = ? AND family_member_id IS NULL", userID).Preload("Broker").Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// GetBrokerTokensByFamilyMemberID returns all broker tokens for a family member
func (s *BrokerTokenServiceImpl) GetBrokerTokensByFamilyMemberID(familyMemberID uint) ([]models.BrokerToken, error) {
	var tokens []models.BrokerToken
	result := s.DB.Where("family_member_id = ?", familyMemberID).Preload("Broker").Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// CreateBrokerToken creates a new broker token
func (s *BrokerTokenServiceImpl) CreateBrokerToken(token models.BrokerToken) (*models.BrokerToken, error) {
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()

	// If expires_at is not set, default to 7 days from now
	if token.ExpiresAt.IsZero() {
		token.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	result := s.DB.Create(&token)
	if result.Error != nil {
		return nil, result.Error
	}

	// Load the broker information
	if err := s.DB.Preload("Broker").First(&token, token.ID).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

// UpdateBrokerToken updates an existing broker token
func (s *BrokerTokenServiceImpl) UpdateBrokerToken(token models.BrokerToken) (*models.BrokerToken, error) {
	// Verify the token exists
	var existingToken models.BrokerToken
	if err := s.DB.First(&existingToken, token.ID).Error; err != nil {
		return nil, err
	}

	token.UpdatedAt = time.Now()

	if err := s.DB.Save(&token).Error; err != nil {
		return nil, err
	}

	// Load the broker information
	if err := s.DB.Preload("Broker").First(&token, token.ID).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

// DeleteBrokerToken deletes a broker token
func (s *BrokerTokenServiceImpl) DeleteBrokerToken(id uint) error {
	return s.DB.Delete(&models.BrokerToken{}, id).Error
}
