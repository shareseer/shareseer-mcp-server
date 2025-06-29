package handlers

import (
	"shareseer-mcp/internal/data"
	"shareseer-mcp/internal/models"
)

// AbstractHandler provides base functionality for all MCP handlers
// Actual implementation details are provided by ShareSeer's data service
type AbstractHandler struct {
	dataProvider data.DataProvider
	rateLimiter  *RateLimiter
	authService  *AuthService
}

// NewHandler creates a new handler instance
func NewHandler(dataProvider data.DataProvider, rateLimiter *RateLimiter, authService *AuthService) *AbstractHandler {
	return &AbstractHandler{
		dataProvider: dataProvider,
		rateLimiter:  rateLimiter,
		authService:  authService,
	}
}

// GetLargestDailyTransactions - Implementation provided by ShareSeer service
func (h *AbstractHandler) GetLargestDailyTransactions(args map[string]interface{}) (*models.ToolResult, error) {
	return &models.ToolResult{
		IsError: true,
		Content: []models.Content{{
			Type: "text",
			Text: "This feature requires ShareSeer's proprietary data service. Please contact ShareSeer for implementation details.",
		}},
	}, nil
}

// GetLargestWeeklyTransactions - Implementation provided by ShareSeer service  
func (h *AbstractHandler) GetLargestWeeklyTransactions(args map[string]interface{}) (*models.ToolResult, error) {
	return &models.ToolResult{
		IsError: true,
		Content: []models.Content{{
			Type: "text", 
			Text: "This feature requires ShareSeer's proprietary data service. Please contact ShareSeer for implementation details.",
		}},
	}, nil
}
