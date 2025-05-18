package services

import (
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/models"
)

// TradeService defines the interface for trade operations
type TradeService interface {
	GetTrades() ([]models.TradeOrder, error)
	GetTradeByID(id uint) (models.TradeOrder, error)
	CreateTrade(trade models.TradeOrder) (models.TradeOrder, error)
	UpdateTrade(id uint, trade models.TradeOrder) (models.TradeOrder, error)
}

// tradeService implements the TradeService interface
type tradeService struct {
	db *gorm.DB
}

// NewTradeService creates a new trade service
func NewTradeService(db *gorm.DB) TradeService {
	return &tradeService{
		db: db,
	}
}

// GetTrades returns all trades
func (s *tradeService) GetTrades() ([]models.TradeOrder, error) {
	var trades []models.TradeOrder
	result := s.db.Find(&trades)
	return trades, result.Error
}

// GetTradeByID returns a trade by ID
func (s *tradeService) GetTradeByID(id uint) (models.TradeOrder, error) {
	var trade models.TradeOrder
	result := s.db.First(&trade, id)
	return trade, result.Error
}

// CreateTrade creates a new trade
func (s *tradeService) CreateTrade(trade models.TradeOrder) (models.TradeOrder, error) {
	result := s.db.Create(&trade)
	return trade, result.Error
}

// UpdateTrade updates a trade
func (s *tradeService) UpdateTrade(id uint, trade models.TradeOrder) (models.TradeOrder, error) {
	var existingTrade models.TradeOrder
	if err := s.db.First(&existingTrade, id).Error; err != nil {
		return models.TradeOrder{}, err
	}

	// Update only allowed fields
	existingTrade.Status = trade.Status
	existingTrade.Unit = trade.Unit
	existingTrade.Zag = trade.Zag
	result := s.db.Save(&existingTrade)
	return existingTrade, result.Error
}
