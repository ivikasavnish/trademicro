package models

import (
	"time"
)

type Symbol struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ExchangeID     string    `json:"exchangeId" gorm:"index"`
	Segment        string    `json:"segment" gorm:"index"`
	Symbol         string    `json:"symbol" gorm:"index"`
	Name           string    `json:"name"`
	ISIN           string    `json:"isin"`
	Instrument     string    `json:"instrument"`
	SecurityID     string    `json:"securityId" gorm:"index"`
	TradingSymbol  string    `json:"tradingSymbol" gorm:"index"`
	DisplayName    string    `json:"displayName"`
	InstrumentType string    `json:"instrumentType"`
	Series         string    `json:"series"`
	LotSize        int       `json:"lotSize"`
	TickSize       float64   `json:"tickSize"`
	ExpiryDate     string    `json:"expiryDate"`
	StrikePrice    float64   `json:"strikePrice"`
	OptionType     string    `json:"optionType"`
	ExpiryFlag     string    `json:"expiryFlag"`
	IsActive       bool      `json:"isActive" gorm:"default:true"`
	LastUpdated    time.Time `json:"lastUpdated"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
