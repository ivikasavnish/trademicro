package models

import (
	"time"
)

// TradeOrder represents a trade order
type TradeOrder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `json:"symbol"`
	Unit      int       `json:"unit"`
	Diff      float64   `json:"diff"`
	Zag       int       `json:"zag"`
	Type      string    `json:"type"`
	User      string    `json:"user"`
	Status    string    `json:"status" gorm:"default:running"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BrokerToken represents a broker API token

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// StartTradeProcessRequest is the request body for starting a trade process
type StartTradeProcessRequest struct {
	Script string   `json:"script"`
	Args   []string `json:"args"`
}
