package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/danizion/rise/internal/dtos"
	"github.com/danizion/rise/internal/models"
	"github.com/danizion/rise/internal/repository"
	"github.com/danizion/rise/internal/storage/redis"
)

// ContactService handles business logic for contacts
type ContactService struct {
	repo  *repository.Repository
	redis *redis.Redis
}

// NewContactService creates a new instance of ContactService
func NewContactService(db *sql.DB, redisClient *redis.Redis) *ContactService {

	return &ContactService{
		repo:  repository.NewRepository(db),
		redis: redisClient,
	}
}

func (s *UserService) GetContact(userID int) (*dtos.CreateUserRequestDto, error) {
	// Use repository to get user
	repoUser, err := s.repo.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Map repository models to DTO
	user := &dtos.CreateUserRequestDto{
		Username: repoUser.Username,
		Email:    repoUser.Email,
		Password: repoUser.HashedPassword,
	}

	return user, nil
}

// CreateContact creates a new contact
func (s *ContactService) CreateContact(contact dtos.CreateContactRequestDto) (int, error) {
	// Map DTO to repository model
	repoContact := models.Contact{
		UserID:      contact.UserID,
		FirstName:   contact.FirstName,
		LastName:    contact.LastName,
		PhoneNumber: contact.PhoneNumber,
		Address:     contact.Address,
	}

	// Use repository to create contact
	contactID, err := s.repo.CreateContact(repoContact)
	if err != nil {
		return 0, fmt.Errorf("failed to create contact: %w", err)
	}

	// Invalidate cache for this user if Redis is available
	if s.redis != nil {
		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(contact.UserID)

		// Invalidate asynchronously to not block the response
		go func() {
			// Ignore errors in background goroutine
			_ = s.redis.InvalidateUserCache(userIDStr)
		}()
	}

	return contactID, nil
}

// GetContactsPaginated retrieves contacts for a user with pagination
func (s *ContactService) GetContactsPaginated(userID, page, pageSize int, firstName, lastName, phoneNumber string) (*dtos.PaginationResult, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	if s.redis != nil {
		// Create filter map
		filters := map[string]string{
			"first_name":   firstName,
			"last_name":    lastName,
			"phone_number": phoneNumber,
		}

		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(userID)

		// Try to get pagination result from cache
		var cachedResult dtos.PaginationResult
		found, err := s.redis.GetCachedPaginationResult(userIDStr, filters, page, pageSize, &cachedResult)
		if err == nil && found {
			// Cache hit - return the pagination result directly
			return &cachedResult, nil
		}
	}

	// Cache miss or Redis not available, get from database
	repoContacts, total, err := s.repo.GetContactsByUserPaginated(userID, page, pageSize, firstName, lastName, phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated contacts: %w", err)
	}

	// Map repository models to DTOs
	contacts := make([]dtos.GetContactsResponseDto, len(repoContacts))
	for i, repoContact := range repoContacts {
		contacts[i] = dtos.GetContactsResponseDto{
			ID:          repoContact.ID,
			UserID:      repoContact.UserID,
			FirstName:   repoContact.FirstName,
			LastName:    repoContact.LastName,
			PhoneNumber: repoContact.PhoneNumber,
			Address:     repoContact.Address,
		}
	}

	// Calculate total pages
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	// Prepare result
	result := &dtos.PaginationResult{
		Items:      contacts,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	// Cache the result if Redis is available
	if s.redis != nil {
		// Create filter map
		filters := map[string]string{
			"first_name":   firstName,
			"last_name":    lastName,
			"phone_number": phoneNumber,
		}

		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(userID)

		// Cache the pagination result
		_ = s.redis.CachePaginationResult(userIDStr, filters, page, pageSize, result)
	}

	return result, nil
}

// UpdateContact updates an existing contact
func (s *ContactService) UpdateContact(updateContactRequestDto dtos.UpdateContactRequestDto) error {
	// Map DTO to repository model
	repoContact := models.Contact{
		ID:          updateContactRequestDto.ID,
		UserID:      updateContactRequestDto.UserID,
		FirstName:   updateContactRequestDto.FirstName,
		LastName:    updateContactRequestDto.LastName,
		PhoneNumber: updateContactRequestDto.PhoneNumber,
		Address:     updateContactRequestDto.Address,
	}

	// Only update fields that are not empty
	updateFields := make(map[string]bool)

	if updateContactRequestDto.FirstName != "" {
		updateFields["first_name"] = true
	}

	if updateContactRequestDto.LastName != "" {
		updateFields["last_name"] = true
	}

	if updateContactRequestDto.PhoneNumber != "" {
		updateFields["phone_number"] = true
	}

	if updateContactRequestDto.Address != "" {
		updateFields["address"] = true
	}

	// Use repository to update updateContactRequestDto
	err := s.repo.UpdateContact(repoContact, updateFields)
	if err != nil {
		return fmt.Errorf("failed to update updateContactRequestDto: %w", err)
	}

	// Invalidate cache for this user if Redis is available
	if s.redis != nil {
		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(updateContactRequestDto.UserID)

		// Invalidate cache for the given user
		_ = s.redis.InvalidateUserCache(userIDStr)
	}

	return nil
}

// DeleteContact deletes a contact by ID and user ID
func (s *ContactService) DeleteContact(userID, contactID int) error {
	// Invalidate cache for this user if Redis is available
	if s.redis != nil {
		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(userID)

		// Invalidate cache for the given user
		_ = s.redis.InvalidateUserCache(userIDStr)

	}

	// Use repository to delete contact
	err := s.repo.DeleteContact(contactID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}
