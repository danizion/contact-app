package dtos

//type CreateContactDto struct {
//	UserID      int    `json:"user_id"`
//	FirstName   string `json:"first_name"`
//	LastName    string `json:"last_name"`
//	PhoneNumber string `json:"phone_number"`
//	Address     string `json:"address,omitempty"`
//}

// GetContactsResponseDto represents a contact for API responses
type GetContactsResponseDto struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address,omitempty"`
}

// UpdateContactRequestDto represents the data for updating a contact
type UpdateContactRequestDto struct {
	ID          int    `json:"contact_id"  binding:"required"`
	UserID      int    `json:"user_id"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Address     string `json:"address,omitempty"`
}

// Define request structure with user ID in body
type GetContactRequestDto struct {
	UserID int `json:"user_id" `
}

// Define request structure for creating a contact
type CreateContactRequestDto struct {
	UserID      int    `json:"user_id"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Address     string `json:"address" binding:"omitempty"`
}

type DeleteContactRequestDto struct {
	UserID    int `json:"user_id" `
	ContactID int `json:"contact_id" binding:"required"`
}

// PaginationResult represents a paginated response
type PaginationResult struct {
	Items      []GetContactsResponseDto `json:"items"`
	TotalCount int                      `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

type CreateUserRequestDto struct {
	Username string `json:"user_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequestDto struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponseDto struct {
	Token  string `json:"token"`
	UserID int    `json:"user_id"`
}
