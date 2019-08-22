package storage

import (
	"testing"

	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/maddevsio/mad-telegram-standup-bot/model"
	"github.com/stretchr/testify/require"
)

func TestNotification(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	mysql, err := NewMySQL(conf)
	require.NoError(t, err)

	n := model.NotificationThread{
		GroupID:          int64(1),
		UserID:           (1),
		NotificationTime: "10",
		AlreadyReminded:  0,
	}

	notification, err := mysql.CreateNotification(n)
	require.NoError(t, err)
	require.Equal(t, int64(1), notification.GroupID)
	require.Equal(t, 1, notification.UserID)
	require.Equal(t, "10", notification.NotificationTime)
	require.Equal(t, 0, notification.AlreadyReminded)

	notification2, err := mysql.CreateNotification(n)
	require.NoError(t, err)

	notifications, err := mysql.ListNotifications()
	require.NoError(t, err)
	require.Equal(t, 2, len(notifications))

	err = mysql.DeleteNotification(notification2.ID)
	require.NoError(t, err)

	err = mysql.DeleteNotification(notification.ID)
	require.NoError(t, err)

	notifications, err = mysql.ListNotifications()
	require.NoError(t, err)
	require.Equal(t, 0, len(notifications))
}
