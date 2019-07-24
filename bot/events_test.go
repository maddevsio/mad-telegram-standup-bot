package bot

import (
	"testing"

	"github.com/BurntSushi/toml"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/text/language"
)

func TestHandleUpdate(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err = bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	bot := Bot{c: conf, bundle: bundle}
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "Foo",
			/* Chat: &tgbotapi.Chat{
				Type: "private",
			}, */
		},
	}

	text, err := bot.handleUpdate(update)
	group, err := b.db.FindGroup(message.Chat.ID)
	require.NoError(t, err)
	assert.Equal(t, "- bad PR, pay attention to the following advises: \n", text)
}
