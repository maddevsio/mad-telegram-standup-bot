package storage

import (
	"testing"

	"github.com/maddevsio/mad-telegram-standup-bot/model"

	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/stretchr/testify/require"
)

func TestGroups(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	mysql, err := NewMySQL(conf)
	require.NoError(t, err)

	g := &model.Group{
		ChatID:           int64(15),
		Title:            "Chat",
		Username:         "User",
		Description:      "foo",
		TZ:               "GMT +6",
		StandupDeadline:  "10:00",
		Language:         "en",
		OnbordingMessage: "Hello, user",
		SubmissionDays:   "Everyday",
		Advises:          "No standup bypass",
	}

	group, err := mysql.CreateGroup(g)
	require.NoError(t, err)
	require.Equal(t, "User", group.Username)
	require.Equal(t, int64(15), group.ChatID)
	require.Equal(t, "GMT +6", group.TZ)

	g.TZ = "Asia/Bishkek"

	_, err = mysql.UpdateGroup(group)
	require.NoError(t, err)
	require.Equal(t, "Asia/Bishkek", group.TZ)

	group, err = mysql.SelectGroup(g.ID)
	require.NoError(t, err)

	group, err = mysql.FindGroup(g.ChatID)
	require.Equal(t, int64(15), g.ChatID)

	group2, err := mysql.CreateGroup(g)
	require.NoError(t, err)

	groups, err := mysql.ListGroups()
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))

	err = mysql.DeleteGroup(group.ID)
	require.NoError(t, err)

	err = mysql.DeleteGroup(group2.ID)
	require.NoError(t, err)

	_, err = mysql.SelectGroup(group.ID)
	require.Error(t, err)

	_, err = mysql.SelectGroup(group2.ID)
	require.Error(t, err)
}
