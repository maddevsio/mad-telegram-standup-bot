package storage

import (
	"github.com/maddevsio/mad-telegram-standup-bot/model"
	"time"
)

// CreateStandup creates standup entry in database
func (m *MySQL) CreateStandup(s *model.Standup) (*model.Standup, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standups` (message_id, created, modified, username, text, chat_id) VALUES (?, ?, ?, ?, ?, ?)",
		s.MessageID, time.Now().UTC(), time.Now().UTC(), s.Username, s.Text, s.ChatID,
	)
	if err != nil {
		return s, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return s, err
	}
	standup, err := m.SelectStandup(id)
	if err != nil {
		return s, err
	}
	return standup, nil
}

// UpdateStandup updates standup entry in database
func (m *MySQL) UpdateStandup(s *model.Standup) (*model.Standup, error) {
	standup := &model.Standup{}
	_, err := m.conn.Exec(
		"UPDATE `standups` SET modified=?, text=? WHERE id=?",
		time.Now().UTC(), s.Text, s.ID,
	)
	if err != nil {
		return nil, err
	}

	standup, err = m.SelectStandup(s.ID)
	if err != nil {
		return nil, err
	}
	return standup, err
}

// SelectStandup selects standup entry from database
func (m *MySQL) SelectStandup(id int64) (*model.Standup, error) {
	s := &model.Standup{}
	err := m.conn.Get(s, "SELECT * FROM `standups` WHERE id=?", id)
	return s, err
}

// SelectStandupByMessageID selects standup entry from database
func (m *MySQL) SelectStandupByMessageID(messageID int, chatID int64) (*model.Standup, error) {
	s := &model.Standup{}
	err := m.conn.Get(s, "SELECT * FROM `standups` WHERE message_id=? and chat_id=?", messageID, chatID)
	return s, err
}

// DeleteStandup deletes standup entry from database
func (m *MySQL) DeleteStandup(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `standups` WHERE id=?", id)
	return err
}

// ListStandups returns array of standup entries from database
func (m *MySQL) ListStandups() ([]*model.Standup, error) {
	items := []*model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standups`")
	return items, err
}

//LastStandupFor returns last standup for Standuper
func (m *MySQL) LastStandupFor(username string, chatID int64) (*model.Standup, error) {
	standup := &model.Standup{}
	err := m.conn.Get(standup, "SELECT * FROM `standups` WHERE username=? and chat_id=? ORDER BY id DESC LIMIT 1", username, chatID)
	return standup, err
}
