package bot

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"

	"github.com/bouk/monkey"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/BurntSushi/toml"
	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/maddevsio/mad-telegram-standup-bot/model"
	"github.com/maddevsio/mad-telegram-standup-bot/storage"

	"github.com/stretchr/testify/require"
)

func TestPrepareShowMessage(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	bot := Bot{c: conf, bundle: bundle}
	require.NoError(t, err)

	group := &model.Group{
		StandupDeadline: "10:00",
		SubmissionDays:  "monday",
	}

	d := time.Date(2019, 7, 15, 0, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	text := bot.prepareShowMessage([]*model.Standuper{}, group)
	assert.Equal(t, "No standupers in the team, /join to start standuping\n\nStandup deadline set at 10:00 on monday", text)

	standuper := []*model.Standuper{
		&model.Standuper{
			Username: "foo",
			Created:  time.Date(2019, 7, 14, 0, 0, 0, 0, time.Local),
			Warnings: 0,
		},
	}

	text = bot.prepareShowMessage(standuper, group)
	assert.Equal(t, "Interns:\n@foo, 1 day on intership, missed standups: 0 times\n\nStandup deadline set at 10:00 on monday", text)

	standupers := []*model.Standuper{
		&model.Standuper{
			Username: "foo",
			Created:  time.Date(2019, 7, 14, 0, 0, 0, 0, time.Local),
			Warnings: 0,
		},

		&model.Standuper{
			Username: "bar",
			Created:  time.Date(2019, 7, 10, 0, 0, 0, 0, time.Local),
			Warnings: 2,
		},
	}

	text = bot.prepareShowMessage(standupers, group)
	assert.Equal(t, "Interns:\n@bar, 5 days on internship, missed standups: 2 times\n@foo, 1 day on intership, missed standups: 0 times\n\nStandup deadline set at 10:00 on monday", text)
}

func TestHelp(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)
	_, err = bundle.LoadMessageFile("../active.ru.toml")
	require.NoError(t, err)

	bot := Bot{c: conf, db: db, bundle: bundle}
	g := &model.Group{
		ChatID:   int64(11),
		Language: "en",
	}

	group, err := db.CreateGroup(g)
	require.NoError(t, err)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: group.ChatID,
			},
		},
	}

	helpMessage := "In order to submit a standup, tag me and write a message with keywords. Direct message me to see the list of keywords needed. Loking forward for your standups! Message @anatoliyfedorenko in case of any unexpected behaviour, submit issues to https://github.com/maddevsio/mad-telegram-standup-bot/issues"
	text, err := bot.Help(update)
	require.NoError(t, err)
	require.Equal(t, helpMessage, text)

	group, err = db.FindGroup(group.ChatID)
	require.NoError(t, err)
	g = &model.Group{
		ID:       group.ID,
		Language: "ru",
	}

	group, err = db.UpdateGroup(g)
	require.NoError(t, err)

	helpMessage = "Чтобы написать стендап тегни меня в сообщении с ключевыми словами. Напиши мне в личку текст стендапа, чтобы я сказал какие ключевые слова пропущены. Жду ваших стендапов! Напиши @anatoliyfedorenko в случае любых проблем связанных со мной, отправляйте запросы на доработку в этот репозиторий https://github.com/maddevsio/mad-telegram-standup-bot/issues"
	text, err = bot.Help(update)
	require.NoError(t, err)
	require.Equal(t, helpMessage, text)
}

func TestJoinLeaveShowCommands(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	bot := Bot{c: conf, db: db, bundle: bundle}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	text, err := bot.LeaveStandupers(update)
	assert.NoError(t, err)
	assert.Equal(t, "You do not standup yet", text)

	text, err = bot.JoinStandupers(update)
	assert.NoError(t, err)
	assert.Equal(t, "Welcome to the standup team, no standup deadline has been setup yet", text)

	text, err = bot.JoinStandupers(update)
	assert.NoError(t, err)
	assert.Equal(t, "You already a part of standup team", text)

	text, err = bot.LeaveStandupers(update)
	assert.NoError(t, err)
	assert.Equal(t, "You no longer have to submit standups, thanks for all your standups and messages", text)

	text, err = bot.Show(update)
	assert.NoError(t, err)
	assert.Equal(t, "No standupers in the team, /join to start standuping\n\nStandup deadline is not set", text)

	text, err = bot.JoinStandupers(update)
	assert.NoError(t, err)
	assert.Equal(t, "Welcome back! Glad to see you again, and looking forward to your standups", text)

	standuper, err := db.FindStanduper(1, 1)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteStanduper(standuper.ID))

	group, err := db.FindGroup(1)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteGroup(group.ID))
}

func TestDeadlines(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	wch := make(chan *model.Group)
	var teams []*model.Team

	bot := Bot{c: conf, db: db, bundle: bundle, watchersChan: wch, teams: teams}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	text, err := bot.EditDeadline(update)
	assert.NoError(t, err)
	assert.Equal(t, "Standup deadline removed", text)

	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 14,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/edit_deadline 12:00",
		},
	}

	text, err = bot.EditDeadline(update)
	assert.NoError(t, err)
	assert.Equal(t, "Edited standup deadline, new deadline is 12:00", text)

	group, err := db.FindGroup(1)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteGroup(group.ID))
}

func TestUpdateOnbordingMessage(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	wch := make(chan *model.Group)
	var teams []*model.Team

	bot := Bot{c: conf, db: db, bundle: bundle, watchersChan: wch, teams: teams}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	text, err := bot.UpdateOnbordingMessage(update)
	assert.NoError(t, err)
	assert.Equal(t, "Onbording message removed", text)

	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 24,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/update_onbording_message 12:00",
		},
	}

	text, err = bot.UpdateOnbordingMessage(update)
	assert.NoError(t, err)
	assert.Equal(t, "Onbording message updated", text)

	group, err := db.FindGroup(1)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteGroup(group.ID))
}

func TestUpdateGroupLanguage(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	wch := make(chan *model.Group)
	var teams []*model.Team

	bot := Bot{c: conf, db: db, bundle: bundle, watchersChan: wch, teams: teams}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	text, err := bot.UpdateGroupLanguage(update)
	assert.NoError(t, err)
	assert.Equal(t, "Group language updated", text)

	group, err := db.FindGroup(1)
	assert.NoError(t, err)
	assert.Equal(t, "en", group.Language)

	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 21,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/update_group_message ru",
		},
	}

	text, err = bot.UpdateGroupLanguage(update)
	assert.NoError(t, err)
	assert.Equal(t, "Group language updated", text)

	group, err = db.FindGroup(1)
	assert.NoError(t, err)
	assert.Equal(t, "ru", group.Language)

	assert.NoError(t, db.DeleteGroup(group.ID))
}

func TestChangeSubmissionDays(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	wch := make(chan *model.Group)
	var teams []*model.Team

	bot := Bot{c: conf, db: db, bundle: bundle, watchersChan: wch, teams: teams}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	text, err := bot.ChangeSubmissionDays(update)
	assert.NoError(t, err)
	assert.Equal(t, "Group Standup submission days updated", text)

	group, err := db.FindGroup(1)
	assert.NoError(t, err)
	assert.Equal(t, "", group.SubmissionDays)

	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 22,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/change_submission_days monday tuesday ",
		},
	}

	text, err = bot.ChangeSubmissionDays(update)
	assert.NoError(t, err)
	assert.Equal(t, "Group Standup submission days updated", text)

	group, err = db.FindGroup(1)
	assert.NoError(t, err)
	assert.Equal(t, "monday tuesday", group.SubmissionDays)

	assert.NoError(t, db.DeleteGroup(group.ID))
}

func TestChangeGroupTimeZone(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	wch := make(chan *model.Group)
	var teams []*model.Team

	bot := Bot{c: conf, db: db, bundle: bundle, watchersChan: wch, teams: teams}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	text, err := bot.ChangeGroupTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "Group timezone is updated, new TZ is Asia/Bishkek", text)

	group, err := db.FindGroup(1)
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Bishkek", group.TZ)

	//case ok
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 9,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/group_tz Asia/Almaty",
		},
	}

	text, err = bot.ChangeGroupTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "Group timezone is updated, new TZ is Asia/Almaty", text)

	group, err = db.FindGroup(update.Message.Chat.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Almaty", group.TZ)

	//case incorrect timezone
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 9,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/group_tz Foo",
		},
	}

	text, err = bot.ChangeGroupTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to recognize new TZ you entered, double check the tz name and try again", text)

	assert.NoError(t, db.DeleteGroup(group.ID))
}

func TestChangeUserTimeZone(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	wch := make(chan *model.Group)
	var teams []*model.Team

	bot := Bot{c: conf, db: db, bundle: bundle, watchersChan: wch, teams: teams}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
		},
	}

	_, err = bot.JoinStandupers(update)
	assert.NoError(t, err)

	text, err := bot.ChangeUserTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "your timezone is updated, new TZ is Asia/Bishkek", text)

	user, err := db.FindStanduper(update.Message.From.ID, update.Message.Chat.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Bishkek", user.TZ)

	//case ok
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 3,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/tz Asia/Tashkent",
		},
	}

	text, err = bot.ChangeUserTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "your timezone is updated, new TZ is Asia/Tashkent", text)

	user, err = db.FindStanduper(update.Message.From.ID, update.Message.Chat.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Tashkent", user.TZ)

	//case incorrect timezone
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 3,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/tz Foo",
		},
	}

	text, err = bot.ChangeUserTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to recognize new TZ you entered, double check the tz name and try again", text)

	//case not standuper
	assert.NoError(t, db.DeleteStanduper(user.ID))
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				tgbotapi.MessageEntity{
					Type:   "bot_command",
					Offset: 0,
					Length: 3,
				},
			},
			From: &tgbotapi.User{
				ID:           1,
				UserName:     "Foo",
				LanguageCode: "en",
			},
			Chat: &tgbotapi.Chat{
				ID:          1,
				Title:       "Foo chat",
				Description: "",
			},
			Text: "/tz Asia/Tashkent",
		},
	}

	text, err = bot.ChangeUserTimeZone(update)
	assert.NoError(t, err)
	assert.Equal(t, "You do not standup yet", text)
}

func TestLeaveStandupers(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)
	_, err = bundle.LoadMessageFile("../active.ru.toml")
	require.NoError(t, err)

	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)

	bot := Bot{c: conf, db: db, bundle: bundle}

	g := &model.Group{
		ChatID:   int64(17),
		Language: "en",
	}

	group, err := db.CreateGroup(g)
	require.NoError(t, err)

	team := &model.Team{
		Group:    group,
		QuitChan: make(chan struct{}),
	}
	bot.teams = append(bot.teams, team)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID: 2,
			},
			Chat: &tgbotapi.Chat{
				ID: group.ChatID,
			},
		},
	}

	text, err := bot.LeaveStandupers(update)
	require.NoError(t, err)
	assert.Equal(t, "You do not standup yet", text)

	g = &model.Group{
		ChatID:   int64(18),
		Language: "ru",
	}

	group, err = db.CreateGroup(g)
	require.NoError(t, err)

	team = &model.Team{
		Group:    group,
		QuitChan: make(chan struct{}),
	}
	bot.teams = append(bot.teams, team)

	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				ID: 3,
			},
			Chat: &tgbotapi.Chat{
				ID: group.ChatID,
			},
		},
	}

	text, err = bot.LeaveStandupers(update)
	require.NoError(t, err)
	assert.Equal(t, "Вы еще не стендапите", text)
}

func TestEditDeadlines(t *testing.T) {
	Test = true
	conf, err := config.Get()
	require.NoError(t, err)
	db, err := storage.NewMySQL(conf)
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)
	_, err = bundle.LoadMessageFile("../active.ru.toml")
	require.NoError(t, err)

	bot := Bot{c: conf, db: db, bundle: bundle}
	g := &model.Group{
		ChatID:   int64(11),
		Language: "en",
	}

	group, err := db.CreateGroup(g)
	require.NoError(t, err)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: group.ChatID,
			},
		},
	}

	infoMessage := "Standup deadline removed"
	text, err := bot.EditDeadline(update)
	require.NoError(t, err)
	require.Equal(t, infoMessage, text)

	group, err = db.FindGroup(group.ChatID)
	require.NoError(t, err)
	g = &model.Group{
		ID:       group.ID,
		Language: "ru",
	}

	group, err = db.UpdateGroup(g)
	require.NoError(t, err)

	infoMessage = "Крайний срок сдачи стендапов отменён"
	text, err = bot.EditDeadline(update)
	require.NoError(t, err)
	require.Equal(t, infoMessage, text)

}
