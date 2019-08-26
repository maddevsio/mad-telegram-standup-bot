package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var n = NotificationThread{
	ChatID:           int64(1),
	UserID:           1,
	NotificationTime: time.Now(),
	ReminderCounter:  0,
}

func TestValidate(t *testing.T) {
	err := Validate(n)
	require.Nil(t, err)
}
