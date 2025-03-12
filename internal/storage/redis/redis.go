package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/danizion/rise/internal/models"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	client *redis.Client
}

func InitRedis() *Redis {
	host := getEnvOrDefault("REDIS_HOST", "localhost")
	port := getEnvOrDefault("REDIS_PORT", "6379")
	password := getEnvOrDefault("REDIS_PASSWORD", "")

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}
	return &Redis{
		client: client,
	}
}

func buildCacheKey(userID string, filters map[string]string, page, limit int) string {
	key := fmt.Sprintf("contacts:user:%s", userID)
	for k, v := range filters {
		if v != "" {
			key += fmt.Sprintf(":%s=%s", k, v)
		}
	}
	key += fmt.Sprintf(":page:%d:limit:%d", page, limit)
	return key
}

func (r *Redis) CacheContacts(userID string, filters map[string]string, page, limit int, contacts []models.Contact) error {
	cacheKey := buildCacheKey(userID, filters, page, limit)
	contactsJSON, err := json.Marshal(contacts)
	if err != nil {
		return err
	}
	// Set the cache with a TTL of 5 minutes.
	return r.client.Set(context.Background(), cacheKey, contactsJSON, 5*time.Minute).Err()
}

func (r *Redis) GetCachedContacts(userID string, filters map[string]string, page, limit int) ([]models.Contact, error) {
	cacheKey := buildCacheKey(userID, filters, page, limit)
	contactsJSON, err := r.client.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		// Cache miss.
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var contacts []models.Contact
	if err := json.Unmarshal([]byte(contactsJSON), &contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}

// CachePaginationResult caches the entire pagination result
func (r *Redis) CachePaginationResult(userID string, filters map[string]string, page, limit int, result interface{}) error {
	cacheKey := buildCacheKey(userID, filters, page, limit)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	// Set the cache with a TTL of 5 minutes.
	return r.client.Set(context.Background(), cacheKey, resultJSON, 5*time.Minute).Err()
}

// GetCachedPaginationResult retrieves the entire pagination result from cache
// Returns (found, error) where found indicates if the key was found in cache
func (r *Redis) GetCachedPaginationResult(userID string, filters map[string]string, page, limit int, result interface{}) (bool, error) {
	cacheKey := buildCacheKey(userID, filters, page, limit)
	resultJSON, err := r.client.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		// Cache miss.
		return false, nil
	} else if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(resultJSON), result); err != nil {
		return false, err
	}
	return true, nil
}

// InvalidateUserCache removes all cached contact entries for a specific user
func (r *Redis) InvalidateUserCache(userID string) error {
	// Create pattern to match all keys for this user
	pattern := fmt.Sprintf("contacts:user:%s:*", userID)

	// Use SCAN to find all keys matching the pattern
	ctx := context.Background()
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	// Delete each matching key
	for iter.Next(ctx) {
		key := iter.Val()
		err := r.client.Del(ctx, key).Err()
		if err != nil {
			log.Printf("Error deleting key %s: %v", key, err)
			// Continue deleting other keys even if one fails
		}
	}

	// Check for errors during iteration
	if err := iter.Err(); err != nil {
		log.Printf("Error scanning Redis keys: %v", err)
		return err
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
