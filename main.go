package main

import (
	"github.com/BurntSushi/toml"
	"github.com/maddevsio/mad-telegram-standup-bot/bot"
	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

func main() {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err := bundle.LoadMessageFile("active.en.toml")
	if err != nil {
		log.Fatal(err)
	}
	_, err = bundle.LoadMessageFile("active.ru.toml")
	if err != nil {
		log.Fatal(err)
	}

	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	b, err := bot.New(c, bundle)
	if err != nil {
		log.Fatal(err)
	}

	err = b.Start()
	if err != nil {
		log.Fatal(err)
	}
}
