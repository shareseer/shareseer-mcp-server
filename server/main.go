package main

import (
	"fmt"
	"log"
	"shareseer-mcp/internal/auth"
	"shareseer-mcp/internal/config"
	"shareseer-mcp/internal/handlers"
	"shareseer-mcp/internal/middleware"
	"shareseer-mcp/internal/models"
	"shareseer-mcp/internal/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Redis
	redisClient, err := redis.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize auth service
	authService := auth.NewService()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient.GetRedisClient(), authService, cfg)

	// Initialize handlers
	handler := handlers.NewHandler(cfg, redisClient, authService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware for MCP clients
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "shareseer-mcp"})
	})

	// Apply rate limiting to MCP tools endpoint
	router.POST("/mcp/tools/call", rateLimiter.Middleware(), handleToolCall(handler))

	// MCP server info
	router.GET("/mcp/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        cfg.MCP.ServerName,
			"version":     cfg.MCP.Version,
			"description": cfg.MCP.Description,
			"tools": []gin.H{
				{
					"name":        "get_company_info",
					"description": "Get basic information about a company",
					"parameters": gin.H{
						"ticker":  "string (required) - Company ticker symbol",
						"api_key": "string (optional) - Your ShareSeer API key",
					},
				},
				{
					"name":        "get_company_filings",
					"description": "Get recent SEC filings for a specific company",
					"parameters": gin.H{
						"ticker":  "string (required) - Company ticker symbol", 
						"limit":   "number (optional) - Maximum number of filings (default: 10)",
						"api_key": "string (optional) - Your ShareSeer API key",
					},
				},
				{
					"name":        "get_recent_filings",
					"description": "Get recent SEC filings across all companies",
					"parameters": gin.H{
						"limit":   "number (optional) - Maximum number of filings (default: 20)",
						"api_key": "string (optional) - Your ShareSeer API key",
					},
				},
				{
					"name":        "get_insider_transactions",
					"description": "Get insider trading transactions for a specific company",
					"parameters": gin.H{
						"ticker":  "string (required) - Company ticker symbol",
						"limit":   "number (optional) - Maximum number of transactions (default: 10)",
						"api_key": "string (optional) - Your ShareSeer API key",
					},
				},
				{
					"name":        "get_recent_insider_activity",
					"description": "Get recent insider trading activity across all companies",
					"parameters": gin.H{
						"limit":   "number (optional) - Maximum number of transactions (default: 15)",
						"api_key": "string (optional) - Your ShareSeer API key",
					},
				},
				{
					"name":        "get_largest_daily_transactions",
					"description": "Get largest daily insider transactions (buyers or sellers)",
					"parameters": gin.H{
						"type":    "string (required) - Transaction type: 'buyers' or 'sellers'",
						"offset":  "number (optional) - Pagination offset (default: 0)",
						"limit":   "number (optional) - Maximum number of transactions (default: 10)",
						"api_key": "string (optional) - Your ShareSeer API key",
					},
				},
				{
					"name":        "get_largest_weekly_transactions",
					"description": "Get largest weekly insider transactions (buyers or sellers)",
					"parameters": gin.H{
						"type":        "string (required) - Transaction type: 'buyers' or 'sellers'",
						"week_offset": "number (optional) - Week offset: 0=current, 1=last week (default: 0)",
						"offset":      "number (optional) - Pagination offset (default: 0)",
						"limit":       "number (optional) - Maximum number of transactions (default: 10)",
						"api_key":     "string (optional) - Your ShareSeer API key",
					},
				},
			},
		})
	})

	// GET endpoints for easier testing (with rate limiting)
	router.GET("/test/filings/:cik", rateLimiter.Middleware(), func(c *gin.Context) {
		cik := c.Param("cik")
		result, err := handler.GetCompanyFilings(map[string]interface{}{
			"cik":   cik,
			"limit": 10,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, result)
	})
	
	router.GET("/test/insiders/:cik", rateLimiter.Middleware(), func(c *gin.Context) {
		cik := c.Param("cik")
		result, err := handler.GetInsiderTransactions(map[string]interface{}{
			"cik":   cik,
			"limit": 5,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, result)
	})

	// API key management endpoints
	router.POST("/api/users", func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		user, err := authService.CreateUser(req.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{
			"user":    user,
			"message": "API key created successfully. Use this key in your MCP tool calls.",
		})
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	
	if cfg.Server.HTTPS {
		log.Printf("Starting ShareSeer MCP server on https://%s", addr)
		log.Printf("MCP info available at: https://%s/mcp/info", addr)
		
		if err := router.RunTLS(addr, cfg.Server.CertFile, cfg.Server.KeyFile); err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		log.Printf("Starting ShareSeer MCP server on http://%s", addr)
		log.Printf("MCP info available at: http://%s/mcp/info", addr)
		
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

func handleToolCall(handler *handlers.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var toolCall models.ToolCall

		if err := c.ShouldBindJSON(&toolCall); err != nil {
			c.JSON(400, gin.H{"error": "Invalid tool call format"})
			return
		}

		var result *models.ToolResult
		var err error

		// Route to appropriate handler based on tool name
		switch toolCall.Name {
		case "get_company_info":
			result, err = handler.GetCompanyInfo(toolCall.Arguments)
		case "get_company_filings":
			result, err = handler.GetCompanyFilings(toolCall.Arguments)
		case "get_recent_filings":
			result, err = handler.GetRecentFilings(toolCall.Arguments)
		case "get_insider_transactions":
			result, err = handler.GetInsiderTransactions(toolCall.Arguments)
		case "get_recent_insider_activity":
			result, err = handler.GetRecentInsiderActivity(toolCall.Arguments)
		case "get_largest_daily_transactions":
			result, err = handler.GetLargestDailyTransactions(toolCall.Arguments)
		case "get_largest_weekly_transactions":
			result, err = handler.GetLargestWeeklyTransactions(toolCall.Arguments)
		default:
			result = &models.ToolResult{
				IsError: true,
				Content: []models.Content{{
					Type: "text",
					Text: fmt.Sprintf("Unknown tool: %s", toolCall.Name),
				}},
			}
		}

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, result)
	}
}