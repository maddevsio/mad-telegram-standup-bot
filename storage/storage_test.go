package storage

import (
	"testing"
	"time"

	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRUDStanduper(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	mysql, err := NewMySQL(conf)
	require.NoError(t, err)

	s := &model.Standuper{
		Created:      time.Date(2019, 7, 10, 0, 0, 0, 0, time.Local),
		Status:       "active",
		UserID:       12345,
		Username:     "foo",
		ChatID:       int64(12345),
		Warnings:     0,
		LanguageCode: "en",
	}

	standuper, err := mysql.CreateStanduper(s)
	require.NoError(t, err)
	assert.Equal(t, "active", standuper.Status)
	assert.NotEqual(t, 0, standuper.ID)
	assert.Equal(t, "", standuper.TZ)

	standuper.TZ = "Asia/Bishkek"

	_, err = mysql.UpdateStanduper(standuper)
	require.NoError(t, err)

	standuper, err = mysql.FindStanduper(12345, int64(12345))
	require.NoError(t, err)
	require.NotNil(t, standuper)

	standupers, err := mysql.ListChatStandupers(int64(12345))
	require.NoError(t, err)
	require.Equal(t, 1, len(standupers))

	err = mysql.DeleteStanduper(standuper.ID)
	require.NoError(t, err)

	standupers, err = mysql.ListChatStandupers(int64(12345))
	require.NoError(t, err)
	require.Equal(t, 0, len(standupers))

	_, err = mysql.CreateStanduper(s)
	require.NoError(t, err)

	_, err = mysql.CreateStanduper(s)
	require.NoError(t, err)

	err = mysql.DeleteGroupStandupers(int64(12345))
	require.NoError(t, err)

	standupers, err = mysql.ListChatStandupers(int64(12345))
	require.NoError(t, err)
	require.Equal(t, 0, len(standupers))

}
