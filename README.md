# Simple Standup Bot for Telegram
Bot helps to conduct daily standup meetings for Mad Devs Internship Program

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
make test
```