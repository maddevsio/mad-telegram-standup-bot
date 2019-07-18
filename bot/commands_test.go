package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/BurntSushi/toml"
	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/maddevsio/mad-internship-bot/storage"

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
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	bot := Bot{c: conf, bundle: bundle}
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				LanguageCode: "en",
			},
		},
	}

	helpMessage := "In order to submit a standup, tag me and write a message with keywords. Direct message me to see the list of keywords needed. Loking forward for your standups! Message @anatoliyfedorenko in case of any unexpected behaviour, submit issues to https://github.com/maddevsio/mad-internship-bot/issues"
	text, err := bot.Help(update)
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
