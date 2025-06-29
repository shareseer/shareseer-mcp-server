package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"shareseer-mcp/internal/models"
)

// Get largest daily insider transactions
func (h *Handler) GetLargestDailyTransactions(args map[string]interface{}) (*models.ToolResult, error) {
	// Extract parameters
	txType, ok := args["type"].(string)
	if !ok || (txType != "buyers" && txType != "sellers") {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: "Error: 'type' parameter is required and must be 'buyers' or 'sellers'",
			}},
		}, nil
	}

	// Optional offset parameter for pagination
	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
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

	// Apply free tier limits (like original ShareSeer handlers)
	if user.Tier == "free" {
		if limit > 3 {
			limit = 3
		}
		if offset > 0 {
			// Free users can't paginate beyond first page
			offset = 0
		}
	}

	// Get transactions from Redis
	transactions, err := h.redisClient.GetLargestDailyTransactions(txType, offset, limit)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving largest daily %s: %v", txType, err),
			}},
		}, nil
	}

	if len(transactions) == 0 {
		return &models.ToolResult{
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("No largest daily %s found", txType),
			}},
		}, nil
	}

	// Format results
	var result strings.Builder
	emoji := "📈"
	if txType == "sellers" {
		emoji = "📉"
	}
	
	result.WriteString(fmt.Sprintf("%s **Largest Daily %s**\n\n", emoji, strings.Title(txType)))

	for i, data := range transactions {
		insiderName := data["name"]
		company := data["company"]
		title := data["title"]
		shares := data["shares"]
		price := data["price"]
		date := data["date"]
		value := data["value"]
		security := data["security"]

		// Format the value
		var valueStr string
		if value != "" {
			if valueInt, err := strconv.Atoi(value); err == nil {
				if valueInt >= 1000000 {
					valueStr = fmt.Sprintf("$%.1fM", float64(valueInt)/1000000)
				} else if valueInt >= 1000 {
					valueStr = fmt.Sprintf("$%.0fK", float64(valueInt)/1000)
				} else {
					valueStr = fmt.Sprintf("$%s", value)
				}
			} else {
				valueStr = fmt.Sprintf("$%s", value)
			}
		}

		// Transaction type description
		transactionType := "Transaction"
		if txType == "buyers" {
			transactionType = "Purchase"
		} else {
			transactionType = "Sale"
		}

		result.WriteString(fmt.Sprintf("%d. **%s** - %s (%s)\n", i+1, insiderName, title, date))
		result.WriteString(fmt.Sprintf("   🏢 %s\n", company))
		result.WriteString(fmt.Sprintf("   📊 %s: %s shares of %s at $%s (%s)\n\n", 
			transactionType, shares, security, price, valueStr))
	}

	// Add pagination info
	if offset > 0 || len(transactions) == limit {
		result.WriteString(fmt.Sprintf("📄 Showing results %d-%d", offset+1, offset+len(transactions)))
		if len(transactions) == limit {
			result.WriteString(" (more available)")
		}
		result.WriteString("\n\n")
	}

	// Add conversion hook for free tier users
	if user.Tier == "free" {
		result.WriteString("💡 Free tier: Limited to 3 results, no pagination. Get unlimited access with ShareSeer Premium\n")
		result.WriteString("🔗 Upgrade at shareseer.com/upgrade?source=mcp")
	}

	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}

// Get largest weekly insider transactions
func (h *Handler) GetLargestWeeklyTransactions(args map[string]interface{}) (*models.ToolResult, error) {
	// Extract parameters
	txType, ok := args["type"].(string)
	if !ok || (txType != "buyers" && txType != "sellers") {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: "Error: 'type' parameter is required and must be 'buyers' or 'sellers'",
			}},
		}, nil
	}

	// Optional week offset parameter (0 = current week, 1 = last week, etc.)
	weekOffset := 0
	if w, ok := args["week_offset"].(float64); ok {
		weekOffset = int(w)
	}

	// Optional offset parameter for pagination
	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
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

	// Apply free tier limits (like original ShareSeer handlers)
	if user.Tier == "free" {
		if limit > 3 {
			limit = 3
		}
		if offset > 0 {
			// Free users can't paginate beyond first page
			offset = 0
		}
		if weekOffset > 0 {
			// Free users can only see current week
			weekOffset = 0
		}
	}

	// Get transactions from Redis
	transactions, targetWeek, currentDate, previousDate, err := h.redisClient.GetLargestWeeklyTransactions(txType, weekOffset, offset, limit)
	if err != nil {
		return &models.ToolResult{
			IsError: true,
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving largest weekly %s: %v", txType, err),
			}},
		}, nil
	}

	if len(transactions) == 0 {
		weekDesc := "this week"
		if weekOffset > 0 {
			weekDesc = fmt.Sprintf("%d week(s) ago", weekOffset)
		}
		return &models.ToolResult{
			Content: []models.Content{{
				Type: "text",
				Text: fmt.Sprintf("No largest weekly %s found for %s", txType, weekDesc),
			}},
		}, nil
	}

	// Format results
	var result strings.Builder
	emoji := "📈"
	if txType == "sellers" {
		emoji = "📉"
	}

	weekDesc := "This Week"
	if weekOffset > 0 {
		weekDesc = fmt.Sprintf("Week %d (%s)", targetWeek, previousDate)
	} else {
		weekDesc = fmt.Sprintf("This Week (%s)", currentDate)
	}

	result.WriteString(fmt.Sprintf("%s **Largest Weekly %s - %s**\n\n", emoji, strings.Title(txType), weekDesc))

	for i, data := range transactions {
		insiderName := data["name"]
		company := data["company"]
		title := data["title"]
		shares := data["shares"]
		price := data["price"]
		date := data["date"]
		value := data["value"]
		security := data["security"]

		// Format the value
		var valueStr string
		if value != "" {
			if valueInt, err := strconv.Atoi(value); err == nil {
				if valueInt >= 1000000 {
					valueStr = fmt.Sprintf("$%.1fM", float64(valueInt)/1000000)
				} else if valueInt >= 1000 {
					valueStr = fmt.Sprintf("$%.0fK", float64(valueInt)/1000)
				} else {
					valueStr = fmt.Sprintf("$%s", value)
				}
			} else {
				valueStr = fmt.Sprintf("$%s", value)
			}
		}

		// Transaction type description
		transactionType := "Transaction"
		if txType == "buyers" {
			transactionType = "Purchase"
		} else {
			transactionType = "Sale"
		}

		result.WriteString(fmt.Sprintf("%d. **%s** - %s (%s)\n", i+1, insiderName, title, date))
		result.WriteString(fmt.Sprintf("   🏢 %s\n", company))
		result.WriteString(fmt.Sprintf("   📊 %s: %s shares of %s at $%s (%s)\n\n", 
			transactionType, shares, security, price, valueStr))
	}

	// Add pagination info
	if offset > 0 || len(transactions) == limit {
		result.WriteString(fmt.Sprintf("📄 Showing results %d-%d", offset+1, offset+len(transactions)))
		if len(transactions) == limit {
			result.WriteString(" (more available)")
		}
		result.WriteString("\n\n")
	}

	// Add conversion hook for free tier users
	if user.Tier == "free" {
		result.WriteString("💡 Free tier: Limited to 3 results, current week only, no pagination. Get full access with ShareSeer Premium\n")
		result.WriteString("🔗 Upgrade at shareseer.com/upgrade?source=mcp")
	}

	return &models.ToolResult{
		Content: []models.Content{{
			Type: "text",
			Text: result.String(),
		}},
	}, nil
}