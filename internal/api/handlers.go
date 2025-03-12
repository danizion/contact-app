package api

import (
	"database/sql"
	"fmt"

	"github.com/danizion/rise/internal/dtos"
	"github.com/danizion/rise/internal/service"
	"github.com/danizion/rise/internal/storage/redis"
	"github.com/gin-gonic/gin"

	"net/http"
	"strconv"
)

// Handler struct holds services required by handler functions
type Handler struct {
	contactService *service.ContactService
	userService    *service.UserService
}

// NewHandler creates a new handler with required services
func NewHandler(db *sql.DB, redisClient *redis.Redis) *Handler {
	return &Handler{
		contactService: service.NewContactService(db, redisClient),
		userService:    service.NewUserService(db),
	}
}

// CreateUser handles user creation requests
func (h *Handler) CreateUser(c *gin.Context) {
	var req dtos.CreateUserRequestDto

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.userService.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Return success response with the new user ID
	c.JSON(http.StatusCreated, gin.H{
		"message": "CreateUserRequestDto created successfully",
		"user_id": userID,
	})
}

// Login handles user authentication requests
func (h *Handler) Login(c *gin.Context) {
	var req dtos.LoginRequestDto

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user with the service
	user, err := h.userService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.userService.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return token response
	c.JSON(http.StatusOK, dtos.LoginResponseDto{
		Token:  token,
		UserID: user.ID,
	})
}

// GetContacts handles GET requests for retrieving contacts
func (h *Handler) GetContacts(c *gin.Context) {
	// Parse request body for user ID
	var req dtos.GetContactRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = h.getUserID(c)
	// Extract query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// Additional query parameters (optional)
	firstName := c.Query("first_name")
	lastName := c.Query("last_name")
	phoneNumber := c.Query("phone_number")
	userID := h.getUserID(c)
	// Always use page size of 10 as specified
	pageSize := 10

	//TODO: send the dto to service not the divided params
	// Get paginated contacts from service
	result, err := h.contactService.GetContactsPaginated(userID, page, pageSize, firstName, lastName, phoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contacts"})
		return
	}

	// Return paginated results
	c.JSON(http.StatusOK, result)
}

// CreateContact handles POST requests for creating a new contact
func (h *Handler) CreateContact(c *gin.Context) {

	// Parse request body
	var req dtos.CreateContactRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = h.getUserID(c)
	//// Create contact dto
	//contactDto := dtos.CreateContactDto{
	//	FirstName:   req.FirstName,
	//	LastName:    req.LastName,
	//	PhoneNumber: req.PhoneNumber,
	//	Address:     req.Address,
	//}

	// Call service to create contact
	contactID, err := h.contactService.CreateContact(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Contact created successfully",
		"contact_id": contactID,
	})
}

// UpdateContact handles PATCH requests for updating a contact
func (h *Handler) UpdateContact(c *gin.Context) {
	// Define request structure for updating a contact with pointers to allow null values

	// Parse request body
	var req dtos.UpdateContactRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = h.getUserID(c)

	// Call service to update contact
	err := h.contactService.UpdateContact(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Contact updated successfully",
	})
}

// DeleteContact handles DELETE requests for deleting a contact
func (h *Handler) DeleteContact(c *gin.Context) {
	// Define request structure for deleting a contact

	// Parse request body
	var req dtos.DeleteContactRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = h.getUserID(c)
	// Call service to delete contact
	err := h.contactService.DeleteContact(req.UserID, req.ContactID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete contact with error: %v", err)})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Contact deleted successfully",
	})
}

func (h *Handler) getUserID(c *gin.Context) int {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user please try login again"})
	}
	id, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type please try login again"})
	}
	return id
}
