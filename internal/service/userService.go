package service

import (
	"database/sql"
	"fmt"
	"github.com/danizion/contact-app/internal/auth"
	"github.com/danizion/contact-app/internal/constants"
	"github.com/danizion/contact-app/internal/dtos"
	"github.com/danizion/contact-app/internal/models"
	"github.com/danizion/contact-app/internal/repository"
	"log"
)

// UserService handles business logic for users
type UserService struct {
	repo *repository.Repository
}

// NewUserService creates a new instance of UserService
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		repo: repository.NewRepository(db),
	}
}

// CreateUserRequestDto is the DTO (Data Transfer Object) for user operations

// CreateUser creates a new user
func (s *UserService) CreateUser(createUserRequestDto dtos.CreateUserRequestDto) (int, error) {
	// Check if username already exists
	existingUser, err := s.repo.GetUserByUsername(createUserRequestDto.Username)
	if err != nil {
		log.Printf("Error checking username: %v", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	if existingUser != nil {
		return 0, fmt.Errorf(constants.ErrUsernameExists)
	}

	// Check if email already exists
	existingUser, err = s.repo.GetUserByEmail(createUserRequestDto.Email)
	if err != nil {
		log.Printf("Error checking email: %v", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	if existingUser != nil {
		return 0, fmt.Errorf(constants.ErrEmailExists)
	}

	// Map DTO to repository models

	hashedPassword, err := auth.HashPassword(createUserRequestDto.Password)
	if nil != err {
		log.Printf("Failed to hash password: %v", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	repoUser := models.User{
		Username:       createUserRequestDto.Username,
		Email:          createUserRequestDto.Email,
		HashedPassword: hashedPassword,
	}

	// Use repository to create user
	userID, err := s.repo.CreateUser(repoUser)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// AuthenticateUser validates user credentials and returns user data if valid
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	// Get user by email from repository
	user, err := s.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		log.Printf("Failed to find user with email %s: %v", email, err)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if !auth.CheckPassword(password, user.HashedPassword) {
		log.Printf("Invalid password for user with email %s", email)
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// GenerateToken creates a JWT token for the authenticated user
func (s *UserService) GenerateToken(userID int, username string) (string, error) {
	// Use the auth package to generate a JWT
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return "", fmt.Errorf("failed to generate authentication token: %w", err)
	}

	return token, nil
}
