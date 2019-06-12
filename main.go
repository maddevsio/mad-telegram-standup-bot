package main

import (
	"github.com/maddevsio/mad-internship-bot/bot"
	"github.com/maddevsio/mad-internship-bot/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	b, err := bot.New(c)
	if err != nil {
		log.Fatal(err)
	}

	err = b.Start()
	if err != nil {
		log.Fatal(err)
	}
}
