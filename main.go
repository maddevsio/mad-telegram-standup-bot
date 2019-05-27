package main

import (
	"github.com/maddevsio/tgsbot/bot"
	"github.com/maddevsio/tgsbot/config"
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
