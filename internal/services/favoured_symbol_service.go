package services

import (
	"errors"
	"time"

	"github.com/vikasavnish/trademicro/internal/models"
	"gorm.io/gorm"
)

type FavouredSymbolService struct {
	DB *gorm.DB
}

func NewFavouredSymbolService(db *gorm.DB) *FavouredSymbolService {
	return &FavouredSymbolService{DB: db}
}

// GetUserFavouredSymbols gets all favoured symbols for a user
func (s *FavouredSymbolService) GetUserFavouredSymbols(userID uint) ([]models.FavouredSymbol, error) {
	var favourites []models.FavouredSymbol

	// Preload the symbol data
	result := s.DB.Preload("Symbol").
		Where("user_id = ?", userID).
		Find(&favourites)

	return favourites, result.Error
}

// AddFavouredSymbol adds a symbol to user's favoured list
func (s *FavouredSymbolService) AddFavouredSymbol(userID, symbolID uint, notes string) (*models.FavouredSymbol, error) {
	// Check if symbol exists
	var symbol models.Symbol
	if err := s.DB.First(&symbol, symbolID).Error; err != nil {
		return nil, errors.New("symbol not found")
	}

	// Check if already in favourites
	var existing models.FavouredSymbol
	result := s.DB.Where("user_id = ? AND symbol_id = ?", userID, symbolID).First(&existing)

	if result.Error == nil {
		// Update existing
		existing.Notes = notes
		existing.UpdatedAt = time.Now()
		s.DB.Save(&existing)
		return &existing, nil
	}

	// Create new
	favourite := models.FavouredSymbol{
		UserID:    userID,
		SymbolID:  symbolID,
		Notes:     notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.DB.Create(&favourite).Error; err != nil {
		return nil, err
	}

	// Reload with symbol data
	s.DB.Preload("Symbol").First(&favourite, favourite.ID)
	return &favourite, nil
}

// RemoveFavouredSymbol removes a symbol from user's favoured list
func (s *FavouredSymbolService) RemoveFavouredSymbol(userID, symbolID uint) error {
	result := s.DB.Where("user_id = ? AND symbol_id = ?", userID, symbolID).
		Delete(&models.FavouredSymbol{})

	if result.RowsAffected == 0 {
		return errors.New("favoured symbol not found")
	}

	return result.Error
}

// GetFavouredSymbolByID gets a specific favoured symbol
func (s *FavouredSymbolService) GetFavouredSymbolByID(id, userID uint) (*models.FavouredSymbol, error) {
	var favourite models.FavouredSymbol

	result := s.DB.Preload("Symbol").
		Where("id = ? AND user_id = ?", id, userID).
		First(&favourite)

	if result.Error != nil {
		return nil, result.Error
	}

	return &favourite, nil
}

// UpdateFavouredSymbol updates a favoured symbol's notes
func (s *FavouredSymbolService) UpdateFavouredSymbol(id, userID uint, notes string) (*models.FavouredSymbol, error) {
	favourite, err := s.GetFavouredSymbolByID(id, userID)
	if err != nil {
		return nil, err
	}

	favourite.Notes = notes
	favourite.UpdatedAt = time.Now()

	if err := s.DB.Save(favourite).Error; err != nil {
		return nil, err
	}

	return favourite, nil
}
