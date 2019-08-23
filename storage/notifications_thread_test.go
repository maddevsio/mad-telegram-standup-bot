package storage

import (
	"testing"

	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/maddevsio/mad-telegram-standup-bot/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotification(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	mysql, err := NewMySQL(conf)
	require.NoError(t, err)

	n := model.NotificationThread{
		ChatID:           int64(1),
		UserID:           (1),
		NotificationTime: "10",
		ReminderCounter:  0,
	}

	notification, err := mysql.CreateNotification(n)
	require.NoError(t, err)
	assert.Equal(t, int64(1), notification.ChatID)
	assert.Equal(t, 1, notification.UserID)
	assert.Equal(t, "10", notification.NotificationTime)
	assert.Equal(t, 0, notification.ReminderCounter)

	notification2, err := mysql.CreateNotification(n)
	require.NoError(t, err)

	notifications, err := mysql.ListNotifications()
	require.NoError(t, err)
	assert.Equal(t, 2, len(notifications))

	err = mysql.DeleteNotification(notification2.ID)
	require.NoError(t, err)

	err = mysql.DeleteNotification(notification.ID)
	require.NoError(t, err)

	notifications, err = mysql.ListNotifications()
	require.NoError(t, err)
	assert.Equal(t, 0, len(notifications))
}
