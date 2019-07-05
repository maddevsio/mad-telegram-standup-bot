# Internship Bot for Telegram
[![Go Report Card](https://goreportcard.com/badge/github.com/maddevsio/mad-internship-bot)](https://goreportcard.com/report/github.com/maddevsio/mad-internship-bot)
[![CircleCI](https://circleci.com/gh/maddevsio/mad-internship-bot.svg?style=svg)](https://circleci.com/gh/maddevsio/mad-internship-bot)

Automates workflow of Mad Devs Internship program requiring mentors to spend less time on repeatative tasks and actions

Currently speaks only Russian language

## Bot Skills

- Onbords new interns with predefined, customized message
- Warns about upcomming standup deadline in case inter did not write standup yet
- Automatic kicks of interns who miss more than 3 standups 
- Helps interns write more applicable, beneficial standups by analyzing standup text and outlining basic problems
- Monitors interns pull requests and gives advises on how to improve them. 
- Interns join and leave standup teams on their own (no time from mentors needed)
- Can adjust to different timezones 
- Detects standups by watching messages with bot tag and defined keywords

## Available commands
```
/help - Display list of available commands
/join - Adds you to standup team of the group
/show - Shows who submit standups
/leave - Removes you from standup team of the group
/edit_deadline - Sets new standup deadline (you can use 10am format or 15:30 format)
/show_deadline - Shows current standup deadline
/remove_deadline - Removes standup deadline at all
/tz - Change timezone of a channel
```

## Local usage
First you need to set env variables:

```
export TELEGRAM_TOKEN=yourTelegramTokenRecievedFromBotFather
export DEBUG=true
```
Then run. Note, you need `Docker` and `docker-compose` installed on your system
```
make run
```
To run tests: 
```
make clear
make test
```

## Install on your server 
1. Build and push bot's image to Dockerhub or any other container registry: 
```
docker build  -t <youraccount>/mad-internship-bot  .
```
```
docker push <youraccount>/mad-internship-bot
```
2. Enter server, install `docker` and `docker-compose` there. Create `docker-compose.yaml` file by the example from this repo
3. Create `.env` file with variables needed to run bot:
```
TELEGRAM_TOKEN=603860531:AAEB95f4tq18RWZtKLFJDFLKFDFfdsfds
DEBUG=false
ONBORDING_MESSAGE="your onbording message here"
```
4. Pull image from registry and run it in the backgroud with
```
docker-compose pull
```
```
docker-compose up -d
```

## Contribution

Any contribution is welcomed! Fork it and send PR or simply request a feature or point to bug with issues. 