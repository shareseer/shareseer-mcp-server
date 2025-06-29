package handlers

import (
	"fmt"
	"shareseer-mcp/internal/auth"
	"shareseer-mcp/internal/config"
	"shareseer-mcp/internal/models"
	"shareseer-mcp/internal/redis"
	"golang.org/x/time/rate"
	"sync"
)

type Handler struct {
	config      *config.Config
	redisClient *redis.Client
	authService *auth.Service
	rateLimiter *RateLimiter
}

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
}

func NewHandler(cfg *config.Config, redisClient *redis.Client, authService *auth.Service) *Handler {
	return &Handler{
		config:      cfg,
		redisClient: redisClient,
		authService: authService,
		rateLimiter: NewRateLimiter(),
	}
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (rl *RateLimiter) Allow(userID, tier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s", userID, tier)
	limiter, exists := rl.limiters[key]
	
	if !exists {
		// Create new rate limiter based on tier
		var limit rate.Limit
		switch tier {
		case "free":
			limit = rate.Limit(10.0 / 3600) // 10 requests per hour
		case "premium":
			limit = rate.Limit(100.0 / 3600) // 100 requests per hour
		case "pro":
			limit = rate.Limit(1000.0 / 3600) // 1000 requests per hour
		default:
			limit = rate.Limit(5.0 / 3600) // Default very low limit
		}
		
		limiter = rate.NewLimiter(limit, 10) // Burst of 10
		rl.limiters[key] = limiter
	}
	
	return limiter.Allow()
}

// Helper functions
func (h *Handler) validateAndGetUser(apiKey string) (*models.User, error) {
	if apiKey == "" {
		// Allow anonymous access with very limited rate limiting
		return &models.User{
			ID:   "anonymous",
			Tier: "free",
		}, nil
	}
	
	return h.authService.ValidateAPIKey(apiKey)
}

func (h *Handler) authError() *models.ToolResult {
	return &models.ToolResult{
		IsError: true,
		Content: []models.Content{{
			Type: "text",
			Text: "Authentication required. Get a free API key at shareseer.com/mcp",
		}},
	}
}

func (h *Handler) rateLimitError() *models.ToolResult {
	return &models.ToolResult{
		IsError: true,
		Content: []models.Content{{
			Type: "text",
			Text: "Rate limit exceeded. Upgrade to ShareSeer Premium for higher limits at shareseer.com/upgrade",
		}},
	}
}

func (h *Handler) tierRestrictedError(feature string) *models.ToolResult {
	return &models.ToolResult{
		IsError: true,
		Content: []models.Content{{
			Type: "text",
			Text: fmt.Sprintf("%s is available in ShareSeer Premium. Upgrade at shareseer.com/upgrade?source=mcp", feature),
		}},
	}
}

// Check if user has access to a tool
func (h *Handler) hasToolAccess(user *models.User, toolName string) bool {
	var allowedTools []string
	
	switch user.Tier {
	case "free":
		allowedTools = h.config.Tiers.Free.Tools
	case "premium":
		allowedTools = h.config.Tiers.Premium.Tools
	case "pro":
		allowedTools = h.config.Tiers.Pro.Tools
	default:
		return false
	}
	
	// Check if user has access to all tools (*)
	for _, tool := range allowedTools {
		if tool == "*" || tool == toolName {
			return true
		}
	}
	
	return false
}