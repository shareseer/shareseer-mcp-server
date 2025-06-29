package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"shareseer-mcp/internal/auth"
	"shareseer-mcp/internal/config"
)

type RateLimiter struct {
	redisClient *redis.Client
	authService *auth.Service
	config      *config.Config
	ctx         context.Context
}

type RateLimitInfo struct {
	HourlyLimit     int
	DailyLimit      int
	HourlyUsed      int
	DailyUsed       int
	HourlyRemaining int
	DailyRemaining  int
	ResetHour       int64
	ResetDay        int64
}

func NewRateLimiter(redisClient *redis.Client, authService *auth.Service, cfg *config.Config) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		authService: authService,
		config:      cfg,
		ctx:         context.Background(),
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from request
		apiKey := rl.extractAPIKey(c)
		if apiKey == "" {
			// Allow requests without API key for now (they'll be handled by auth in handlers)
			c.Next()
			return
		}

		// Validate API key and get user tier
		user, err := rl.authService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Check rate limits
		rateLimitInfo, allowed := rl.checkRateLimit(apiKey, user.Tier)
		if !allowed {
			// Add rate limit headers
			rl.addRateLimitHeaders(c, rateLimitInfo)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("You have exceeded your %s tier limits", user.Tier),
				"limits": gin.H{
					"hourly": rateLimitInfo.HourlyLimit,
					"daily":  rateLimitInfo.DailyLimit,
				},
				"usage": gin.H{
					"hourly": rateLimitInfo.HourlyUsed,
					"daily":  rateLimitInfo.DailyUsed,
				},
				"reset_times": gin.H{
					"next_hour": rateLimitInfo.ResetHour,
					"next_day":  rateLimitInfo.ResetDay,
				},
			})
			c.Abort()
			return
		}

		// Increment usage counters
		rl.incrementUsage(apiKey)

		// Add rate limit headers to successful responses
		rl.addRateLimitHeaders(c, rateLimitInfo)

		// Store user in context for handlers to use
		c.Set("user", user)
		c.Set("api_key", apiKey)

		c.Next()
	}
}

func (rl *RateLimiter) extractAPIKey(c *gin.Context) string {
	// Try Authorization header first (Bearer token)
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try query parameter
	if apiKey := c.Query("api_key"); apiKey != "" {
		return apiKey
	}

	// For MCP tools/call endpoint, skip JSON body parsing to avoid consuming request body
	// The API key will be validated later in the handler after JSON parsing
	if strings.Contains(c.Request.URL.Path, "/mcp/tools/call") {
		return ""
	}

	// Try api_key parameter in request body for POST requests (other endpoints)
	if c.Request.Method == "POST" {
		if apiKey, exists := c.GetPostForm("api_key"); exists {
			return apiKey
		}
		
		// Try JSON body
		var body map[string]interface{}
		if c.ShouldBindJSON(&body) == nil {
			if apiKey, ok := body["api_key"].(string); ok {
				return apiKey
			}
			if args, ok := body["arguments"].(map[string]interface{}); ok {
				if apiKey, ok := args["api_key"].(string); ok {
					return apiKey
				}
			}
		}
	}

	return ""
}

func (rl *RateLimiter) checkRateLimit(apiKey, tier string) (*RateLimitInfo, bool) {
	now := time.Now()
	hourKey := fmt.Sprintf("rate_limit:hourly:%s:%s", apiKey, now.Format("2006010215"))
	dayKey := fmt.Sprintf("rate_limit:daily:%s:%s", apiKey, now.Format("20060102"))

	// Get current usage
	pipe := rl.redisClient.Pipeline()
	hourlyCmd := pipe.Get(rl.ctx, hourKey)
	dailyCmd := pipe.Get(rl.ctx, dayKey)
	_, err := pipe.Exec(rl.ctx)

	hourlyUsed := 0
	dailyUsed := 0

	if err == nil {
		if hourlyVal, err := hourlyCmd.Result(); err == nil {
			hourlyUsed, _ = strconv.Atoi(hourlyVal)
		}
		if dailyVal, err := dailyCmd.Result(); err == nil {
			dailyUsed, _ = strconv.Atoi(dailyVal)
		}
	}

	// Get limits based on tier
	limits := rl.getTierLimits(tier)
	
	// Calculate reset times
	nextHour := now.Truncate(time.Hour).Add(time.Hour).Unix()
	nextDay := now.Truncate(24 * time.Hour).Add(24 * time.Hour).Unix()

	rateLimitInfo := &RateLimitInfo{
		HourlyLimit:     limits["hourly"],
		DailyLimit:      limits["daily"],
		HourlyUsed:      hourlyUsed,
		DailyUsed:       dailyUsed,
		HourlyRemaining: limits["hourly"] - hourlyUsed,
		DailyRemaining:  limits["daily"] - dailyUsed,
		ResetHour:       nextHour,
		ResetDay:        nextDay,
	}

	// Check if under limits
	allowed := hourlyUsed < limits["hourly"] && dailyUsed < limits["daily"]
	
	return rateLimitInfo, allowed
}

func (rl *RateLimiter) incrementUsage(apiKey string) {
	now := time.Now()
	hourKey := fmt.Sprintf("rate_limit:hourly:%s:%s", apiKey, now.Format("2006010215"))
	dayKey := fmt.Sprintf("rate_limit:daily:%s:%s", apiKey, now.Format("20060102"))

	pipe := rl.redisClient.Pipeline()
	
	// Increment counters
	pipe.Incr(rl.ctx, hourKey)
	pipe.Expire(rl.ctx, hourKey, time.Hour)
	pipe.Incr(rl.ctx, dayKey)
	pipe.Expire(rl.ctx, dayKey, 24*time.Hour)
	
	pipe.Exec(rl.ctx)
}

func (rl *RateLimiter) getTierLimits(tier string) map[string]int {
	switch tier {
	case "premium":
		return map[string]int{
			"hourly": rl.config.RateLimiting.PremiumTier.RequestsPerHour,
			"daily":  rl.config.RateLimiting.PremiumTier.RequestsPerDay,
		}
	case "pro":
		return map[string]int{
			"hourly": rl.config.RateLimiting.ProTier.RequestsPerHour,
			"daily":  rl.config.RateLimiting.ProTier.RequestsPerDay,
		}
	default: // free
		return map[string]int{
			"hourly": rl.config.RateLimiting.FreeTier.RequestsPerHour,
			"daily":  rl.config.RateLimiting.FreeTier.RequestsPerDay,
		}
	}
}

func (rl *RateLimiter) addRateLimitHeaders(c *gin.Context, info *RateLimitInfo) {
	c.Header("X-RateLimit-Limit-Hourly", strconv.Itoa(info.HourlyLimit))
	c.Header("X-RateLimit-Remaining-Hourly", strconv.Itoa(info.HourlyRemaining))
	c.Header("X-RateLimit-Limit-Daily", strconv.Itoa(info.DailyLimit))
	c.Header("X-RateLimit-Remaining-Daily", strconv.Itoa(info.DailyRemaining))
	c.Header("X-RateLimit-Reset-Hour", strconv.FormatInt(info.ResetHour, 10))
	c.Header("X-RateLimit-Reset-Day", strconv.FormatInt(info.ResetDay, 10))
}