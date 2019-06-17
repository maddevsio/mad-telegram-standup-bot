package bot

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bouk/monkey"
	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/maddevsio/mad-internship-bot/storage"
	"github.com/stretchr/testify/require"
)

func TestSubmittedStandupToday(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	mysql, err := storage.NewMySQL(conf)
	require.NoError(t, err)
	bot, err := New(conf)
	require.NoError(t, err)

	d := time.Date(2019, 6, 17, 4, 20, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	standup, err := mysql.CreateStandup(&model.Standup{
		MessageID: 1,
		Username:  "Foo",
		ChatID:    int64(12345),
		Text:      "Yesterday, today, blockers test message",
	})
	require.NoError(t, err)

	logrus.Info(standup)

	d = time.Date(2019, 6, 17, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	submitted := bot.submittedStandupToday(&model.Standuper{
		Username: "Foo",
		ChatID:   int64(12345),
		TZ:       "Asia/Bishkek",
	})

	require.True(t, submitted)
}
