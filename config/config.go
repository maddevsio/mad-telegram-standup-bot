package config

import "github.com/kelseyhightower/envconfig"

// BotConfig ...
type BotConfig struct {
	TelegramToken string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	DatabaseURL   string `envconfig:"DATABASE_URL" required:"true"`
	Debug         bool   `envconfig:"DEBUG" default:"false"`
}

// Get config data from environment
func Get() (*BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("", &c)
	return &c, err
}
