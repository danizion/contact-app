package models

import "time"

type Contact struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	PhoneNumber string    `db:"phone_number"`
	Address     string    `db:"address"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
