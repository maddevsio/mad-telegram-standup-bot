package storage

import (
	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"

	"github.com/maddevsio/mad-internship-bot/model"
)

// CreateStanduper creates Standuper
func (m *MySQL) CreateStanduper(s *model.Standuper) (*model.Standuper, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standupers` (created, user_id, username, chat_id, language_code, warnings, tz) VALUES (?, ?, ?, ?, ?, ?, ?)",
		s.Created, s.UserID, s.Username, s.ChatID, s.LanguageCode, 0, s.TZ,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	s.ID = id
	return s, nil
}

// UpdateStanduper updates Standuper entry in database
func (m *MySQL) UpdateStanduper(s *model.Standuper) (*model.Standuper, error) {
	m.conn.Exec(
		"UPDATE `standupers` SET username=?, language_code=?, warnings=?, tz=? WHERE id=?",
		s.Username, s.LanguageCode, s.Warnings, s.TZ, s.ID,
	)
	err := m.conn.Get(s, "SELECT * FROM `standupers` WHERE id=?", s.ID)
	return s, err
}

// SelectStanduper selects Standuper entry from database
func (m *MySQL) SelectStanduper(id int64) (*model.Standuper, error) {
	s := &model.Standuper{}
	err := m.conn.Get(s, "SELECT * FROM `standupers` WHERE id=?", id)
	return s, err
}

// FindStanduper selects Standuper entry from database
func (m *MySQL) FindStanduper(name string, chatID int64) (*model.Standuper, error) {
	s := &model.Standuper{}
	err := m.conn.Get(s, "SELECT * FROM `standupers` WHERE username=? and chat_id=?", name, chatID)
	return s, err
}

// ListChatStandupers returns array of Standuper entries from database filtered by chat
func (m *MySQL) ListChatStandupers(chatID int64) ([]*model.Standuper, error) {
	standupers := []*model.Standuper{}
	err := m.conn.Select(&standupers, "SELECT * FROM `standupers` where chat_id=?", chatID)
	return standupers, err
}

// DeleteStanduper deletes Standuper entry from database
func (m *MySQL) DeleteStanduper(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `standupers` WHERE id=?", id)
	return err
}

// DeleteGroupStandupers deletes Standuper entry from database
func (m *MySQL) DeleteGroupStandupers(chatID int64) error {
	_, err := m.conn.Exec("DELETE FROM `standupers` WHERE chat_id=?", chatID)
	return err
}
