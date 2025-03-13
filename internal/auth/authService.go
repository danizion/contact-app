package auth

import (
	"github.com/danizion/rise/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Secret key used to sign JWT tokens - in production this should be stored securely
var jwtSecretKey = []byte(utils.GetEnvOrDefault("AUTH_SECRET", "im-a-secret-key"))

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// GetJWTSecret returns the secret key used for JWT signing and verification
func GetJWTSecret() []byte {
	return jwtSecretKey
}

func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost is generally 10; you can adjust based on your security/performance tradeoffs.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword verifies that the provided plain text password matches the stored hashed password.
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateJWT creates a new JWT token for the authenticated user
func GenerateJWT(userID int, username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims using HS256 signing method.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign token using the secret key.
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
