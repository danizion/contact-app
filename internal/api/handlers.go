package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/danizion/rise/internal/constants"
	"github.com/danizion/rise/internal/dtos"
	"github.com/danizion/rise/internal/service"
	"github.com/danizion/rise/internal/storage/redis"
	"github.com/gin-gonic/gin"
)

// Handler for contact and users routes holds contact and user services to apply all logic
type Handler struct {
	contactService *service.ContactService
	userService    *service.UserService
}

func NewHandler(db *sql.DB, redisClient *redis.Redis) *Handler {
	return &Handler{
		contactService: service.NewContactService(db, redisClient),
		userService:    service.NewUserService(db),
	}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dtos.CreateUserRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid create user request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, err := h.userService.CreateUser(req)
	if err != nil {
		if strings.Contains(err.Error(), constants.ErrUsernameExists) {
			slog.Error("Failed to create user", "error", "username already exists", "username", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": constants.ErrUsernameExists})
			return
		}
		if strings.Contains(err.Error(), constants.ErrEmailExists) {
			slog.Error("Failed to create user", "error", "email already exists", "email", req.Email)
			c.JSON(http.StatusConflict, gin.H{"error": constants.ErrEmailExists})
			return
		}
		slog.Error("Failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	slog.Info("User created successfully", "userID", userID)
	// Return success response with the new user ID
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"userID":  userID,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req dtos.LoginRequestDto

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid login request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("Login attempt", "email", req.Email)

	// Authenticate user
	user, err := h.userService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		slog.Error("Login failed", "error", err, "email", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate and sign token
	token, err := h.userService.GenerateToken(user.ID, user.Username)
	if err != nil {
		slog.Error("Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	slog.Info("Login successful", "userID", user.ID, "email", req.Email)

	// Return the JWT token
	c.JSON(http.StatusOK, dtos.LoginResponseDto{
		Token:  token,
		UserID: user.ID,
	})
}

func (h *Handler) GetContacts(c *gin.Context) {
	var req dtos.GetContactRequestDto
	if err := c.ShouldBindQuery(&req); err != nil {
		slog.Error("Invalid get contacts request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = h.getUserID(c)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	req.Page = page

	// Get filter parameters and populate the DTO
	req.FirstName = c.Query("first_name")
	req.LastName = c.Query("last_name")
	req.PhoneNumber = c.Query("phone_number")
	req.Address = c.Query("address")

	req.PageSize = constants.DefaultPageSize

	slog.Info("Getting contacts", "userID", req.UserID, "page", req.Page, "pageSize", req.PageSize)

	// Get paginated contacts from service
	result, err := h.contactService.GetContacts(req)
	if err != nil {
		slog.Error("Failed to retrieve contacts", "error", err, "userID", req.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contacts"})
		return
	}

	slog.Info("Retrieved contacts", "count", len(result.Items), "total", result.TotalCount, "userID", req.UserID)

	// Return paginated results
	c.JSON(http.StatusOK, result)
}

// CreateContact handles POST requests for creating a new contact
func (h *Handler) CreateContact(c *gin.Context) {
	// Parse request body
	var req dtos.CreateContactRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid create contact request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = h.getUserID(c)

	slog.Info("Creating new contact", "userID", req.UserID)

	// Call service to create contact
	contactID, err := h.contactService.CreateContact(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			slog.Error("Contact creation failed", "error", err, "userID", req.UserID)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		slog.Error("Failed to create contact", "error", err, "userID", req.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	slog.Info("Contact created successfully", "contactID", contactID, "userID", req.UserID)

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Contact created successfully",
		"contact_id": contactID,
	})
}

func (h *Handler) UpdateContact(c *gin.Context) {
	// Get contact ID from URL parameter
	contactID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("Invalid contact ID", "id", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	var req dtos.UpdateContactRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid update contact request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = h.getUserID(c)
	req.ID = contactID

	slog.Info("Updating contact", "contactID", contactID, "userID", req.UserID)

	// Call service to update contact
	err = h.contactService.UpdateContact(req)
	if err != nil {
		slog.Error("Failed to update contact", "error", err, "contactID", contactID)
		if strings.Contains(err.Error(), "contact not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	slog.Info("Contact updated successfully", "contactID", contactID, "userID", req.UserID)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Contact updated successfully",
	})
}

func (h *Handler) DeleteContact(c *gin.Context) {
	// Get contact ID from URL parameter
	contactID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("Invalid contact ID", "id", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	userID := h.getUserID(c)

	slog.Info("Deleting contact", "contactID", contactID, "userID", userID)

	// Call service to delete contact
	err = h.contactService.DeleteContact(userID, contactID)
	if err != nil {
		slog.Error("Failed to delete contact", "error", err, "contactID", contactID)
		if strings.Contains(err.Error(), "contact not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete contact with error: %v", err)})
		return
	}

	slog.Info("Contact deleted successfully", "contactID", contactID, "userID", userID)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Contact deleted successfully",
	})
}

func (h *Handler) getUserID(c *gin.Context) int {
	userID, exists := c.Get("userID")
	if !exists {
		slog.Error("Failed to retrieve user ID from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user please try login again"})
	}
	id, ok := userID.(int)
	if !ok {
		slog.Error("Invalid user ID type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type please try login again"})
	}
	return id
}
