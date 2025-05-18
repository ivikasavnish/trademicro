package models

import (
	"time"
)

// FamilyMember represents a family member linked to a main user account
type FamilyMember struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Pin         string    `json:"pin,omitempty" gorm:"column:pin"`
	PortfolioID int64     `json:"portfolioId" gorm:"column:portfolio_id"`
	IsActive    bool      `json:"isActive" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
	ClientToken string    `json:"clientToken,omitempty" gorm:"column:client_token"`
	ClientID    string    `json:"clientId,omitempty" gorm:"column:client_id"`
	UserID      uint      `json:"userId" gorm:"column:user_id"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for FamilyMember model
func (FamilyMember) TableName() string {
	return "family_members"
}

// FamilyMemberRequest is used for creating and updating family members
type FamilyMemberRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Pin         string `json:"pin,omitempty"`
	PortfolioID int64  `json:"portfolioId"`
	IsActive    *bool  `json:"isActive,omitempty"`
	ClientToken string `json:"clientToken,omitempty"`
	ClientID    string `json:"clientId,omitempty"`
}
