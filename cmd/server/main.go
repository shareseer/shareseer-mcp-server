package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	
	"shareseer-mcp/internal/auth"
	"shareseer-mcp/internal/config"
	"shareseer-mcp/internal/data"
	"shareseer-mcp/internal/handlers"
	"shareseer-mcp/internal/middleware"
	
	"github.com/gin-gonic/gin"
)

var Version = "dev"

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var configPath = flag.String("config", "configs/config.yaml", "Configuration file path")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ShareSeer MCP Server %s\n", Version)
		return
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize data provider (requires ShareSeer service)
	dataProvider, err := data.NewDataProvider(cfg.Data)
	if err != nil {
		log.Fatalf("Failed to initialize data provider: %v", err)
	}

	// Initialize auth service
	authService := auth.NewService()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(nil, authService, cfg)

	// Initialize handlers
	handler := handlers.NewHandler(dataProvider, rateLimiter, authService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": Version})
	})

	// MCP info endpoint
	router.GET("/mcp/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "shareseer",
			"version":     Version,
			"description": "SEC filings, insider transactions, and financial data",
			"message":     "This is the public MCP server interface. Actual functionality requires ShareSeer's proprietary data service.",
		})
	})

	// Note: Actual MCP tools implementation requires ShareSeer's data service
	router.POST("/mcp/tools/call", func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": "Not implemented",
			"message": "This public version requires ShareSeer's proprietary data service to function. Please contact ShareSeer for access.",
		})
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("ShareSeer MCP Server %s starting on %s", Version, addr)
	log.Printf("Note: This public version requires ShareSeer's data service for full functionality")
	
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
