package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-telegram-standup-bot/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
)

//HandleUpdate function for private conversion and ...
func (b *Bot) handleUpdate(update tgbotapi.Update) (string, error) {
	message := update.Message

	if message == nil {
		message = update.EditedMessage
	}

	if message.Chat.Type == "private" {
		ok, errors := b.isStandup(message.Text, message.From.LanguageCode)
		if !ok {
			localizer := i18n.NewLocalizer(b.bundle, message.From.LanguageCode)
			text, err := localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "notStandup",
					Other: "Seems like this is not a standup, double check keywords for errors. ",
				},
			})
			if err != nil {
				log.Error(err)
			}
			text += strings.Join(errors, "\n")
			msg := tgbotapi.NewMessage(message.Chat.ID, text)
			msg.ReplyToMessageID = message.MessageID
			_, err = b.tgAPI.Send(msg)
			return text, err
		}

		localizer := i18n.NewLocalizer(b.bundle, message.From.LanguageCode)
		text, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "checkStandup",
				Other: "Good standup, post it to the group!",
			},
		})
		if err != nil {
			log.Error(err)
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyToMessageID = message.MessageID
		_, err = b.tgAPI.Send(msg)
		return text, err
	}
	var res string
	group, err := b.db.FindGroup(message.Chat.ID)
	if err != nil {
		log.Info("Could not find the group, creating...")
		group, err = b.db.CreateGroup(&model.Group{
			ChatID:          message.Chat.ID,
			Title:           message.Chat.Title,
			Username:        message.Chat.UserName,
			Description:     message.Chat.Description,
			StandupDeadline: "",
			TZ:              "Asia/Bishkek", // default value...
			Language:        "en",           // default value...
			SubmissionDays:  "monday tuesday wednesday thirsday friday saturday sunday",
		})
		if err != nil {
			return "", err
		}

		b.watchersChan <- group
	}

	if message.From.IsBot {
		return "", nil
	}

	if message.IsCommand() {
		return "", b.HandleCommand(update)
	}

	if message.Text != "" {
		b.HandleMessageEvent(message)

		containsPR, prs := containsPullRequests(message.Text)
		if containsPR {
			for _, pr := range prs {
				warnings := b.analyzePullRequest(pr, group.Language)
				if len(warnings) == 0 {
					localizer := i18n.NewLocalizer(b.bundle, group.Language)
					goodPR, err := localizer.Localize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "goodPR",
							Other: "- good PR, review indeed needed!",
						},
					})
					if err != nil {
						log.Error(err)
					}
					msg := tgbotapi.NewMessage(message.Chat.ID, *pr.HTMLURL+goodPR)
					msg.ReplyToMessageID = message.MessageID
					msg.DisableWebPagePreview = true
					b.tgAPI.Send(msg)
					res = goodPR
					return res, err
				}
				localizer := i18n.NewLocalizer(b.bundle, group.Language)
				badPR, err := localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "badPR",
						Other: "- bad PR, pay attention to the following advises: \n",
					},
				})
				if err != nil {
					log.Error(err)
				}
				text := *pr.HTMLURL + badPR
				text += strings.Join(warnings, "\n")
				msg := tgbotapi.NewMessage(message.Chat.ID, text)
				msg.ReplyToMessageID = message.MessageID
				msg.DisableWebPagePreview = true
				b.tgAPI.Send(msg)
				res = badPR
				return res, err

			}
		}
	}

	if message.LeftChatMember != nil {
		return "", b.HandleChannelLeftEvent(update)
	}

	if message.NewChatMembers != nil {
		return "", b.HandleChannelJoinEvent(update)
	}

	return "", nil
}

//HandleMessageEvent function to analyze and save standups
func (b *Bot) HandleMessageEvent(message *tgbotapi.Message) error {
	group, err := b.db.FindGroup(message.Chat.ID)
	if err != nil {
		log.Info("Could not find the group, creating...")
		group, err = b.db.CreateGroup(&model.Group{
			ChatID:          message.Chat.ID,
			Title:           message.Chat.Title,
			Username:        message.Chat.UserName,
			Description:     message.Chat.Description,
			StandupDeadline: "",
			TZ:              "Asia/Bishkek", // default value...
			Language:        "en",           // default value...
			SubmissionDays:  "monday tuesday wednesday thirsday friday saturday sunday",
		})
		if err != nil {
			return err
		}

		b.watchersChan <- group
	}

	localizer := i18n.NewLocalizer(b.bundle, group.Language)

	if !strings.Contains(message.Text, b.tgAPI.Self.UserName) {
		return nil
	}

	ok, _ := b.isStandup(message.Text, group.Language)

	if !ok {
		return fmt.Errorf("Message is not a standup")
	}

	st, err := b.db.SelectStandupByMessageID(message.MessageID, message.Chat.ID)
	if err != nil {
		log.Info("standup does not yet exist, create new standup")
		_, err := b.db.CreateStandup(&model.Standup{
			MessageID: message.MessageID,
			Created:   time.Now().UTC(),
			Modified:  time.Now().UTC(),
			Username:  message.From.UserName,
			Text:      message.Text,
			ChatID:    message.Chat.ID,
		})

		if err != nil {
			return err
		}

		greatStandup, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "greatStandup",
				Other: "Standup accepted, have a nice day!",
			},
		})
		if err != nil {
			log.Error(err)
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, greatStandup)
		msg.ReplyToMessageID = message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.UpdateStandup(st)
	if err != nil {
		log.Error("Could not update standup: ", err)
		return err
	}

	standupUpdated, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "standupUpdated",
			Other: "Standup was successfully updated!",
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, standupUpdated)
	msg.ReplyToMessageID = message.MessageID
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
		ok := b.removeTeam(event.Message.Chat.ID)
		if !ok {
			log.Error("Could not remove the team from list")
		}
		return nil
	}

	standuper, err := b.db.FindStanduper(member.ID, event.Message.Chat.ID)
	if err != nil {
		return nil
	}

	standuper.Status = "removed"
	_, err = b.db.UpdateStanduper(standuper)
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

			group, err := b.db.FindGroup(event.Message.Chat.ID)
			if err != nil {
				log.Info("Could not find the group, creating...")
				group, err = b.db.CreateGroup(&model.Group{
					ChatID:          event.Message.Chat.ID,
					Title:           event.Message.Chat.Title,
					Username:        event.Message.Chat.UserName,
					Description:     event.Message.Chat.Description,
					StandupDeadline: "",
					TZ:              "Asia/Bishkek", // default value...
					Language:        "en",           // default value...
					SubmissionDays:  "monday tuesday wednesday thirsday friday saturday sunday",
				})
				if err != nil {
					return err
				}

				b.watchersChan <- group
			}

			text := "Hello! I will help you to not forget about standups and write them properly. \n\n Additional setup include: \n '/edit_deadline 10am' (example how you edit deadline) \n '/update_onbording_message type the message here' \n '/update_group_language ru' (default is en) \n	'/group_tz Asia/Bishkek' set up your TimeZone \n '/change_submission_days monday tuesday wednesday ... ' (select days you want the bot to track standups) \n Message @anatoliyfedorenko if you find any bug or unexpected behaviour :)"

			// Send greeting message after success group save
			_, err = b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, text))
			return err
		}

		if member.IsBot {
			//Skip adding bot to standupers
			return nil
		}

		group, err := b.db.FindGroup(event.Message.Chat.ID)
		if err != nil {
			group, err = b.db.CreateGroup(&model.Group{
				ChatID:          event.Message.Chat.ID,
				Title:           event.Message.Chat.Title,
				Description:     event.Message.Chat.Description,
				StandupDeadline: "",
				TZ:              "Asia/Bishkek", // default value...
				Language:        "en",
				SubmissionDays:  "monday tuesday wednesday thirsday friday saturday sunday",
			})
			if err != nil {
				return err
			}
		}

		localizer := i18n.NewLocalizer(b.bundle, group.Language)

		welcome, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "welcomePart",
				Other: "Hello, @{{.Intern}}! Welcome to {{.GroupName}}",
			},
			TemplateData: map[string]interface{}{
				"Intern":    member.UserName,
				"GroupName": event.Message.Chat.Title,
			},
		})
		if err != nil {
			log.Error(err)
		}

		_, err = b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, welcome+group.OnbordingMessage))
		return err
	}
	return nil
}
