package data

import (
	"fmt"
	"time"
)

// shareSeerDataProvider implements DataProvider for ShareSeer data
// Internal implementation details are abstracted
type shareSeerDataProvider struct {
	connection dataConnection
	config     providerConfig
}

type dataConnection interface {
	Query(query string, params ...interface{}) ([]map[string]string, error)
	IsConnected() bool
}

type providerConfig struct {
	Host     string
	Port     string
	Database string
	ReadOnly bool
}

// newShareSeerDataProvider creates a new ShareSeer data provider
func newShareSeerDataProvider(config map[string]interface{}) (DataProvider, error) {
	// Extract configuration
	providerConfig := providerConfig{
		Host:     getConfigString(config, "host", "localhost"),
		Port:     getConfigString(config, "port", "6379"),
		Database: getConfigString(config, "database", "1"),
		ReadOnly: getConfigBool(config, "read_only", true),
	}
	
	// Create connection (implementation specific)
	conn, err := createDataConnection(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create data connection: %w", err)
	}
	
	return &shareSeerDataProvider{
		connection: conn,
		config:     providerConfig,
	}, nil
}

// GetCompanyByTicker retrieves company information by ticker symbol
func (p *shareSeerDataProvider) GetCompanyByTicker(ticker string) (*Company, error) {
	// Query implementation abstracted
	results, err := p.connection.Query("GET_COMPANY_BY_TICKER", ticker)
	if err != nil {
		return nil, err
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("company not found")
	}
	
	return p.mapToCompany(results[0]), nil
}

// SearchCompanies searches for companies by name or ticker
func (p *shareSeerDataProvider) SearchCompanies(query string, limit int) ([]*Company, error) {
	results, err := p.connection.Query("SEARCH_COMPANIES", query, limit)
	if err != nil {
		return nil, err
	}
	
	var companies []*Company
	for _, result := range results {
		companies = append(companies, p.mapToCompany(result))
	}
	
	return companies, nil
}

// GetCompanyFilings retrieves SEC filings for a company
func (p *shareSeerDataProvider) GetCompanyFilings(companyID string, limit int) ([]*Filing, error) {
	results, err := p.connection.Query("GET_COMPANY_FILINGS", companyID, limit)
	if err != nil {
		return nil, err
	}
	
	var filings []*Filing
	for _, result := range results {
		filings = append(filings, p.mapToFiling(result))
	}
	
	return filings, nil
}

// GetRecentFilings retrieves recent SEC filings across all companies
func (p *shareSeerDataProvider) GetRecentFilings(limit int) ([]*Filing, error) {
	results, err := p.connection.Query("GET_RECENT_FILINGS", limit)
	if err != nil {
		return nil, err
	}
	
	var filings []*Filing
	for _, result := range results {
		filings = append(filings, p.mapToFiling(result))
	}
	
	return filings, nil
}

// GetInsiderTransactions retrieves insider transactions for a company
func (p *shareSeerDataProvider) GetInsiderTransactions(companyID string, limit int) ([]*InsiderTransaction, error) {
	results, err := p.connection.Query("GET_INSIDER_TRANSACTIONS", companyID, limit)
	if err != nil {
		return nil, err
	}
	
	var transactions []*InsiderTransaction
	for _, result := range results {
		transactions = append(transactions, p.mapToInsiderTransaction(result))
	}
	
	return transactions, nil
}

// GetRecentInsiderActivity retrieves recent insider activity across all companies
func (p *shareSeerDataProvider) GetRecentInsiderActivity(limit int) ([]*InsiderTransaction, error) {
	results, err := p.connection.Query("GET_RECENT_INSIDER_ACTIVITY", limit)
	if err != nil {
		return nil, err
	}
	
	var transactions []*InsiderTransaction
	for _, result := range results {
		transactions = append(transactions, p.mapToInsiderTransaction(result))
	}
	
	return transactions, nil
}

// GetLargestDailyTransactions retrieves largest daily insider transactions
func (p *shareSeerDataProvider) GetLargestDailyTransactions(txType string, offset, limit int) ([]*InsiderTransaction, error) {
	results, err := p.connection.Query("GET_LARGEST_DAILY_TRANSACTIONS", txType, offset, limit)
	if err != nil {
		return nil, err
	}
	
	var transactions []*InsiderTransaction
	for _, result := range results {
		transactions = append(transactions, p.mapToInsiderTransaction(result))
	}
	
	return transactions, nil
}

// GetLargestWeeklyTransactions retrieves largest weekly insider transactions
func (p *shareSeerDataProvider) GetLargestWeeklyTransactions(txType string, weekOffset, offset, limit int) ([]*InsiderTransaction, string, string, error) {
	results, err := p.connection.Query("GET_LARGEST_WEEKLY_TRANSACTIONS", txType, weekOffset, offset, limit)
	if err != nil {
		return nil, "", "", err
	}
	
	var transactions []*InsiderTransaction
	for _, result := range results {
		transactions = append(transactions, p.mapToInsiderTransaction(result))
	}
	
	// Calculate dates for display
	now := time.Now()
	currentDate := now.Format("01/02/2006")
	targetDate := now.AddDate(0, 0, -7*weekOffset)
	previousDate := targetDate.Format("01/02/2006")
	
	return transactions, currentDate, previousDate, nil
}

// IsHealthy checks if the data provider is healthy
func (p *shareSeerDataProvider) IsHealthy() bool {
	return p.connection.IsConnected()
}

// Helper functions for mapping data (implementation specific)
func (p *shareSeerDataProvider) mapToCompany(data map[string]string) *Company {
	return &Company{
		ID:       data["id"],
		Ticker:   data["ticker"],
		Name:     data["name"],
		Exchange: data["exchange"],
		Sector:   data["sector"],
	}
}

func (p *shareSeerDataProvider) mapToFiling(data map[string]string) *Filing {
	return &Filing{
		ID:          data["id"],
		CompanyID:   data["company_id"],
		Company:     data["company"],
		FormType:    data["form_type"],
		Date:        data["date"],
		URL:         data["url"],
		Description: data["description"],
	}
}

func (p *shareSeerDataProvider) mapToInsiderTransaction(data map[string]string) *InsiderTransaction {
	return &InsiderTransaction{
		ID:              data["id"],
		CompanyID:       data["company_id"],
		Company:         data["company"],
		Date:            data["date"],
		InsiderName:     data["insider_name"],
		Title:           data["title"],
		TransactionType: data["transaction_type"],
		SecurityType:    data["security_type"],
		Shares:          data["shares"],
		Price:           data["price"],
		Value:           data["value"],
		SharesAfter:     data["shares_after"],
		URL:             data["url"],
	}
}

// Helper functions for configuration
func getConfigString(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getConfigBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

// createDataConnection creates a connection to the data source
// Implementation details are abstracted and may vary by deployment
func createDataConnection(config providerConfig) (dataConnection, error) {
	// This would contain the actual connection logic
	// For public repo, we'll keep this abstract
	return &abstractDataConnection{config: config}, nil
}

// abstractDataConnection provides a generic data connection interface
type abstractDataConnection struct {
	config providerConfig
}

func (c *abstractDataConnection) Query(query string, params ...interface{}) ([]map[string]string, error) {
	// Implementation abstracted - would contain actual query logic
	// This allows the public repo to compile without revealing internal details
	return nil, fmt.Errorf("data connection not implemented - this is a template")
}

func (c *abstractDataConnection) IsConnected() bool {
	// Implementation abstracted
	return false
}