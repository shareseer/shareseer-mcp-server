package models

import "time"

type User struct {
	ID       string    `json:"id"`
	APIKey   string    `json:"api_key"`
	Tier     string    `json:"tier"` // "free", "premium", "pro"
	Email    string    `json:"email"`
	Created  time.Time `json:"created"`
	LastUsed time.Time `json:"last_used"`
}

type Company struct {
	CIK    string `json:"cik"`
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
}

type Filing struct {
	CIK         string    `json:"cik"`
	Company     string    `json:"company"`
	FormType    string    `json:"form_type"`
	Date        time.Time `json:"date"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
}

// InsiderTransaction matches ShareSeer's structure
type InsiderTransaction struct {
	Date        string `json:"date"`        // When
	CoName      string `json:"company"`     // Company name
	InsiderName string `json:"name"`        // Insider's name
	Title       string `json:"title"`       // Position title of the insider
	Activity    string `json:"code"`        // Type of activity
	Stock       string `json:"security"`    // Common Stock or ...
	Act         string `json:"transaction"` // Acquired or Disposed
	Amount      string `json:"shares"`      // Number of shares
	Price       string `json:"price"`       // Price
	Remaining   string `json:"sharesAfter"` // Remaining number of shares after this transaction
	URL         string `json:"link"`        // SEC URL for transaction details
	Impact      string `json:"value"`       // Amount * Price
}

type FinancialData struct {
	CIK      string    `json:"cik"`
	Company  string    `json:"company"`
	Ticker   string    `json:"ticker"`
	Period   string    `json:"period"`   // "Q1 2024", "FY 2023"
	Date     time.Time `json:"date"`
	Revenue  float64   `json:"revenue"`
	NetIncome float64  `json:"net_income"`
	EPS      float64   `json:"eps"`
	Assets   float64   `json:"assets"`
	Debt     float64   `json:"debt"`
}

type RiskAssessment struct {
	CIK     string `json:"cik"`
	Company string `json:"company"`
	Ticker  string `json:"ticker"`
	Risks   string `json:"risks"`
	Updated time.Time `json:"updated"`
}

// MCP Protocol Types
type MCPRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     string      `json:"id"`
}

type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
	ID     string      `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}