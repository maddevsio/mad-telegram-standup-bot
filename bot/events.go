package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) error {
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
					Other: "Seems like this is not a standup, double check keywords for errors \n\n",
				},
			})
			if err != nil {
				log.Error(err)
			}
			text += strings.Join(errors, "\n")
			msg := tgbotapi.NewMessage(message.Chat.ID, text)
			msg.ReplyToMessageID = message.MessageID
			_, err = b.tgAPI.Send(msg)
			return err
		}

		advises, _ := b.analyzeStandup(message.Text, message.From.LanguageCode)

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
		if len(advises) != 0 {
			text, err = localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "checkStandupWithAdvises",
					One:   "Standup is going to be more usefull with one advise: \n {{.Advises}}",
					Two:   "Standup is going to be more usefull with two advises: \n {{.Advises}}",
					Few:   "Standup is going to be more usefull with several advises: \n {{.Advises}}",
					Many:  "Standup is going to be more usefull with several advises: \n {{.Advises}}",
					Other: "Standup is going to be more usefull with several advises: \n {{.Advises}}",
				},
				TemplateData: map[string]interface{}{
					"Advises": strings.Join(advises, "\n"),
				},
				PluralCount: len(advises),
			})
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyToMessageID = message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	if message.From.IsBot {
		return nil
	}

	containsPR, prs := containsPullRequests(message.Text)
	if containsPR {
		for _, pr := range prs {
			warnings := b.analyzePullRequest(pr, message.From.LanguageCode)
			if len(warnings) == 0 {
				localizer := i18n.NewLocalizer(b.bundle, message.From.LanguageCode)
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
			} else {
				localizer := i18n.NewLocalizer(b.bundle, message.From.LanguageCode)
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
			}
		}
	}

	if message.IsCommand() {
		return b.HandleCommand(update)
	}

	if message.Text != "" {
		return b.HandleMessageEvent(message)
	}

	if message.LeftChatMember != nil {
		return b.HandleChannelLeftEvent(update)
	}

	if message.NewChatMembers != nil {
		return b.HandleChannelJoinEvent(update)
	}

	return nil
}

//HandleMessageEvent function to analyze and save standups
func (b *Bot) HandleMessageEvent(message *tgbotapi.Message) error {
	localizer := i18n.NewLocalizer(b.bundle, message.From.LanguageCode)

	if !strings.Contains(message.Text, b.tgAPI.Self.UserName) {
		return nil
	}

	ok, _ := b.isStandup(message.Text, message.From.LanguageCode)

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

		advises, _ := b.analyzeStandup(message.Text, message.From.LanguageCode)
		greatStandup, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "greatStandup",
				Other: "Standup accepted and it looks awesome!",
			},
		})
		if err != nil {
			log.Error(err)
		}
		text := greatStandup

		if len(advises) != 0 {
			text, err = localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "acceptStandupWithAdvises",
					One:   "Standup is accepted, but let me give you one advise: \n {{.Advises}}",
					Two:   "Standup is accepted, but let me give you two advises: \n {{.Advises}}",
					Few:   "Standup is accepted, but let me give you several advises: \n {{.Advises}}",
					Many:  "Standup is accepted, but let me give you several advises: \n {{.Advises}}",
					Other: "Standup is accepted, but let me give you several advises: \n {{.Advises}}",
				},
				TemplateData: map[string]interface{}{
					"Advises": strings.Join(advises, "\n"),
				},
				PluralCount: len(advises),
			})
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
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

			group, err := b.db.FindGroup(event.Message.Chat.ID)
			if err != nil {
				log.Info("Could not find the group, creating...")
				group, err = b.db.CreateGroup(&model.Group{
					ChatID:          event.Message.Chat.ID,
					Title:           event.Message.Chat.Title,
					Username:        event.Message.Chat.UserName,
					Description:     event.Message.Chat.Description,
					StandupDeadline: "10:00",
					TZ:              "Asia/Bishkek", // default value...
					Language:        "en",           // default value...
				})
				if err != nil {
					return err
				}

				b.watchersChan <- group
			}

			localizer := i18n.NewLocalizer(b.bundle, group.Language)
			text, err := localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "welcomeMessage",
					Other: "Hello! I will help you to not forget about standups and write them properly. tag @anatoliyfedorenko if you find any bug or unexpected behaiviour :)",
				},
			})
			if err != nil {
				log.Error(err)
			}

			// Send greeting message after success group save
			_, err = b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, text))
			return err
		}

		if member.IsBot {
			//Skip adding bot to standupers
			return nil
		}
		//if it is a regular user, greet with welcoming message and add to standupers
		_, err := b.db.FindStanduper(member.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
		if err == nil {
			return nil
		}

		_, err = b.db.CreateStanduper(&model.Standuper{
			UserID:       member.ID,
			Username:     member.UserName,
			ChatID:       event.Message.Chat.ID,
			LanguageCode: member.LanguageCode,
			TZ:           "Asia/Bishkek", // default value...
		})
		if err != nil {
			log.Error("CreateStanduper failed: ", err)
			return nil
		}

		group, err := b.db.FindGroup(event.Message.Chat.ID)
		if err != nil {
			group, err = b.db.CreateGroup(&model.Group{
				ChatID:          event.Message.Chat.ID,
				Title:           event.Message.Chat.Title,
				Description:     event.Message.Chat.Description,
				StandupDeadline: "10:00",
				TZ:              "Asia/Bishkek", // default value...
				Language:        "ru_RU",        // default value...
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
