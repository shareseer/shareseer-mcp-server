package handlers

import (
	"fmt"
	"strings"
	"strconv"
	"shareseer-mcp/internal/models"
)

// Get insider transactions for a company
func (h *Handler) GetInsiderTransactions(args map[string]interface{}) (*models.ToolResult, error) {
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
	if !h.hasToolAccess(user, "get_insider_transactions") {
		return h.tierRestrictedError("Insider transaction history"), nil
	}
	
	// Get company name first
	companyName, _ := h.redisClient.GetCompanyName(cik)
	
	// Get insider transactions from Redis
	transactions, err := h.redisClient.GetInsiderTransactions(cik, limit)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving insider transactions: %v", err),
			}},
		}, nil
	}
	
	if len(transactions) == 0 {
		return &models.ToolResult{
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("No insider transactions found for %s", companyName),
			}},
		}, nil
	}
	
	// Debug info - check if we have empty transactions
	if len(transactions) > 0 {
		first := transactions[0]
		if len(first) == 0 {
			return &models.ToolResult{
				Content: []models.Content{{
					Type: "text",
					Text: fmt.Sprintf("Found %d transaction keys but data is empty. First key: %v", len(transactions), first["transaction_key"]),
				}},
			}, nil
		}
	}
	
	// Format results
	var result strings.Builder
	result.WriteString(fmt.Sprintf("🏢 **Insider Transactions for %s**\n\n", companyName))
	
	for i, data := range transactions {
		// Extract transaction details using actual ShareSeer field names
		insiderName := data["name"]
		insiderTitle := data["title"]
		transactionCode := data["code"]
		transactionType := getTransactionTypeDescription(transactionCode)
		shares := data["shares"]
		price := data["price"]
		date := data["date"]
		value := data["value"]
		security := data["security"]
		_ = data["company"] // company name (not used in display since we already have it)
		
		// Format the value (already calculated in ShareSeer)
		var valueStr string
		if value != "" {
			if valueInt, err := strconv.Atoi(value); err == nil {
				valueStr = fmt.Sprintf(" ($%s)", humanizeValue(valueInt))
			}
		}
		
		result.WriteString(fmt.Sprintf("%d. **%s** - %s (%s)\n", i+1, insiderName, insiderTitle, date))
		result.WriteString(fmt.Sprintf("   📊 %s: %s shares of %s at %s%s\n", 
			transactionType, shares, security, price, valueStr))
		result.WriteString("\n")
	}
	
	// Add conversion hook for free tier users
	if user.Tier == "free" {
		result.WriteString("💡 Get complete insider trading history and real-time alerts with ShareSeer Premium")
		result.WriteString("\n🔗 Track insider patterns at shareseer.com/upgrade?source=mcp")
	}
	
	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}

// Get recent insider transactions across all companies
func (h *Handler) GetRecentInsiderActivity(args map[string]interface{}) (*models.ToolResult, error) {
	// Optional limit parameter
	limit := 15
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
	if !h.hasToolAccess(user, "get_recent_insider_activity") {
		return h.tierRestrictedError("Recent insider activity feed"), nil
	}
	
	// Get recent insider activity
	recentInsiders, err := h.redisClient.ZRevRange("recent:insiders", 0, int64(limit-1))
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving recent insider activity: %v", err),
			}},
		}, nil
	}
	
	var result strings.Builder
	result.WriteString("🔍 **Recent Insider Trading Activity**\n\n")
	
	for _, insiderKey := range recentInsiders {
		// Get transactions for this date
		dailyTransactions, err := h.redisClient.ZRevRange(insiderKey, 0, 9) // Top 10 for each day
		if err != nil {
			continue
		}
		
		if len(dailyTransactions) > 0 {
			// Extract date from key
			datePart := strings.TrimPrefix(insiderKey, "today:insiders:")
			result.WriteString(fmt.Sprintf("**%s:**\n", datePart))
			
			for _, transactionKey := range dailyTransactions {
				transactionData, err := h.redisClient.HGetAll(transactionKey)
				if err != nil {
					continue
				}
				
				companyName := transactionData["company_name"]
				insiderName := transactionData["reporting_owner_name"]
				transactionType := getTransactionTypeDescription(transactionData["transaction_code"])
				shares := transactionData["transaction_shares"]
				
				if companyName != "" && insiderName != "" {
					result.WriteString(fmt.Sprintf("- %s: %s %s %s shares\n", 
						companyName, insiderName, transactionType, shares))
				}
			}
			result.WriteString("\n")
		}
	}
	
	if user.Tier == "free" {
		result.WriteString("💡 Get real-time insider trading alerts and advanced filtering with ShareSeer Premium")
		result.WriteString("\n🔗 Never miss insider activity at shareseer.com/upgrade?source=mcp")
	}
	
	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}

// Helper function to get field value from multiple possible field names
func getFieldValue(data map[string]string, fieldNames []string) string {
	for _, fieldName := range fieldNames {
		if value, ok := data[fieldName]; ok && value != "" {
			return value
		}
	}
	return ""
}

// Helper function to humanize large numbers
func humanizeValue(value int) string {
	if value >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(value)/1000000)
	} else if value >= 1000 {
		return fmt.Sprintf("%.0fK", float64(value)/1000)
	}
	return fmt.Sprintf("%d", value)
}

// Helper function to get human-readable transaction type
func getTransactionTypeDescription(code string) string {
	transactionCodes := map[string]string{
		"P": "Purchase",
		"S": "Sale", 
		"A": "Grant/Award",
		"D": "Disposition",
		"F": "Tax Payment",
		"M": "Option Exercise",
		"G": "Gift",
		"V": "Other",
	}
	
	if desc, ok := transactionCodes[code]; ok {
		return desc
	}
	return code
}