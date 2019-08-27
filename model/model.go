package model

import (
	"errors"
	"strings"
	"time"
)

//Group represents separate chat that bot was added to to handle standups
type Group struct {
	ID               int64
	ChatID           int64  `db:"chat_id" json:"chat_id,omitempty"`
	Title            string `db:"title" json:"title"`
	Username         string `db:"username" json:"username"`
	Description      string `db:"description" json:"description,omitempty"`
	TZ               string `db:"tz" json:"tz"`
	Language         string `db:"language" json:"language"`
	StandupDeadline  string `db:"standup_deadline" json:"standup_deadline,omitempty"`
	OnbordingMessage string `db:"onbording_message" json:"onbording_message,omitempty"`
	SubmissionDays   string `db:"submission_days" json:"submission_days,omitempty"`
}

//Team is a helper struct to watch after different channels deadlines
type Team struct {
	Group    *Group
	QuitChan chan struct{}
}

//Stop finish tracking group
func (t *Team) Stop() {
	close(t.QuitChan)
}

// Standuper rerpesents standuper
type Standuper struct {
	ID           int64     `db:"id" json:"id"`
	Created      time.Time `db:"created" json:"created"`
	Status       string    `db:"status" json:"status"`
	UserID       int       `db:"user_id" json:"user_id"`
	Username     string    `db:"username" json:"username"`
	ChatID       int64     `db:"chat_id" json:"chat_id"`
	Warnings     int       `db:"warnings" json:"warnings,omitempty"`
	LanguageCode string    `db:"language_code" json:"language_code"`
	TZ           string    `db:"tz" json:"tz"`
}

// Standup model used for serialization/deserialization stored standups
type Standup struct {
	ID        int64     `db:"id" json:"id"`
	MessageID int       `db:"message_id" json:"message_id"`
	Created   time.Time `db:"created" json:"created"`
	Modified  time.Time `db:"modified" json:"modified"`
	Username  string    `db:"username" json:"userName"`
	Text      string    `db:"text" json:"text"`
	ChatID    int64     `db:"chat_id" json:"chat_id"`
}

// NotificationThread ...
type NotificationThread struct {
	ID               int64     `db:"id" json:"id"`
	ChatID           int64     `db:"chat_id" json:"chat_id"`
	UserID           int       `db:"user_id" json:"user_id"`
	NotificationTime time.Time `db:"notification_time" json:"notification_time"`
	ReminderCounter  int       `db:"reminder_counter" json:"reminder_counter"`
}

// Validate ...
func Validate(nt NotificationThread) error {
	if nt.ChatID == 0 {
		return errors.New("Field ChatID is empty")
	}
	if nt.UserID == 0 {
		return errors.New("Field UserID is empty")
	}
	if strings.TrimSpace(nt.NotificationTime.String()) == "" {
		return errors.New("Field NotificationTime is empty")
	}
	if nt.ReminderCounter < 0 {
		return errors.New("Field ReminderCounter is empty")
	}
	return nil
}
