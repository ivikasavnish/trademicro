package models

import (
	"time"
)

type FavouredSymbol struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	SymbolID  uint      `gorm:"index" json:"symbolId"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relations (for eager loading)
	Symbol Symbol `gorm:"foreignKey:SymbolID" json:"symbol,omitempty"`
}
