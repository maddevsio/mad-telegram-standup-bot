package storage

import (
	"testing"
	"time"

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
		Username:         "User1",
		NotificationTime: time.Now(),
		ReminderCounter:  0,
	}

	timeTest := n.NotificationTime

	notification, err := mysql.CreateNotificationThread(n)
	require.NoError(t, err)
	assert.Equal(t, int64(1), notification.ChatID)
	assert.Equal(t, "User1", notification.Username)
	assert.Equal(t, timeTest, notification.NotificationTime)
	assert.Equal(t, 0, notification.ReminderCounter)

	notification2, err := mysql.CreateNotificationThread(n)
	require.NoError(t, err)

	notifications, err := mysql.ListNotificationsThread(notification2.ChatID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(notifications))

	err = mysql.DeleteNotificationThread(notification2.ID)
	require.NoError(t, err)

	err = mysql.DeleteNotificationThread(notification.ID)
	require.NoError(t, err)

	n = model.NotificationThread{
		ChatID:           int64(1),
		Username:         "User2",
		NotificationTime: time.Now(),
		ReminderCounter:  0,
	}

	nt, err := mysql.CreateNotificationThread(n)
	require.NoError(t, err)

	err = mysql.UpdateNotificationThread(nt.ID, nt.ChatID, time.Now())
	require.NoError(t, err)

	notifications, err = mysql.ListNotificationsThread(nt.ChatID)
	require.NoError(t, err)
	for _, thread := range notifications {
		assert.Equal(t, 1, thread.ReminderCounter)
	}

	err = mysql.DeleteNotificationThread(nt.ID)
	require.NoError(t, err)
}
