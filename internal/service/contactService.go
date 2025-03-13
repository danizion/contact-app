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

// ContactService handles business logic for contacts has a pointer for repository for db interaction and redis for cache interaction
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

func (s *ContactService) CreateContact(contact dtos.CreateContactRequestDto) (int, error) {
	// Check if contact with same name exists
	exists, err := s.repo.IsContactExists(contact.UserID, contact.FirstName, contact.LastName)
	if err != nil {
		return 0, fmt.Errorf("failed to check existing contact: %w", err)
	}
	if exists {
		return 0, fmt.Errorf("contact with name %s %s already exists. Please use update to change the number or use a different name",
			contact.FirstName, contact.LastName)
	}

	// Map DTO to model
	repoContact := models.Contact{
		UserID:      contact.UserID,
		FirstName:   contact.FirstName,
		LastName:    contact.LastName,
		PhoneNumber: contact.PhoneNumber,
		Address:     contact.Address,
	}

	contactID, err := s.repo.CreateContact(repoContact)
	if err != nil {
		return 0, fmt.Errorf("failed to create contact: %w", err)
	}

	// Invalidate cache for this user if Redis is available
	if s.redis != nil {
		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(contact.UserID)

		err := s.redis.InvalidateUserCache(userIDStr)
		if err != nil {
			return 0, err
		}
	}

	return contactID, nil
}

// GetContacts retrieves contacts for a user with pagination
func (s *ContactService) GetContacts(req dtos.GetContactRequestDto) (*dtos.PaginationResult, error) {
	// Validate pagination parameters

	if s.redis != nil {
		// Create filter map
		filters := map[string]string{
			"first_name":   req.FirstName,
			"last_name":    req.LastName,
			"phone_number": req.PhoneNumber,
			"address":      req.Address,
		}

		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(req.UserID)

		// Try to get pagination result from cache
		var cachedResult dtos.PaginationResult
		found, err := s.redis.GetCachedPaginationResult(userIDStr, filters, req.Page, req.PageSize, &cachedResult)
		if err == nil && found {
			// Cache hit - return the pagination result directly
			return &cachedResult, nil
		}
	}

	// Cache miss or Redis not available, get from database
	repoContacts, total, err := s.repo.GetContactsByUserPaginated(req.UserID, req.Page, req.PageSize, req.FirstName, req.LastName, req.PhoneNumber, req.Address)
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
	totalPages := total / req.PageSize
	if total%req.PageSize > 0 {
		totalPages++
	}

	// Prepare result
	result := &dtos.PaginationResult{
		Items:      contacts,
		TotalCount: total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	// Cache the result if Redis is available
	if s.redis != nil {
		// Create filter map
		filters := map[string]string{
			"first_name":   req.FirstName,
			"last_name":    req.LastName,
			"phone_number": req.PhoneNumber,
		}

		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(req.UserID)

		// Cache the pagination result
		err := s.redis.CachePaginationResult(userIDStr, filters, req.Page, req.PageSize, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// UpdateContact updates an existing contact, only update none empty fields
func (s *ContactService) UpdateContact(updateContactRequestDto dtos.UpdateContactRequestDto) error {
	// Map DTO to model
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

	err := s.repo.UpdateContact(repoContact, updateFields)
	if err != nil {
		return err
	}

	// Invalidate cache for this user if Redis is available
	if s.redis != nil {
		// Convert userID to string for cache key
		userIDStr := strconv.Itoa(updateContactRequestDto.UserID)

		// Invalidate cache for the given user
		err := s.redis.InvalidateUserCache(userIDStr)
		if err != nil {
			return err
		}
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
		err := s.redis.InvalidateUserCache(userIDStr)
		if err != nil {
			return err
		}

	}

	err := s.repo.DeleteContact(contactID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}
