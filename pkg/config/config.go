package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	APIKey      string `mapstructure:"api-key"`
	APIURL      string `mapstructure:"api-url"`
	Debug       bool   `mapstructure:"debug"`
	MaxAttempts int    `mapstructure:"max-attempts"`
	Interval    int    `mapstructure:"interval"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		APIURL:      "https://proof.testnet.polymer.zone",
		Debug:       false,
		MaxAttempts: 20,
		Interval:    3000, // in milliseconds
	}
}

// LoadConfig loads configuration from viper
func LoadConfig() (Config, error) {
	defaultConfig := DefaultConfig()

	// Set defaults if not explicitly provided
	if !viper.IsSet("api-url") {
		viper.Set("api-url", defaultConfig.APIURL)
	}
	if !viper.IsSet("debug") {
		viper.Set("debug", defaultConfig.Debug)
	}
	if !viper.IsSet("max-attempts") {
		viper.Set("max-attempts", defaultConfig.MaxAttempts)
	}
	if !viper.IsSet("interval") {
		viper.Set("interval", defaultConfig.Interval)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("API key is required. Set it using --api-key flag, POLYMER_API_KEY environment variable, or in the config file")
	}

	if c.MaxAttempts <= 0 {
		return errors.New("max-attempts must be greater than 0")
	}

	if c.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}

	return nil
}
