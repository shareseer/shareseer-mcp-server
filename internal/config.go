package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		HTTPS    bool   `yaml:"https"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"server"`
	
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		ReadOnly bool   `yaml:"read_only"`
	} `yaml:"redis"`
	
	MCP struct {
		ServerName  string `yaml:"server_name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
	} `yaml:"mcp"`
	
	RateLimiting struct {
		FreeTier struct {
			RequestsPerDay  int `yaml:"requests_per_day"`
			RequestsPerHour int `yaml:"requests_per_hour"`
		} `yaml:"free_tier"`
		PremiumTier struct {
			RequestsPerDay  int `yaml:"requests_per_day"`
			RequestsPerHour int `yaml:"requests_per_hour"`
		} `yaml:"premium_tier"`
		ProTier struct {
			RequestsPerDay  int `yaml:"requests_per_day"`
			RequestsPerHour int `yaml:"requests_per_hour"`
		} `yaml:"pro_tier"`
	} `yaml:"rate_limiting"`
	
	Tiers struct {
		Free struct {
			DataRangeMonths int      `yaml:"data_range_months"`
			Tools           []string `yaml:"tools"`
		} `yaml:"free"`
		Premium struct {
			DataRangeMonths int      `yaml:"data_range_months"`
			Tools           []string `yaml:"tools"`
		} `yaml:"premium"`
		Pro struct {
			DataRangeMonths int      `yaml:"data_range_months"`
			Tools           []string `yaml:"tools"`
		} `yaml:"pro"`
	} `yaml:"tiers"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	
	return config, nil
}