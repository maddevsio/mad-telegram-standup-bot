package bot

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"

	"github.com/bouk/monkey"
	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/maddevsio/mad-internship-bot/storage"
	"github.com/stretchr/testify/require"
)

func TestSubmittedStandupToday(t *testing.T) {
	t.Skip("need to configure connection to DB properly")
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

func TestStringReplace(t *testing.T) {
	text := `"test "`
	new := strings.Replace(text, `"`, "", -1)
	new = strings.TrimSpace(new)
	assert.Equal(t, "test", new)
}

func TestAnalyzeStandup(t *testing.T) {
	testCases := []struct {
		points int
		text   string
	}{
		{0, "yesterday, today, blockers"},
		{1, "@comedian yesterday, today, blockers"},
		{3, "@comedian yesterday, today @bot can you help me?, blockers"},
	}

	for _, tc := range testCases {
		_, points := analyzeStandup(tc.text)
		assert.Equal(t, tc.points, points)
	}
}

func TestContainsProblems(t *testing.T) {
	testCases := []struct {
		text        string
		result      bool
		bonusPoints int
	}{
		{"*Проблемы* много разных, раз два три 4 5", true, 8},
		{"*Проблемы* много разных, раз два три 4 5 проблем", true, 9},
	}

	for _, tc := range testCases {
		ok, points := containsProblems(tc.text)
		assert.Equal(t, tc.result, ok)
		assert.Equal(t, tc.bonusPoints, points)
	}

}
