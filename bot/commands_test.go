package bot

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/BurntSushi/toml"
	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/stretchr/testify/require"
)

func TestShow(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	bot, err := New(conf, bundle)
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
	assert.Equal(t, "Interns:\n@foo, 1 day on intership, missed standups: 0 times\n@bar, 5 days on internship, missed standups: 2 times\n\nStandup deadline set at 10:00 on monday", text)

}
