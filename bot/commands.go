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

//HandleCommand handles imcomming commands
func (b *Bot) HandleCommand(event tgbotapi.Update) error {
	switch event.Message.Command() {
	case "help":
		return b.Help(event)
	case "join":
		return b.JoinStandupers(event)
	case "show":
		return b.Show(event)
	case "leave":
		return b.LeaveStandupers(event)
	case "edit_deadline":
		return b.EditDeadline(event)
	case "update_onbording_message":
		return b.UpdateOnbordingMessage(event)
	case "update_group_language":
		return b.UpdateGroupLanguage(event)
	case "change_submission_days":
		return b.ChangeSubmissionDays(event)
	case "group_tz":
		return b.ChangeGroupTimeZone(event)
	case "tz":
		return b.ChangeUserTimeZone(event)
	default:
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "I do not know this command...")
		_, err := b.tgAPI.Send(msg)
		return err
	}
}

//Help displays help message
func (b *Bot) Help(event tgbotapi.Update) error {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)
	helpText, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "helpText",
			Other: `In order to submit a standup, tag me and write a message with keywords. Direct message me to see the list of keywords needed. Loking forward for your standups! Message @anatoliyfedorenko in case of any unexpected behaviour, submit issues to https://github.com/maddevsio/mad-internship-bot/issues`,
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, helpText)
	_, err = b.tgAPI.Send(msg)
	return err
}

//JoinStandupers assign user a standuper role
func (b *Bot) JoinStandupers(event tgbotapi.Update) error {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)
	_, err := b.db.FindStanduper(event.Message.From.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
	if err == nil {
		youAlreadyStandup, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "youAlreadyStandup",
				Other: "You already a part of standup team",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, youAlreadyStandup)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.CreateStanduper(&model.Standuper{
		UserID:       event.Message.From.ID,
		Username:     event.Message.From.UserName,
		ChatID:       event.Message.Chat.ID,
		LanguageCode: event.Message.From.LanguageCode,
		TZ:           "Asia/Bishkek", // default value...
	})
	if err != nil {
		log.Error("CreateStanduper failed: ", err)
		createStanduperFailed, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "createStanduperFailed",
				Other: "Could not add you to standup team",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, createStanduperFailed)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	group, err := b.db.FindGroup(event.Message.Chat.ID)
	if err != nil {
		group, err = b.db.CreateGroup(&model.Group{
			ChatID:           event.Message.Chat.ID,
			Title:            event.Message.Chat.Title,
			Description:      event.Message.Chat.Description,
			StandupDeadline:  "",
			TZ:               "Asia/Bishkek", // default value...
			OnbordingMessage: "",
			SubmissionDays:   "monday tuesday wednesday thirsday friday saturday sunday",
		})
		if err != nil {
			return err
		}
	}

	var msg tgbotapi.MessageConfig

	if group.StandupDeadline == "" {
		welcomeWithNoDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "welcomeNoDedline",
				Other: "Welcome to the standup team, no standup deadline has been setup yet",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, welcomeWithNoDeadline)
	} else {
		welcomeWithDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "welcomeWithDedline",
				Other: "Welcome to the standup team, please, submit your standups no later than {{.Deadline}}",
			},
			TemplateData: map[string]interface{}{
				"Deadline": group.StandupDeadline,
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, welcomeWithDeadline)
	}

	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//Show standupers
func (b *Bot) Show(event tgbotapi.Update) error {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)

	standupers, err := b.db.ListChatStandupers(event.Message.Chat.ID)
	if err != nil {
		return err
	}

	list := []string{}
	for _, standuper := range standupers {
		list = append(list, "@"+standuper.Username)
	}

	if len(list) == 0 {
		showNoStandupers, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "showNoStandupers",
				Other: "No standupers in the team, /join to start standuping",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, showNoStandupers)
		_, err = b.tgAPI.Send(msg)
		return err
	}

	showStandupers, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "showStandupers",
			One:   "Only {{.Standupers}} standups in the team, /join to start standuping",
			Two:   "{{.Standupers}} submit standups in the team",
			Few:   "{{.Standupers}} submit standups in the team",
			Many:  "{{.Standupers}} submit standups in the team",
			Other: "{{.Standupers}} submit standups in the team",
		},
		TemplateData: map[string]interface{}{
			"Standupers": strings.Join(list, ", "),
		},
		PluralCount: len(list),
	})
	if err != nil {
		log.Error(err)
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, showStandupers)
	_, err = b.tgAPI.Send(msg)
	if err != nil {
		log.Error(err)
	}

	group, err := b.db.FindGroup(event.Message.Chat.ID)
	if err != nil {
		group, err = b.db.CreateGroup(&model.Group{
			ChatID:          event.Message.Chat.ID,
			Title:           event.Message.Chat.Title,
			Description:     event.Message.Chat.Description,
			StandupDeadline: "",
			TZ:              "Asia/Bishkek", // default value...
			SubmissionDays:  "monday tuesday wednesday thirsday friday saturday sunday",
		})
		if err != nil {
			return err
		}
	}

	if group.StandupDeadline == "" {
		noStandupDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noStandupDeadline",
				Other: "Standup deadline is not set",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, noStandupDeadline)
	} else {
		standupDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "standupDeadline",
				Other: "Standup deadline set at {{.Deadline}} on {{.Weekdays}}",
			},
			TemplateData: map[string]interface{}{
				"Deadline": group.StandupDeadline,
				"Weekdays": group.SubmissionDays,
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, standupDeadline)
	}

	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//LeaveStandupers standupers
func (b *Bot) LeaveStandupers(event tgbotapi.Update) error {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)

	standuper, err := b.db.FindStanduper(event.Message.From.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
	if err != nil {
		notStanduper, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "notStanduper",
				Other: "You do not standup yet",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, notStanduper)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	err = b.db.DeleteStanduper(standuper.ID)
	if err != nil {
		log.Error("DeleteStanduper failed: ", err)
		failedLeaveStanupers, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedLeaveStanupers",
				Other: "Could not remove you from standup team",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedLeaveStanupers)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	leaveStanupers, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "leaveStanupers",
			Other: "You no longer have to submit standups, thanks for all your standups and messages",
		},
	})
	if err != nil {
		log.Error(err)
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, leaveStanupers)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//EditDeadline modifies standup time
func (b *Bot) EditDeadline(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	deadline := event.Message.CommandArguments()

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		log.Error("findTeam failed")
		return fmt.Errorf("failed to find sutable team for edit deadline")
	}

	localizer := i18n.NewLocalizer(b.bundle, team.Group.Language)

	if strings.TrimSpace(deadline) == "" {
		team.Group.StandupDeadline = ""

		_, err = b.db.UpdateGroup(team.Group)
		if err != nil {
			log.Error("Remove Deadline failed: ", err)
			failedRemoveStandupDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "failedRemoveStandupDeadline",
					Other: "Could not remove standup deadline",
				},
			})
			if err != nil {
				log.Error(err)
			}
			msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedRemoveStandupDeadline)
			msg.ReplyToMessageID = event.Message.MessageID
			_, err = b.tgAPI.Send(msg)
			return err
		}
		log.Error("Remove Deadline failed: ", err)
		removeStandupDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "removeStandupDeadline",
				Other: "Standup deadline removed",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, removeStandupDeadline)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	team.Group.StandupDeadline = deadline

	log.Info(team.Group)

	_, err = b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in EditDeadline failed: ", err)
		failedUpdateStandupDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateStandupDeadline",
				Other: "Could not edit standup deadline",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedUpdateStandupDeadline)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	updateStandupDeadline, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateStandupDeadline",
			Other: "Edited standup deadline, new deadline is {{.Deadline}}",
		},
		TemplateData: map[string]interface{}{
			"Deadline": deadline,
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, updateStandupDeadline)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

func (b *Bot) UpdateOnbordingMessage(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	onbordingMessage := event.Message.CommandArguments()

	log.Info("Onbording Message: ", onbordingMessage)

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		log.Error("findTeam failed")
		return fmt.Errorf("failed to find sutable team for edit deadline")
	}

	localizer := i18n.NewLocalizer(b.bundle, team.Group.Language)

	if strings.TrimSpace(onbordingMessage) == "" {
		team.Group.OnbordingMessage = ""

		_, err = b.db.UpdateGroup(team.Group)
		if err != nil {
			log.Error("Remove Deadline failed: ", err)
			failedRemoveOnbordingMessage, err := localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "failedRemoveOnbordingMessage",
					Other: "Could not remove remove onbording message",
				},
			})
			if err != nil {
				log.Error(err)
			}
			msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedRemoveOnbordingMessage)
			msg.ReplyToMessageID = event.Message.MessageID
			_, err = b.tgAPI.Send(msg)
			return err
		}
		log.Error("Remove Deadline failed: ", err)
		removeOnbordingMessage, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "removeOnbordingMessage",
				Other: "Standup deadline removed",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, removeOnbordingMessage)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	team.Group.OnbordingMessage = onbordingMessage

	log.Info(team.Group)

	group, err := b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in EditDeadline failed: ", err)
		failedUpdateOnbordingMessage, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateOnbordingMessage",
				Other: "Could not edit onbording message",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedUpdateOnbordingMessage)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	log.Info("Group after update: ", group)

	updateOnbordingMessage, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateOnbordingMessage",
			Other: "Onbording message updated",
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, updateOnbordingMessage)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

func (b *Bot) UpdateGroupLanguage(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	language := event.Message.CommandArguments()

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		log.Error("findTeam failed")
		return fmt.Errorf("failed to find sutable team for edit deadline")
	}

	localizer := i18n.NewLocalizer(b.bundle, language)

	team.Group.Language = language

	if strings.TrimSpace(language) == "" {
		team.Group.Language = "en"
	}

	log.Info(team.Group)

	group, err := b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in Change language failed: ", err)
		failedUpdateLanguage, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateLanguage",
				Other: "Could not edit group language",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedUpdateLanguage)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	log.Info("Group after update: ", group)

	updateGroupLanguage, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateGroupLanguage",
			Other: "Group language updated",
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, updateGroupLanguage)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//ChangeSubmissionDays changes days on which interns should submit standups
func (b *Bot) ChangeSubmissionDays(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	submissionDays := event.Message.CommandArguments()

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		log.Error("findTeam failed")
		return fmt.Errorf("failed to find sutable team for edit deadline")
	}

	localizer := i18n.NewLocalizer(b.bundle, team.Group.Language)

	team.Group.SubmissionDays = strings.ToLower(submissionDays)

	group, err := b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in Change language failed: ", err)
		failedUpdateSubmissionDays, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateSubmissionDays",
				Other: "Could not edit standup submission days",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedUpdateSubmissionDays)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	log.Info("Group after update: ", group)

	updateGroupSubmissionDays, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateGroupSubmissionDays",
			Other: "Group Standup submission days updated",
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, updateGroupSubmissionDays)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//ChangeGroupTimeZone modifies time zone of the group
func (b *Bot) ChangeGroupTimeZone(event tgbotapi.Update) error {

	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	tz := event.Message.CommandArguments()

	if strings.TrimSpace(tz) == "" {
		return nil
	}

	team := b.findTeam(event.Message.Chat.ID)
	log.Info("Current team: ", team)

	if team == nil {
		group, err := b.db.CreateGroup(&model.Group{
			ChatID:          event.Message.Chat.ID,
			Title:           event.Message.Chat.Title,
			Description:     event.Message.Chat.Description,
			StandupDeadline: "",
			TZ:              "Asia/Bishkek", // default value...
			SubmissionDays:  "monday tuesday wednesday thirsday friday saturday sunday",
		})
		if err != nil {
			return err
		}
		b.watchersChan <- group
		team = b.findTeam(event.Message.Chat.ID)
	}

	team.Group.TZ = tz

	localizer := i18n.NewLocalizer(b.bundle, team.Group.Language)

	log.Info("localizer ", localizer)

	_, err = time.LoadLocation(tz)
	if err != nil {
		log.Error("UpdateGroup in ChangeTimeZone failed: ", err)
		failedRecognizeTZ, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedRecognizeTZ",
				Other: "Failed to recognize new TZ you entered, double check the tz name and try again",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedRecognizeTZ)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in ChangeTimeZone failed: ", err)
		failedUpdateTZ, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateTZ",
				Other: "Failed to update Timezone",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedUpdateTZ)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	updateTZ, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateTZ",
			Other: "Group timezone is updated, new TZ is {{.TZ}}",
		},
		TemplateData: map[string]interface{}{
			"TZ": tz,
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, updateTZ)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//ChangeUserTimeZone assign user a different time zone
func (b *Bot) ChangeUserTimeZone(event tgbotapi.Update) error {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)

	tz := event.Message.CommandArguments()

	if strings.TrimSpace(tz) == "" {
		return nil
	}

	st, err := b.db.FindStanduper(event.Message.From.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
	if err != nil {
		notStanduper, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "notStanduper",
				Other: "You do not standup yet",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, notStanduper)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	st.TZ = tz

	_, err = time.LoadLocation(tz)
	if err != nil {
		log.Error("LoadLocation in ChangeUserTimeZone failed: ", err)
		failedRecognizeTZ, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedRecognizeTZ",
				Other: "Failed to recognize new TZ you entered, double check the tz name and try again",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedRecognizeTZ)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.UpdateStanduper(st)
	if err != nil {
		log.Error("UpdateStanduper in ChangeUserTimeZone failed: ", err)
		failedUpdateTZ, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateTZ",
				Other: "Failed to update Timezone",
			},
		})
		if err != nil {
			log.Error(err)
		}
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, failedUpdateTZ)
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	updateTZ, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateUserTZ",
			Other: "your timezone is updated, new TZ is {{.TZ}}",
		},
		TemplateData: map[string]interface{}{
			"TZ": tz,
		},
	})
	if err != nil {
		log.Error(err)
	}
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, updateTZ)
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

func (b *Bot) senderIsAdminInChannel(sendername string, chatID int64) (bool, error) {
	isAdmin := false
	chat := tgbotapi.ChatConfig{
		ChatID:             chatID,
		SuperGroupUsername: "",
	}
	admins, err := b.tgAPI.GetChatAdministrators(chat)
	if err != nil {
		return false, err
	}
	for _, admin := range admins {
		if admin.User.UserName == sendername {
			isAdmin = true
			return true, nil
		}
	}
	return isAdmin, nil
}
