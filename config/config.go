package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds all bot configuration.
// It is loaded from a JSON file (config.json) with env-var overrides.
type Config struct {
	// Polymarket CLOB API
	CLOBHost string `json:"clob_host"`
	// Polymarket Gamma API (market discovery)
	GammaHost string `json:"gamma_host"`
	// Wallet private key (hex, without 0x prefix)
	PrivateKey string `json:"private_key"`
	// API credentials (from Polymarket Builder program)
	APIKey        string `json:"api_key"`
	APISecret     string `json:"api_secret"`
	APIPassphrase string `json:"api_passphrase"`
	// Chain ID: 137 = Polygon mainnet, 80002 = Amoy testnet
	ChainID int `json:"chain_id"`

	// Strategy settings
	Strategy StrategyConfig `json:"strategy"`

	// Risk management
	Risk RiskConfig `json:"risk"`

	// Polling interval (Go duration string, e.g. "15s")
	PollIntervalStr string `json:"poll_interval"`
	PollInterval    time.Duration `json:"-"`

	// Dry run: log actions but don't submit orders
	DryRun bool `json:"dry_run"`

	// Logging level: debug | info | warn | error
	LogLevel string `json:"log_level"`
}

// StrategyConfig holds per-strategy parameters.
type StrategyConfig struct {
	// "monitor" | "spread" | "momentum"
	Name string `json:"name"`

	// Keywords to filter markets (e.g. ["bitcoin", "election"])
	MarketKeywords []string `json:"market_keywords"`

	// Maximum number of markets to track simultaneously
	MaxMarkets int `json:"max_markets"`

	// Spread strategy params
	SpreadBps     float64 `json:"spread_bps"`
	OrderSizeUSDC float64 `json:"order_size_usdc"`

	// Momentum strategy params
	MomentumThreshold float64 `json:"momentum_threshold"`
	MomentumLookback  int     `json:"momentum_lookback"`
}

// RiskConfig holds risk management limits.
type RiskConfig struct {
	MaxExposureUSDC       float64 `json:"max_exposure_usdc"`
	MaxMarketExposureUSDC float64 `json:"max_market_exposure_usdc"`
	StopLossUSDC          float64 `json:"stop_loss_usdc"`
}

// Load reads config from a JSON file, with env-var overrides.
// If the file doesn't exist, defaults are used.
func Load(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	// Parse duration string
	if cfg.PollIntervalStr != "" {
		d, err := time.ParseDuration(cfg.PollIntervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid poll_interval %q: %w", cfg.PollIntervalStr, err)
		}
		cfg.PollInterval = d
	}

	// Environment variable overrides
	if v := os.Getenv("POLYMARKET_PRIVATE_KEY"); v != "" {
		cfg.PrivateKey = v
	}
	if v := os.Getenv("POLYMARKET_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("POLYMARKET_API_SECRET"); v != "" {
		cfg.APISecret = v
	}
	if v := os.Getenv("POLYMARKET_API_PASSPHRASE"); v != "" {
		cfg.APIPassphrase = v
	}

	return cfg, nil
}

func defaults() *Config {
	return &Config{
		CLOBHost:     "https://clob.polymarket.com",
		GammaHost:    "https://gamma-api.polymarket.com",
		ChainID:      137,
		PollInterval: 10 * time.Second,
		LogLevel:     "info",
		Strategy: StrategyConfig{
			Name:              "monitor",
			MaxMarkets:        5,
			SpreadBps:         50,
			OrderSizeUSDC:     10,
			MomentumThreshold: 0.03,
			MomentumLookback:  5,
		},
		Risk: RiskConfig{
			MaxExposureUSDC:       500,
			MaxMarketExposureUSDC: 100,
			StopLossUSDC:          -50,
		},
	}
}
