package db

import (
	"database/sql"
	"fmt"
	"github.com/danizion/contact-app/internal/utils"
	_ "github.com/lib/pq"
	"log"
)

func Init() *sql.DB {
	host := utils.GetEnvOrDefault("POSTGRES_HOST", "localhost")
	port := utils.GetEnvOrDefault("POSTGRES_PORT", "5433")
	user := utils.GetEnvOrDefault("POSTGRES_USER", "myuser")
	password := utils.GetEnvOrDefault("POSTGRES_PASSWORD", "mypassword")
	dbname := utils.GetEnvOrDefault("POSTGRES_DB", "mydb")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Establish a connection to the database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	err = initializeSchemaFromSQL(db)
	if err != nil {
		log.Fatalf("Error initializing the database schema: %v", err)
	}
	return db
}

func initializeSchemaFromSQL(db *sql.DB) error {
	// Read the contents of the schema.sql file
	const schema = `
	CREATE TABLE IF NOT EXISTS users
(
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(50) NOT NULL UNIQUE,
                       email VARCHAR(100) NOT NULL UNIQUE,
                       hashed_password VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS contacts (
                          id SERIAL PRIMARY KEY,
                          user_id INTEGER NOT NULL,
                          first_name VARCHAR(100) NOT NULL,
                          last_name VARCHAR(100) NOT NULL,
                          phone_number VARCHAR(20) NOT NULL,
                          address TEXT,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                          FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
	`

	// Execute the SQL commands in the schema file
	_, err := db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema script: %w", err)
	}

	return nil
}
