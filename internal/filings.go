package handlers

import (
	"fmt"
	"strings"
	"shareseer-mcp/internal/models"
)

// Get company filings
func (h *Handler) GetCompanyFilings(args map[string]interface{}) (*models.ToolResult, error) {
	// Extract ticker parameter
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
	
	// Look up CIK from ticker using direct ticker lookup
	cik, err := h.redisClient.GetCIKByTicker(ticker)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Company '%s' not found", ticker),
			}},
		}, nil
	}
	
	// Optional limit parameter
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
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
	
	// Check tool access
	if !h.hasToolAccess(user, "get_company_filings") {
		return h.tierRestrictedError("Company filings history"), nil
	}
	
	// Get filings from Redis
	filings, err := h.redisClient.GetCompanyFilings(cik, limit)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving filings: %v", err),
			}},
		}, nil
	}
	
	if len(filings) == 0 {
		return &models.ToolResult{
			Content: []models.Content{{
				Type: "text",
				Text: "No filings found for this company",
			}},
		}, nil
	}
	
	// Get company name for display
	companyName, _ := h.redisClient.GetCompanyName(cik)
	if companyName == "" {
		companyName = "Company"
	}
	
	// Format results
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Recent SEC Filings for %s:\n\n", companyName))
	
	for i, filing := range filings {
		formType, _ := filing["form_type"].(string)
		date, _ := filing["date"].(string)
		url, _ := filing["url"].(string)
		
		result.WriteString(fmt.Sprintf("%d. **%s** - %s\n", i+1, formType, date))
		result.WriteString(fmt.Sprintf("   📄 %s\n", url))
		
		if reportLink, ok := filing["report_link"].(string); ok && reportLink != "" {
			result.WriteString(fmt.Sprintf("   📊 Excel: %s\n", reportLink))
		}
		result.WriteString("\n")
	}
	
	// Add conversion hook for free tier users
	if user.Tier == "free" {
		result.WriteString("💡 Get complete filing history and real-time alerts with ShareSeer Premium")
		result.WriteString("\n🔗 Upgrade at shareseer.com/upgrade?source=mcp")
	}
	
	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}

// Get recent filings across all companies
func (h *Handler) GetRecentFilings(args map[string]interface{}) (*models.ToolResult, error) {
	// Optional limit parameter
	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
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
	
	// Check tool access
	if !h.hasToolAccess(user, "get_recent_filings") {
		return h.tierRestrictedError("Recent filings feed"), nil
	}
	
	// Get recent filings key
	recentFilings, err := h.redisClient.ZRevRange("recent:filings", 0, int64(limit-1))
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving recent filings: %v", err),
			}},
		}, nil
	}
	
	var result strings.Builder
	result.WriteString("📈 **Recent SEC Filings**\n\n")
	
	for _, filingKey := range recentFilings {
		// Get filings for this date
		dailyFilings, err := h.redisClient.ZRevRange(filingKey, 0, 4) // Top 5 for each day
		if err != nil {
			continue
		}
		
		if len(dailyFilings) > 0 {
			// Extract date from key like "today:financials:20241215"
			datePart := strings.TrimPrefix(filingKey, "today:financials:")
			result.WriteString(fmt.Sprintf("**%s:**\n", datePart))
			
			for _, filing := range dailyFilings {
				parts := strings.Split(filing, "|")
				if len(parts) >= 2 {
					companyName := parts[0]
					formType := strings.Split(parts[1], "|")[0]
					result.WriteString(fmt.Sprintf("- %s filed %s\n", companyName, formType))
				}
			}
			result.WriteString("\n")
		}
	}
	
	if user.Tier == "free" {
		result.WriteString("💡 Get real-time filing alerts and advanced filtering with ShareSeer Premium")
		result.WriteString("\n🔗 Upgrade at shareseer.com/upgrade?source=mcp")
	}
	
	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}