package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"shareseer-mcp/internal/models"
	"time"
	"context"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewService() *Service {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       2, // ShareSeer users database
	})
	
	return &Service{
		redisClient: rdb,
		ctx:         context.Background(),
	}
}

// Generate a new API key
func (s *Service) GenerateAPIKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "sk-shareseer-" + hex.EncodeToString(bytes)
}

// Validate API key and return user
func (s *Service) ValidateAPIKey(apiKey string) (*models.User, error) {
	// Look up email from API key
	apiKeyLookup := "api_key:" + apiKey
	email, err := s.redisClient.Get(s.ctx, apiKeyLookup).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}
	
	// Get user data from ShareSeer
	userKey := "email:" + email
	userMap, err := s.redisClient.HGetAll(s.ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	
	// Determine tier based on Premium status and expiration
	tier := s.GetUserTier(userMap["is_premium"], userMap["exp_date"])
	
	user := &models.User{
		ID:       email, // Use email as ID
		APIKey:   apiKey,
		Tier:     tier,
		Email:    email,
		Created:  time.Now(), // We don't track creation time anymore
		LastUsed: time.Now(),
	}
	
	return user, nil
}

// GetUserTier determines user tier based on ShareSeer subscription status
func (s *Service) GetUserTier(premium, expirationDate string) string {
	if premium != "true" {
		return "free"
	}
	
	if expirationDate == "" {
		return "free"
	}
	
	expiry, err := time.Parse("2006-01-02T15:04:05Z", expirationDate)
	if err != nil || time.Now().After(expiry) {
		return "free" // Expired subscription
	}
	
	return "premium" // Active subscription
}

// CreateUser creates a new user with API key
func (s *Service) CreateUser(email string) (*models.User, error) {
	// Generate new API key
	apiKey := s.GenerateAPIKey()
	
	// Store API key -> email mapping
	apiKeyLookup := "api_key:" + apiKey
	err := s.redisClient.Set(s.ctx, apiKeyLookup, email, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store API key: %v", err)
	}
	
	// Create user object
	user := &models.User{
		ID:       email,
		APIKey:   apiKey,
		Tier:     "free", // Default to free tier
		Email:    email,
		Created:  time.Now(),
		LastUsed: time.Now(),
	}
	
	return user, nil
}

// Get user by API key (same as ValidateAPIKey for consistency)
func (s *Service) GetUser(apiKey string) (*models.User, error) {
	return s.ValidateAPIKey(apiKey)
}