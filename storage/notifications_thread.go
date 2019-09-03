package storage

import (
	"time"

	"github.com/maddevsio/mad-telegram-standup-bot/model"
)

// CreateNotificationThread create notifications
func (m *MySQL) CreateNotificationThread(s model.NotificationThread) (model.NotificationThread, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `notifications_thread` (chat_id, username, notification_time, reminder_counter) VALUES (?, ?, ?, ?)",
		s.ChatID, s.Username, s.NotificationTime, s.ReminderCounter,
	)
	if err != nil {
		return s, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return s, err
	}
	s.ID = id
	return s, nil
}

// DeleteNotificationThread deletes notification entry from database
func (m *MySQL) DeleteNotificationThread(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `notifications_thread` WHERE id=?", id)
	return err
}

// ListNotificationsThread returns array of notifications entries from database
func (m *MySQL) ListNotificationsThread(chatID int64) ([]model.NotificationThread, error) {
	items := []model.NotificationThread{}
	err := m.conn.Select(&items, "SELECT * FROM `notifications_thread` WHERE chat_id= ?", chatID)
	return items, err
}

// UpdateNotificationThread update field reminder counter
func (m *MySQL) UpdateNotificationThread(id int64, chatID int64, t time.Time) error {
	_, err := m.conn.Exec("UPDATE `notifications_thread` SET reminder_counter=reminder_counter+1, notification_time=? WHERE id=? AND chat_id=?", t, id, chatID)
	return err
}
