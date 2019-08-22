package storage

import (
	"testing"

	"github.com/maddevsio/mad-telegram-standup-bot/model"

	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/stretchr/testify/assert"
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
	}

	group, err := mysql.CreateGroup(g)
	assert.NoError(t, err)
	assert.Equal(t, "User", group.Username)
	assert.Equal(t, int64(15), group.ChatID)
	assert.Equal(t, "GMT +6", group.TZ)

	g.TZ = "Asia/Bishkek"

	_, err = mysql.UpdateGroup(group)
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Bishkek", group.TZ)

	group, err = mysql.SelectGroup(g.ID)
	assert.NoError(t, err)

	group, err = mysql.FindGroup(g.ChatID)
	assert.Equal(t, int64(15), g.ChatID)

	group2, err := mysql.CreateGroup(g)
	assert.NoError(t, err)

	groups, err := mysql.ListGroups()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(groups))

	err = mysql.DeleteGroup(group.ID)
	assert.NoError(t, err)

	err = mysql.DeleteGroup(group2.ID)
	assert.NoError(t, err)

	_, err = mysql.SelectGroup(group.ID)
	assert.Error(t, err)

	_, err = mysql.SelectGroup(group2.ID)
	assert.Error(t, err)
}
