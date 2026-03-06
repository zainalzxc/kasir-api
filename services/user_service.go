package services

import (
	"kasir-api/models"
	"kasir-api/repositories"
	"kasir-api/utils"
)

// UserService handles user management logic
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAll()
}

// CreateUser creates a new user
func (s *UserService) CreateUser(username, password, role string) (*models.User, error) {
	if role != models.RoleAdmin && role != models.RoleKasir {
		return nil, models.ErrInvalidRole
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:    username,
		Password:    hashedPassword,
		NamaLengkap: username, // Default to username if not provided
		Role:        role,
		IsActive:    true,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdatePassword updates a user's password (by admin)
func (s *UserService) UpdatePassword(id int, newPassword string) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return err // Ensure user exists
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(id, hashedPassword)
}

// DeleteUser deletes a user (except themselves)
func (s *UserService) DeleteUser(targetID, currentAdminID int) error {
	if targetID == currentAdminID {
		return models.ErrCannotDeleteSelf
	}

	_, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return err // Check if user exists before deleting
	}

	return s.userRepo.Delete(targetID)
}
