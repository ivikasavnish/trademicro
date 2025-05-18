package services

import (
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/models"
)

// FamilyMemberService defines the interface for family member operations
type FamilyMemberService interface {
	GetFamilyMembersByUserID(userID uint) ([]models.FamilyMember, error)
	CreateFamilyMember(member models.FamilyMember) (models.FamilyMember, error)
	UpdateFamilyMember(id uint, userID uint, member models.FamilyMember) (models.FamilyMember, error)
	DeleteFamilyMember(id uint, userID uint) error
	GetFamilyMemberByID(id uint) (models.FamilyMember, error)
}

// familyMemberService implements the FamilyMemberService interface
type familyMemberService struct {
	db *gorm.DB
}

// NewFamilyMemberService creates a new family member service
func NewFamilyMemberService(db *gorm.DB) FamilyMemberService {
	return &familyMemberService{
		db: db,
	}
}

// GetFamilyMembersByUserID returns all family members for a user
func (s *familyMemberService) GetFamilyMembersByUserID(userID uint) ([]models.FamilyMember, error) {
	var members []models.FamilyMember
	result := s.db.Where("user_id = ?", userID).Find(&members)
	return members, result.Error
}

// CreateFamilyMember creates a new family member
func (s *familyMemberService) CreateFamilyMember(member models.FamilyMember) (models.FamilyMember, error) {
	result := s.db.Create(&member)
	return member, result.Error
}

// UpdateFamilyMember updates a family member
func (s *familyMemberService) UpdateFamilyMember(id uint, userID uint, member models.FamilyMember) (models.FamilyMember, error) {
	var existingMember models.FamilyMember
	if err := s.db.First(&existingMember, id).Error; err != nil {
		return models.FamilyMember{}, err
	}

	// Verify ownership
	if existingMember.UserID != userID {
		return models.FamilyMember{}, gorm.ErrRecordNotFound
	}

	// Update allowed fields
	existingMember.Name = member.Name
	existingMember.Email = member.Email
	existingMember.Phone = member.Phone
	existingMember.Pin = member.Pin
	existingMember.PortfolioID = member.PortfolioID
	existingMember.IsActive = member.IsActive
	existingMember.UpdatedAt = member.UpdatedAt

	result := s.db.Save(&existingMember)
	return existingMember, result.Error
}

// DeleteFamilyMember deletes a family member
func (s *familyMemberService) DeleteFamilyMember(id uint, userID uint) error {
	var member models.FamilyMember
	if err := s.db.First(&member, id).Error; err != nil {
		return err
	}

	// Verify ownership
	if member.UserID != userID {
		return gorm.ErrRecordNotFound
	}

	return s.db.Delete(&models.FamilyMember{}, id).Error
}

// GetFamilyMemberByID returns a family member by ID
func (s *familyMemberService) GetFamilyMemberByID(id uint) (models.FamilyMember, error) {
	var member models.FamilyMember
	result := s.db.First(&member, id)
	return member, result.Error
}
