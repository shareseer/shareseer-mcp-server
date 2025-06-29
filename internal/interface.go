package data

// DataProvider defines the interface for accessing financial data
// Implementation details are abstracted away from the public repository
type DataProvider interface {
	// Company data operations
	GetCompanyByTicker(ticker string) (*Company, error)
	SearchCompanies(query string, limit int) ([]*Company, error)
	
	// SEC filings operations
	GetCompanyFilings(companyID string, limit int) ([]*Filing, error)
	GetRecentFilings(limit int) ([]*Filing, error)
	
	// Insider trading operations
	GetInsiderTransactions(companyID string, limit int) ([]*InsiderTransaction, error)
	GetRecentInsiderActivity(limit int) ([]*InsiderTransaction, error)
	GetLargestDailyTransactions(txType string, offset, limit int) ([]*InsiderTransaction, error)
	GetLargestWeeklyTransactions(txType string, weekOffset, offset, limit int) ([]*InsiderTransaction, string, string, error)
	
	// Health check
	IsHealthy() bool
}

// Company represents basic company information
type Company struct {
	ID       string `json:"id"`        // Internal company identifier
	Ticker   string `json:"ticker"`    // Stock ticker symbol
	Name     string `json:"name"`      // Company name
	Exchange string `json:"exchange"`  // Stock exchange
	Sector   string `json:"sector"`    // Business sector
}

// Filing represents an SEC filing
type Filing struct {
	ID          string `json:"id"`          // Internal filing identifier
	CompanyID   string `json:"company_id"`  // Internal company identifier  
	Company     string `json:"company"`     // Company name
	FormType    string `json:"form_type"`   // SEC form type (10-K, 10-Q, etc.)
	Date        string `json:"date"`        // Filing date
	URL         string `json:"url"`         // SEC filing URL
	Description string `json:"description"` // Filing description
}

// InsiderTransaction represents an insider trading transaction
type InsiderTransaction struct {
	ID            string `json:"id"`             // Internal transaction identifier
	CompanyID     string `json:"company_id"`     // Internal company identifier
	Company       string `json:"company"`        // Company name
	Date          string `json:"date"`           // Transaction date
	InsiderName   string `json:"insider_name"`   // Name of the insider
	Title         string `json:"title"`          // Insider's position/title
	TransactionType string `json:"transaction_type"` // Buy/Sell/Exercise/etc.
	SecurityType  string `json:"security_type"`  // Type of security
	Shares        string `json:"shares"`         // Number of shares
	Price         string `json:"price"`          // Price per share
	Value         string `json:"value"`          // Total transaction value
	SharesAfter   string `json:"shares_after"`   // Shares owned after transaction
	URL           string `json:"url"`            // SEC filing URL
}

// NewDataProvider creates a new data provider instance
// The specific implementation is determined by configuration
func NewDataProvider(config map[string]interface{}) (DataProvider, error) {
	// Implementation details hidden in private package
	return newShareSeerDataProvider(config)
}