package handlers

import (
	"fmt"
	"strings"
	"shareseer-mcp/internal/models"
)

// Search companies by ticker or name
func (h *Handler) SearchCompanies(args map[string]interface{}) (*models.ToolResult, error) {
	// Extract parameters
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: "Error: 'query' parameter is required",
			}},
		}, nil
	}
	
	// Optional limit parameter
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	
	// Get user from API key for tier checking
	apiKey, _ := args["api_key"].(string)
	user, err := h.validateAndGetUser(apiKey)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Authentication error: %v", err),
			}},
		}, nil
	}
	
	// Check rate limiting
	if !h.rateLimiter.Allow(user.ID, user.Tier) {
		return h.rateLimitError(), nil
	}
	
	// Search companies in Redis
	companies, err := h.redisClient.SearchCompanies(query, limit)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Search error: %v", err),
			}},
		}, nil
	}
	
	if len(companies) == 0 {
		return &models.ToolResult{
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("No companies found matching '%s'", query),
			}},
		}, nil
	}
	
	// Format results
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d companies matching '%s':\n\n", len(companies), query))
	
	for i, company := range companies {
		result.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, 
			company["name"], company["ticker"]))
		
		if cik, ok := company["cik"]; ok {
			result.WriteString(fmt.Sprintf("   CIK: %s\n", cik))
		}
	}
	
	// Add conversion hook for free tier users
	if user.Tier == "free" {
		result.WriteString("\n💡 Upgrade to ShareSeer Premium for advanced search and filtering")
		result.WriteString("\n🔗 Get more features at shareseer.com/upgrade?source=mcp")
	}
	
	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}

// Get basic company information
func (h *Handler) GetCompanyInfo(args map[string]interface{}) (*models.ToolResult, error) {
	ticker, ok := args["ticker"].(string)
	if !ok || ticker == "" {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: "Error: 'ticker' parameter is required",
			}},
		}, nil
	}
	
	// Get user and check rate limits
	apiKey, _ := args["api_key"].(string)
	user, err := h.validateAndGetUser(apiKey)
	if err != nil {
		return h.authError(), nil
	}
	
	if !h.rateLimiter.Allow(user.ID, user.Tier) {
		return h.rateLimitError(), nil
	}
	
	// Get company information from Redis using ticker
	company, err := h.redisClient.GetCompanyByTicker(ticker)
	if err != nil {
		return &models.ToolResult{
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Company '%s' not found", ticker),
			}},
		}, nil
	}
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("## %s (%s)\n\n", 
		company["name"], strings.ToUpper(ticker)))
	
	if exchange, ok := company["exchange"]; ok {
		result.WriteString(fmt.Sprintf("**Exchange:** %s\n", exchange))
	} else {
		result.WriteString(fmt.Sprintf("**Exchange:** %s\n", "NASDAQ")) // Default
	}
	
	if sector, ok := company["sector"]; ok {
		result.WriteString(fmt.Sprintf("**Sector:** %s\n", sector))
	}
	
	// Add link to more detailed info
	result.WriteString(fmt.Sprintf("\n📊 View detailed analysis at shareseer.com/company/%s", ticker))
	
	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}