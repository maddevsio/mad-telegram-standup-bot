package storage

import (
	"github.com/maddevsio/mad-telegram-standup-bot/model"
)

// CreateNotification create notifications
func (m *MySQL) CreateNotification(s model.NotificationThread) (model.NotificationThread, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `notifications_thread` (group_id, user_id, notification_time, already_reminded) VALUES (?, ?, ?, ?)",
		s.GroupID, s.UserID, s.NotificationTime, s.AlreadyReminded,
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

// DeleteNotification deletes notification entry from database
func (m *MySQL) DeleteNotification(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `notifications_thread` WHERE id=?", id)
	return err
}

// ListNotifications returns array of notifications entries from database
func (m *MySQL) ListNotifications() ([]*model.NotificationThread, error) {
	items := []*model.NotificationThread{}
	err := m.conn.Select(&items, "SELECT * FROM `notifications_thread`")
	return items, err
}
