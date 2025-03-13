package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/danizion/rise/internal/models"
	"github.com/jmoiron/sqlx"
)

// Repository defines the structure of the repository for database interaction
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new instance of the Repository
func NewRepository(db *sql.DB) *Repository {
	sqlxDB := sqlx.NewDb(db, "postgres")
	return &Repository{db: sqlxDB}
}

// CreateUser inserts a new user into the "users" table
func (r *Repository) CreateUser(user models.User) (int, error) {
	query := `INSERT INTO users (username, email, hashed_password) 
			  VALUES ($1, $2, $3) RETURNING id`
	var userID int
	err := r.db.QueryRow(query, user.Username, user.Email, user.HashedPassword).Scan(&userID)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return 0, err
	}
	return userID, nil
}

// GetUser retrieves a user by ID from the "users" table
func (r *Repository) GetUser(userID int) (*models.User, error) {
	query := `SELECT id, username, email, hashed_password, created_at, updated_at 
			  FROM users WHERE id = $1`
	var user models.User
	err := r.db.Get(&user, query, userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email from the "users" table
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, hashed_password, created_at, updated_at 
			  FROM users WHERE email = $1`
	var user models.User
	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by email: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username from the "users" table
func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, hashed_password, created_at, updated_at 
			  FROM users WHERE username = $1`
	var user models.User
	err := r.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by username: %v", err)
		return nil, err
	}
	return &user, nil
}

// CreateContact inserts a new contact into the "contacts" table
func (r *Repository) CreateContact(contact models.Contact) (int, error) {
	query := `INSERT INTO contacts (user_id, first_name, last_name, phone_number, address) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var contactID int
	err := r.db.QueryRow(query, contact.UserID, contact.FirstName, contact.LastName, contact.PhoneNumber, contact.Address).Scan(&contactID)
	if err != nil {
		log.Printf("Error creating contact: %v", err)
		return 0, err
	}
	return contactID, nil
}

// GetContactsByUser retrieves all contacts for a specific user
func (r *Repository) GetContactsByUser(userID int) ([]models.Contact, error) {
	query := `SELECT id, user_id, first_name, last_name, phone_number, address, created_at, updated_at 
			  FROM contacts WHERE user_id = $1`
	var contacts []models.Contact
	err := r.db.Select(&contacts, query, userID)
	if err != nil {
		log.Printf("Error fetching contacts: %v", err)
		return nil, err
	}
	return contacts, nil
}

// GetContactsByUserPaginated retrieves contacts for a user with pagination
func (r *Repository) GetContactsByUserPaginated(userID int, page, pageSize int, firstName, lastName, phoneNumber string, address string) ([]models.Contact, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Initialize parameters
	params := []interface{}{userID}
	paramIndex := 1

	// Build the base query with conditional filters
	baseQuery := `FROM contacts WHERE user_id = $1`

	// Add optional filters if provided
	if firstName != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND first_name ILIKE $%d", paramIndex)
		params = append(params, "%"+firstName+"%")
	}

	if lastName != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND last_name ILIKE $%d", paramIndex)
		params = append(params, "%"+lastName+"%")
	}

	if phoneNumber != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND phone_number ILIKE $%d", paramIndex)
		params = append(params, "%"+phoneNumber+"%")
	}

	if address != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND address ILIKE $%d", paramIndex)
		params = append(params, "%"+address+"%")
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.Get(&total, countQuery, params...)
	if err != nil {
		log.Printf("Error counting contacts: %v", err)
		return nil, 0, err
	}

	// Get paginated contacts
	limitOffset := fmt.Sprintf(" ORDER BY id LIMIT %d OFFSET %d", pageSize, offset)
	query := `SELECT id, user_id, first_name, last_name, phone_number, address, created_at, updated_at ` + baseQuery + limitOffset
	var contacts []models.Contact
	err = r.db.Select(&contacts, query, params...)
	if err != nil {
		log.Printf("Error fetching paginated contacts: %v", err)
		return nil, 0, err
	}

	return contacts, total, nil
}

// GetContactsTotalCount retrieves only the total count of contacts matching the criteria
func (r *Repository) GetContactsTotalCount(userID int, firstName, lastName, phoneNumber string) (int, error) {
	// Initialize parameters
	params := []interface{}{userID}
	paramIndex := 1

	// Build the base query with conditional filters
	baseQuery := `FROM contacts WHERE user_id = $1`

	// Add optional filters if provided
	if firstName != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND first_name ILIKE $%d", paramIndex)
		params = append(params, "%"+firstName+"%")
	}

	if lastName != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND last_name ILIKE $%d", paramIndex)
		params = append(params, "%"+lastName+"%")
	}

	if phoneNumber != "" {
		paramIndex++
		baseQuery += fmt.Sprintf(" AND phone_number ILIKE $%d", paramIndex)
		params = append(params, "%"+phoneNumber+"%")
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.Get(&total, countQuery, params...)
	if err != nil {
		log.Printf("Error counting contacts: %v", err)
		return 0, err
	}

	return total, nil
}

// UpdateContact updates an existing contact in the database
func (r *Repository) UpdateContact(contact models.Contact, updateFields map[string]bool) error {
	// First verify the contact exists and belongs to the specified user
	checkQuery := `SELECT COUNT(*) FROM contacts WHERE id = $1 AND user_id = $2`
	var count int
	err := r.db.Get(&count, checkQuery, contact.ID, contact.UserID)
	if err != nil {
		log.Printf("Error checking contact ownership: %v", err)
		return err
	}

	if count == 0 {
		return fmt.Errorf("contact not found or does not belong to the specified user")
	}

	// Build dynamic update query based on provided fields
	query := `UPDATE contacts SET`
	params := []interface{}{}
	paramIndex := 0

	// Only include fields that were explicitly marked for update
	updates := []string{}

	if updateFields["first_name"] {
		paramIndex++
		updates = append(updates, fmt.Sprintf(" first_name = $%d", paramIndex))
		params = append(params, contact.FirstName)
	}

	if updateFields["last_name"] {
		paramIndex++
		updates = append(updates, fmt.Sprintf(" last_name = $%d", paramIndex))
		params = append(params, contact.LastName)
	}

	if updateFields["phone_number"] {
		paramIndex++
		updates = append(updates, fmt.Sprintf(" phone_number = $%d", paramIndex))
		params = append(params, contact.PhoneNumber)
	}

	if updateFields["address"] {
		paramIndex++
		updates = append(updates, fmt.Sprintf(" address = $%d", paramIndex))
		params = append(params, contact.Address)
	}

	// If no fields to update, return early
	if len(updates) == 0 {
		return nil
	}

	// Add updated_at timestamp
	paramIndex++
	updates = append(updates, fmt.Sprintf(" updated_at = $%d", paramIndex))
	params = append(params, "NOW()")

	// Combine updates
	query += strings.Join(updates, ",")

	// Add WHERE clause
	paramIndex++
	query += fmt.Sprintf(" WHERE id = $%d", paramIndex)
	params = append(params, contact.ID)

	paramIndex++
	query += fmt.Sprintf(" AND user_id = $%d", paramIndex)
	params = append(params, contact.UserID)

	// Execute the update
	_, err = r.db.Exec(query, params...)
	if err != nil {
		log.Printf("Error updating contact: %v", err)
		return err
	}

	return nil
}

// DeleteContact deletes a contact by ID and user ID
func (r *Repository) DeleteContact(userID, contactID int) error {
	// First verify the contact exists and belongs to the specified user
	checkQuery := `SELECT COUNT(*) FROM contacts WHERE  user_id = $1 AND id = $2`
	var count int
	err := r.db.Get(&count, checkQuery, contactID, userID)
	if err != nil {
		log.Printf("Error checking contact ownership: %v", err)
		return err
	}

	if count == 0 {
		return fmt.Errorf("contact not found or does not belong to the specified user")
	}

	// Delete the contact
	query := `DELETE FROM contacts WHERE user_id = $1 AND id = $2`
	_, err = r.db.Exec(query, contactID, userID)
	if err != nil {
		log.Printf("Error deleting contact: %v", err)
		return err
	}

	return nil
}

// IsContactExists checks if a contact with the same first and last name exists for a user
func (r *Repository) IsContactExists(userID int, firstName, lastName string) (bool, error) {
	query := `SELECT COUNT(*) FROM contacts WHERE user_id = $1 AND first_name = $2 AND last_name = $3`
	var count int
	err := r.db.Get(&count, query, userID, firstName, lastName)
	if err != nil {
		log.Printf("Error checking existing contact: %v", err)
		return false, err
	}
	return count > 0, nil
}
