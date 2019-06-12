package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/model"
	log "github.com/sirupsen/logrus"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) error {

	if update.Message.IsCommand() {
		return b.HandleCommand(update)
	}

	if update.Message.Text != "" {
		return b.HandleMessageEvent(update)
	}

	if update.Message.LeftChatMember != nil {
		return b.HandleChannelLeftEvent(update)
	}

	if update.Message.NewChatMembers != nil {
		return b.HandleChannelJoinEvent(update)
	}

	//? need to handle user change username
	return nil
}

//HandleMessageEvent function to analyze and save standups
func (b *Bot) HandleMessageEvent(event tgbotapi.Update) error {

	if !strings.Contains(event.Message.Text, b.tgAPI.Self.UserName) {
		return nil
	}

	if !isStandup(event.Message.Text) {
		return fmt.Errorf("Message is not a standup")
	}

	_, err := b.db.CreateStandup(&model.Standup{
		MessageID: event.Message.MessageID,
		Created:   time.Now().UTC(),
		Modified:  time.Now().UTC(),
		Username:  event.Message.From.UserName,
		Text:      event.Message.Text,
		ChatID:    event.Message.Chat.ID,
	})

	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Thanks, standup accepted!")
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//HandleChannelLeftEvent function to remove bot and standupers from channels
func (b *Bot) HandleChannelLeftEvent(event tgbotapi.Update) error {
	member := event.Message.LeftChatMember
	// if user is a bot
	if member.UserName == b.tgAPI.Self.UserName {
		team := b.findTeam(event.Message.Chat.ID)
		if team == nil {
			return fmt.Errorf("Could not find sutable team")
		}
		team.Stop()

		err := b.db.DeleteGroupStandupers(event.Message.Chat.ID)
		if err != nil {
			return err
		}
		err = b.db.DeleteGroup(team.Group.ID)
		if err != nil {
			return err
		}
		return nil
	}

	standuper, err := b.db.FindStanduper(member.UserName, event.Message.Chat.ID)
	if err != nil {
		return nil
	}
	err = b.db.DeleteStanduper(standuper.ID)
	if err != nil {
		return err
	}
	return nil
}

//HandleChannelJoinEvent function to add bot and standupers t0 channels
func (b *Bot) HandleChannelJoinEvent(event tgbotapi.Update) error {
	for _, member := range *event.Message.NewChatMembers {
		// if user is a bot
		if member.UserName == b.tgAPI.Self.UserName {

			_, err := b.db.FindGroup(event.Message.Chat.ID)
			if err != nil {
				log.Info("Could not find the group, creating...")
				group, err := b.db.CreateGroup(&model.Group{
					ChatID:          event.Message.Chat.ID,
					Title:           event.Message.Chat.Title,
					Username:        event.Message.Chat.UserName,
					Description:     event.Message.Chat.Description,
					StandupDeadline: "10:00",
					TZ:              "Asia/Bishkek", // default value...
				})
				if err != nil {
					return err
				}

				b.watchersChan <- group
			}
			// Send greeting message after success group save
			text := "Hello! Nice to meet you all! I am here to help you with standups :)"
			_, err = b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, text))
			return err
		}
		//if it is a regular user, greet with welcoming message
		text := fmt.Sprintf("Hello, @%v! Welcome to %v! To start standuping, click on /join", member.UserName, event.Message.Chat.Title)
		_, err := b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, text))
		return err
	}
	return nil
}
