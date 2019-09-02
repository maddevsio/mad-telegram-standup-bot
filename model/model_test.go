package model

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	n   NotificationThread
	exp error
}{
	{NotificationThread{ChatID: int64(0), Username: "User1", NotificationTime: time.Now(), ReminderCounter: 0}, errors.New("Field ChatID is empty")},
	{NotificationThread{ChatID: int64(1), Username: "", NotificationTime: time.Now(), ReminderCounter: 0}, errors.New("Field Username is empty")},
	{NotificationThread{ChatID: int64(1), Username: "User1", NotificationTime: time.Now(), ReminderCounter: -1}, errors.New("Field ReminderCounter is empty")},
	{NotificationThread{ChatID: int64(1), Username: "User1", NotificationTime: time.Now(), ReminderCounter: 1}, nil},
}

func TestValidateNotificationThread(t *testing.T) {
	for _, e := range tests {
		res := Validate(e.n)
		assert.Equal(t, res, e.exp)
	}
}
