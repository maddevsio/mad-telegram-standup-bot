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

type internInfo struct {
	internName       string
	timeSinceAdded   string
	missedStandups   string
	daysOnInternship int
}

// Test helps to use some functions without need to use external APIs
var Test bool

//HandleCommand handles imcomming commands
func (b *Bot) HandleCommand(event tgbotapi.Update) (err error) {
	var message string
	switch event.Message.Command() {
	case "help":
		message, err = b.Help(event)
		if err != nil {
			log.Error("Help failed: ", err)
		}
	case "join":
		message, err = b.JoinStandupers(event)
		if err != nil {
			log.Error("JoinStandupers failed: ", err)
		}
	case "show":
		message, err = b.Show(event)
		if err != nil {
			log.Error("Show failed: ", err)
		}
	case "leave":
		message, err = b.LeaveStandupers(event)
		if err != nil {
			log.Error("LeaveStandupers failed: ", err)
		}
	case "edit_deadline":
		message, err = b.EditDeadline(event)
		if err != nil {
			log.Error("EditDeadline failed: ", err)
		}
	case "update_onbording_message":
		message, err = b.UpdateOnbordingMessage(event)
		if err != nil {
			log.Error("UpdateOnbordingMessage failed: ", err)
		}
	case "update_group_language":
		message, err = b.UpdateGroupLanguage(event)
		if err != nil {
			log.Error("UpdateGroupLanguage failed: ", err)
		}
	case "change_submission_days":
		message, err = b.ChangeSubmissionDays(event)
		if err != nil {
			log.Error("ChangeSubmissionDays failed: ", err)
		}
	case "group_tz":
		message, err = b.ChangeGroupTimeZone(event)
		if err != nil {
			log.Error("ChangeGroupTimeZone failed: ", err)
		}
	case "tz":
		message, err = b.ChangeUserTimeZone(event)
		if err != nil {
			log.Error("ChangeUserTimeZone failed: ", err)
		}
	default:
		message = "I do not know this command..."
	}

	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, message)
	_, err = b.tgAPI.Send(msg)
	return err
}

//Help displays help message
func (b *Bot) Help(event tgbotapi.Update) (string, error) {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)
	helpText, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "helpText",
			Other: `In order to submit a standup, tag me and write a message with keywords. Direct message me to see the list of keywords needed. Loking forward for your standups! Message @anatoliyfedorenko in case of any unexpected behaviour, submit issues to https://github.com/maddevsio/mad-internship-bot/issues`,
		},
	})
	return helpText, err
}

//JoinStandupers assign user a standuper role
func (b *Bot) JoinStandupers(event tgbotapi.Update) (string, error) {
	localizer := i18n.NewLocalizer(b.bundle, event.Message.From.LanguageCode)
	standuper, err := b.db.FindStanduper(event.Message.From.ID, event.Message.Chat.ID) // user[1:] to remove leading @
	if err == nil {
		switch standuper.Status {
		case "active":
			return localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "youAlreadyStandup",
					Other: "You already a part of standup team",
				},
			})

		case "paused", "deleted":
			standuper.Status = "active"
			_, err := b.db.UpdateStanduper(standuper)
			if err != nil {
				return "", err
			}

			return localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "welcomeBack",
					Other: "Welcome back! Glad to see you again, and looking forward to your standups",
				},
			})
		}
	}

	_, err = b.db.CreateStanduper(&model.Standuper{
		Created:      time.Now(),
		Status:       "active",
		UserID:       event.Message.From.ID,
		Username:     event.Message.From.UserName,
		ChatID:       event.Message.Chat.ID,
		LanguageCode: event.Message.From.LanguageCode,
		TZ:           "Asia/Bishkek", // default value...
	})
	if err != nil {
		log.Error("CreateStanduper failed: ", err)
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "createStanduperFailed",
				Other: "Could not add you to standup team",
			},
		})
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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	if group.StandupDeadline == "" {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "welcomeNoDedline",
				Other: "Welcome to the standup team, no standup deadline has been setup yet",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "welcomeWithDedline",
			Other: "Welcome to the standup team, please, submit your standups no later than {{.Deadline}}",
		},
		TemplateData: map[string]interface{}{
			"Deadline": group.StandupDeadline,
		},
	})
}

//Show standupers
func (b *Bot) Show(event tgbotapi.Update) (string, error) {

	standupers, err := b.db.ListChatStandupers(event.Message.Chat.ID)
	if err != nil {
		return "", err
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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	message := b.prepareShowMessage(standupers, group)

	return message, nil
}

func (b *Bot) prepareShowMessage(standupers []*model.Standuper, group *model.Group) string {

	localizer := i18n.NewLocalizer(b.bundle, group.Language)

	var internsInfo string
	var interns []internInfo

	for _, standuper := range standupers {
		var info internInfo

		info.internName = "@" + standuper.Username + ", "
		if standuper.Username == "" {
			info.internName = fmt.Sprintf("[stranger](tg://user?id=%v)", standuper.UserID)
		}

		daysOnInternship := time.Now().UTC().Sub(standuper.Created).Hours() / 24
		internshipDuration, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "internshipDuration",
				One:   "{{.Days}} day on intership, ",
				Two:   "{{.Days}} days on internship, ",
				Few:   "{{.Days}} days on internship, ",
				Many:  "{{.Days}} days on internship, ",
				Other: "{{.Days}} days on internship, ",
			},
			PluralCount: int(daysOnInternship),
			TemplateData: map[string]interface{}{
				"Days": int(daysOnInternship),
			},
		})
		if err != nil {
			log.Error(err)
		}

		info.timeSinceAdded = internshipDuration

		missedStandups, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "missedStandups",
				One:   "missed only deadline",
				Two:   "missed standups: {{.missedTimes}} times",
				Few:   "missed standups: {{.missedTimes}} times",
				Many:  "missed standups: {{.missedTimes}} times",
				Other: "missed standups: {{.missedTimes}} times",
			},
			PluralCount: standuper.Warnings,
			TemplateData: map[string]interface{}{
				"missedTimes": standuper.Warnings,
			},
		})
		if err != nil {
			log.Error(err)
		}

		info.missedStandups = missedStandups
		info.daysOnInternship = int(daysOnInternship)

		interns = append(interns, info)
	}

	interns = sortInterns(interns)

	for _, info := range interns {
		internsInfo += info.internName + info.timeSinceAdded + info.missedStandups + "\n"
	}

	showStandupers, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "showStandupers",
			Other: "Interns:",
		},
	})
	if err != nil {
		log.Error(err)
	}

	if len(internsInfo) == 0 {
		showStandupers, err = localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "showNoStandupers",
				Other: "No standupers in the team, /join to start standuping",
			},
		})
		if err != nil {
			log.Error(err)
		}
	}

	standupersInfo := showStandupers + "\n" + internsInfo

	var standupDeadlineInfo string

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
		standupDeadlineInfo = noStandupDeadline
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
		standupDeadlineInfo = standupDeadline
	}

	return standupersInfo + "\n" + standupDeadlineInfo
}

func sortInterns(entries []internInfo) []internInfo {
	var members []internInfo

	for i := 0; i < len(entries); i++ {
		if !sweep(entries, i) {
			break
		}
	}

	for _, item := range entries {
		members = append(members, item)
	}

	return members
}

func sweep(entries []internInfo, prevPasses int) bool {
	var N = len(entries)
	var didSwap = false
	var firstIndex = 0
	var secondIndex = 1

	for secondIndex < (N - prevPasses) {

		var firstItem = entries[firstIndex]
		var secondItem = entries[secondIndex]
		if entries[firstIndex].daysOnInternship < entries[secondIndex].daysOnInternship {
			entries[firstIndex] = secondItem
			entries[secondIndex] = firstItem
			didSwap = true
		}
		firstIndex++
		secondIndex++
	}

	return didSwap
}

//LeaveStandupers standupers
func (b *Bot) LeaveStandupers(event tgbotapi.Update) (string, error) {
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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	team := b.findTeam(group.ChatID)
	if team == nil {
		return "", fmt.Errorf("team %v not found", group.ChatID)
	}

	localizer := i18n.NewLocalizer(b.bundle, team.Group.Language)

	standuper, err := b.db.FindStanduper(event.Message.From.ID, event.Message.Chat.ID) // user[1:] to remove leading @
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "notStanduper",
				Other: "You do not standup yet",
			},
		})
	}

	standuper.Status = "paused"

	_, err = b.db.UpdateStanduper(standuper)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedLeaveStanupers",
				Other: "Could not remove you from standup team",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "leaveStanupers",
			Other: "You no longer have to submit standups, thanks for all your standups and messages",
		},
	})
}

//EditDeadline modifies standup time
func (b *Bot) EditDeadline(event tgbotapi.Update) (string, error) {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		return "", err
	}

	if !isAdmin {
		return "", fmt.Errorf("user not admin")
	}

	deadline := event.Message.CommandArguments()

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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	team := b.findTeam(group.ChatID)
	if team == nil {
		return "", fmt.Errorf("team %v not found", group.ChatID)
	}

	localizer := i18n.NewLocalizer(b.bundle, team.Group.Language)

	if strings.TrimSpace(deadline) == "" {
		team.Group.StandupDeadline = ""

		_, err = b.db.UpdateGroup(team.Group)
		if err != nil {
			return localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "failedRemoveStandupDeadline",
					Other: "Could not remove standup deadline",
				},
			})
		}

		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "removeStandupDeadline",
				Other: "Standup deadline removed",
			},
		})
	}

	team.Group.StandupDeadline = deadline

	_, err = b.db.UpdateGroup(team.Group)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateStandupDeadline",
				Other: "Could not edit standup deadline",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateStandupDeadline",
			Other: "Edited standup deadline, new deadline is {{.Deadline}}",
		},
		TemplateData: map[string]interface{}{
			"Deadline": deadline,
		},
	})
}

//UpdateOnbordingMessage updates welcoming message for the group
func (b *Bot) UpdateOnbordingMessage(event tgbotapi.Update) (string, error) {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		return "", err
	}

	if !isAdmin {
		return "", fmt.Errorf("user not admin")
	}

	onbordingMessage := event.Message.CommandArguments()

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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	localizer := i18n.NewLocalizer(b.bundle, group.Language)

	if strings.TrimSpace(onbordingMessage) == "" {
		group.OnbordingMessage = ""

		_, err = b.db.UpdateGroup(group)
		if err != nil {
			return localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "failedRemoveOnbordingMessage",
					Other: "Could not remove remove onbording message",
				},
			})
		}
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "removeOnbordingMessage",
				Other: "Onbording message removed",
			},
		})
	}

	group.OnbordingMessage = onbordingMessage

	_, err = b.db.UpdateGroup(group)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateOnbordingMessage",
				Other: "Could not edit onbording message",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateOnbordingMessage",
			Other: "Onbording message updated",
		},
	})
}

//UpdateGroupLanguage updates primary language for the group
func (b *Bot) UpdateGroupLanguage(event tgbotapi.Update) (string, error) {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		return "", err
	}

	if !isAdmin {
		return "", fmt.Errorf("user not admin")
	}

	language := event.Message.CommandArguments()

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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	localizer := i18n.NewLocalizer(b.bundle, language)

	group.Language = language

	if strings.TrimSpace(language) == "" {
		group.Language = "en"
	}

	_, err = b.db.UpdateGroup(group)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateLanguage",
				Other: "Could not edit group language",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateGroupLanguage",
			Other: "Group language updated",
		},
	})
}

//ChangeSubmissionDays changes days on which interns should submit standups
func (b *Bot) ChangeSubmissionDays(event tgbotapi.Update) (string, error) {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		return "", err
	}

	if !isAdmin {
		return "", fmt.Errorf("user not admin")
	}

	submissionDays := strings.TrimSpace(event.Message.CommandArguments())

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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	localizer := i18n.NewLocalizer(b.bundle, group.Language)

	group.SubmissionDays = strings.ToLower(submissionDays)

	_, err = b.db.UpdateGroup(group)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateSubmissionDays",
				Other: "Could not edit standup submission days",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateGroupSubmissionDays",
			Other: "Group Standup submission days updated",
		},
	})
}

//ChangeGroupTimeZone modifies time zone of the group
func (b *Bot) ChangeGroupTimeZone(event tgbotapi.Update) (string, error) {

	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		return "", err
	}

	if !isAdmin {
		return "", fmt.Errorf("user not admin")
	}

	tz := event.Message.CommandArguments()

	if strings.TrimSpace(tz) == "" {
		tz = "Asia/Bishkek"
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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	group.TZ = tz

	localizer := i18n.NewLocalizer(b.bundle, group.Language)
	_, err = time.LoadLocation(tz)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedRecognizeTZ",
				Other: "Failed to recognize new TZ you entered, double check the tz name and try again",
			},
		})
	}

	_, err = b.db.UpdateGroup(group)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateTZ",
				Other: "Failed to update Timezone",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateTZ",
			Other: "Group timezone is updated, new TZ is {{.TZ}}",
		},
		TemplateData: map[string]interface{}{
			"TZ": tz,
		},
	})
}

//ChangeUserTimeZone assign user a different time zone
func (b *Bot) ChangeUserTimeZone(event tgbotapi.Update) (string, error) {
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
			return "", err
		}

		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
	}

	localizer := i18n.NewLocalizer(b.bundle, group.Language)

	tz := event.Message.CommandArguments()

	if strings.TrimSpace(tz) == "" {
		tz = "Asia/Bishkek"
	}

	st, err := b.db.FindStanduper(event.Message.From.ID, event.Message.Chat.ID) // user[1:] to remove leading @
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "notStanduper",
				Other: "You do not standup yet",
			},
		})
	}

	st.TZ = tz

	_, err = time.LoadLocation(tz)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedRecognizeTZ",
				Other: "Failed to recognize new TZ you entered, double check the tz name and try again",
			},
		})
	}

	_, err = b.db.UpdateStanduper(st)
	if err != nil {
		return localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateTZ",
				Other: "Failed to update Timezone",
			},
		})
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateUserTZ",
			Other: "your timezone is updated, new TZ is {{.TZ}}",
		},
		TemplateData: map[string]interface{}{
			"TZ": tz,
		},
	})
}

func (b *Bot) senderIsAdminInChannel(sendername string, chatID int64) (bool, error) {
	if Test {
		return true, nil
	}
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
