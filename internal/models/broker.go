package models

import (
	"time"
)

// Broker represents a trading broker that can be integrated with the platform
type Broker struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name" gorm:"unique"`
	Code        string    `json:"code" gorm:"unique"`
	Description string    `json:"description"`
	APIBaseURL  string    `json:"apiBaseUrl" gorm:"column:api_base_url"`
	IsActive    bool      `json:"isActive" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName specifies the table name for Broker model
func (Broker) TableName() string {
	return "brokers"
}

// BrokerRequest is used for creating and updating brokers
type BrokerRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	APIBaseURL  string `json:"apiBaseUrl"`
	IsActive    *bool  `json:"isActive,omitempty"`
}

// BrokerToken represents an authentication token for a broker
type BrokerToken struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	UserID         uint          `json:"userId" gorm:"column:user_id"`
	FamilyMemberID *uint         `json:"familyMemberId,omitempty" gorm:"column:family_member_id"`
	BrokerID       uint          `json:"brokerId" gorm:"column:broker_id"`
	ClientID       string        `json:"clientId" gorm:"column:client_id"`
	AccessToken    string        `json:"accessToken,omitempty" gorm:"column:access_token"`
	RefreshToken   string        `json:"refreshToken,omitempty" gorm:"column:refresh_token"`
	ExpiresAt      time.Time     `json:"expiresAt" gorm:"column:expires_at"`
	IsActive       bool          `json:"isActive" gorm:"column:is_active;default:true"`
	CreatedAt      time.Time     `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time     `json:"updatedAt" gorm:"column:updated_at"`
	User           *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	FamilyMember   *FamilyMember `json:"familyMember,omitempty" gorm:"foreignKey:FamilyMemberID"`
	Broker         *Broker       `json:"broker,omitempty" gorm:"foreignKey:BrokerID"`
}

// TableName specifies the table name for BrokerToken model
func (BrokerToken) TableName() string {
	return "broker_tokens"
}

// BrokerTokenRequest is used for creating and updating broker tokens
type BrokerTokenRequest struct {
	BrokerID       uint      `json:"brokerId" binding:"required"`
	FamilyMemberID *uint     `json:"familyMemberId,omitempty"`
	ClientID       string    `json:"clientId" binding:"required"`
	AccessToken    string    `json:"accessToken" binding:"required"`
	RefreshToken   string    `json:"refreshToken,omitempty"`
	ExpiresAt      time.Time `json:"expiresAt,omitempty"`
}
