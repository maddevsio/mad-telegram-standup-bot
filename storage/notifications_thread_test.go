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
		UserID:           (1),
		NotificationTime: time.Now(),
		ReminderCounter:  0,
	}

	timeTest := n.NotificationTime

	notification, err := mysql.CreateNotificationThread(n)
	require.NoError(t, err)
	assert.Equal(t, int64(1), notification.ChatID)
	assert.Equal(t, 1, notification.UserID)
	assert.Equal(t, timeTest, notification.NotificationTime)
	assert.Equal(t, 0, notification.ReminderCounter)

	notification2, err := mysql.CreateNotificationThread(n)
	require.NoError(t, err)

	notifications, err := mysql.ListNotificationsThread()
	require.NoError(t, err)
	assert.Equal(t, 2, len(notifications))

	err = mysql.DeleteNotificationThread(notification2.ID)
	require.NoError(t, err)

	err = mysql.DeleteNotificationThread(notification.ID)
	require.NoError(t, err)

	notifications, err = mysql.ListNotificationsThread()
	require.NoError(t, err)
	assert.Equal(t, 0, len(notifications))

	n = model.NotificationThread{
		ChatID:           int64(1),
		UserID:           1,
		NotificationTime: time.Now(),
		ReminderCounter:  0,
	}

	nt, err := mysql.CreateNotificationThread(n)
	require.NoError(t, err)

	nThread, err := mysql.SelectNotificationThread(1, int64(1))
	require.NoError(t, err)
	assert.Equal(t, nThread.ID, nt.ID)

	err = mysql.UpdateNotificationThread(nThread.ID)
	require.NoError(t, err)

	nThread, err = mysql.SelectNotificationThread(1, int64(1))
	assert.Equal(t, nThread.ReminderCounter, 1)

	err = mysql.UpdateNotificationThread(nThread.ID)
	require.NoError(t, err)

	nThread, err = mysql.SelectNotificationThread(1, int64(1))
	assert.Equal(t, nThread.ReminderCounter, 2)

	err = mysql.DeleteNotificationThread(nThread.ID)
	require.NoError(t, err)

}
