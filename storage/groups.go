package storage

import (
	"github.com/maddevsio/mad-internship-bot/model"
)

// CreateGroup creates Group
func (m *MySQL) CreateGroup(group *model.Group) (*model.Group, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `groups` (chat_id, title, username, description, standup_deadline, tz, language, onbording_message, submission_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		group.ChatID, group.Title, group.Username, group.Description, group.StandupDeadline, group.TZ, group.Language, group.OnbordingMessage, group.SubmissionDays,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	group.ID = id
	return group, nil
}

// UpdateGroup updates Group entry in database
func (m *MySQL) UpdateGroup(group *model.Group) (*model.Group, error) {
	_, err := m.conn.Exec(
		"UPDATE `groups` SET title=?, tz=?, language=?, username=?, description=?, standup_deadline=?, onbording_message=?, submission_days=? WHERE id=?",
		group.Title, group.TZ, group.Language, group.Username, group.Description, group.StandupDeadline, group.OnbordingMessage, group.SubmissionDays, group.ID,
	)
	if err != nil {
		return group, err
	}
	err = m.conn.Get(group, "SELECT * FROM `groups` WHERE id=?", group.ID)
	return group, err
}

// SelectGroup selects Group entry from database
func (m *MySQL) SelectGroup(id int64) (*model.Group, error) {
	group := &model.Group{}
	err := m.conn.Get(group, "SELECT * FROM `groups` WHERE id=?", id)
	return group, err
}

// FindGroup selects Group entry from database
func (m *MySQL) FindGroup(chatID int64) (*model.Group, error) {
	group := &model.Group{}
	err := m.conn.Get(group, "SELECT * FROM `groups` WHERE chat_id=?", chatID)
	return group, err
}

// ListGroups returns array of Group entries from database filtered by chat
func (m *MySQL) ListGroups() ([]*model.Group, error) {
	groups := []*model.Group{}
	err := m.conn.Select(&groups, "SELECT * FROM `groups`")
	return groups, err
}

// DeleteGroup deletes Group entry from database
func (m *MySQL) DeleteGroup(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `groups` WHERE id=?", id)
	return err
}
