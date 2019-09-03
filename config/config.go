package config

import (
	"github.com/kelseyhightower/envconfig"
)

// BotConfig ...
type BotConfig struct {
	TelegramToken    string `envconfig:"TELEGRAM_TOKEN" required:"false"`
	DatabaseURL      string `envconfig:"DATABASE_URL" required:"false" default:"telegram:telegram@tcp(localhost:3306)/telegram?parseTime=true"`
	MaxReminders     int    `envconfig:"MAX_REMINDERS" default:"3"`
	NotificationTime int    `envconfig:"NOTIFICATION_TIME" default:"30"`
}

// Get config data from environment
func Get() (*BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("", &c)
	return &c, err
}
