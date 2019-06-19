package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/model"
	log "github.com/sirupsen/logrus"
)

//HandleCommand handles imcomming commands
func (b *Bot) HandleCommand(event tgbotapi.Update) error {
	switch event.Message.Command() {
	case "help":
		return b.Help(event)
	case "join":
		return b.JoinStandupers(event)
	case "show":
		return b.Show(event)
	case "leave":
		return b.LeaveStandupers(event)
	case "edit_deadline":
		return b.EditDeadline(event)
	case "show_deadline":
		return b.ShowDeadline(event)
	case "remove_deadline":
		return b.RemoveDeadline(event)
	case "group_tz":
		return b.ChangeGroupTimeZone(event)
	case "tz":
		return b.ChangeUserTimeZone(event)
	default:
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "I do not know this command...")
		_, err := b.tgAPI.Send(msg)
		return err
	}
}

//Help displays help message
func (b *Bot) Help(event tgbotapi.Update) error {
	text := `Список доступных команд:
	/help - Показывает список команд
	/join - Добавляет вас в стендаперы группы
	/show - Показывает кто сдает стендапы в группе
	/leave - Бот перестаёт отслеживать ваши стендапы 
	/edit_deadline - Назначить крайний срок сдачи стендапов (форматы: 13:50, 1:50pm)
	/show_deadline - Показывает срок сдачи стендапов
	/remove_deadline - Убирает срок сдачи дедлайнов 
	/group_tz - Изменить часовой пояс группы (по умолчнию: Asia/Bishkek)
	/tz - Изменить часовой пояс стендапера (по умолчанию: Asia/Bishkek)

	С нетерпением жду ваших стендапов! За все мои ошибки отвечает @anatoliyfedorenko
	`
	msg := tgbotapi.NewMessage(event.Message.Chat.ID, text)
	_, err := b.tgAPI.Send(msg)
	return err
}

//JoinStandupers assign user a standuper role
func (b *Bot) JoinStandupers(event tgbotapi.Update) error {
	_, err := b.db.FindStanduper(event.Message.From.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
	if err == nil {
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Вы уже стендапите")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err := b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.CreateStanduper(&model.Standuper{
		UserID:       event.Message.From.ID,
		Username:     event.Message.From.UserName,
		ChatID:       event.Message.Chat.ID,
		LanguageCode: event.Message.From.LanguageCode,
		TZ:           "Asia/Bishkek", // default value...
	})
	if err != nil {
		log.Error("CreateStanduper failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог добавить в стендаперы")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err := b.tgAPI.Send(msg)
		return err
	}

	group, err := b.db.FindGroup(event.Message.Chat.ID)
	if err != nil {
		group, err = b.db.CreateGroup(&model.Group{
			ChatID:          event.Message.Chat.ID,
			Title:           event.Message.Chat.Title,
			Description:     event.Message.Chat.Description,
			StandupDeadline: "10:00",
			TZ:              "Asia/Bishkek", // default value...
		})
		if err != nil {
			return err
		}
	}

	var msg tgbotapi.MessageConfig

	if group.StandupDeadline == "" {
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, "Добро пожаловать в команду! Срок сдачи стендапов пока не установлен")
	} else {
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, fmt.Sprintf("Добро пожаловать в команду! Пишите ваши стендапы ежедневно до %s. В выходные пишите стендапы по желанию", group.StandupDeadline))
	}

	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//Show standupers
func (b *Bot) Show(event tgbotapi.Update) error {
	standupers, err := b.db.ListChatStandupers(event.Message.Chat.ID)
	if err != nil {
		return err
	}

	if len(standupers) == 0 {
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Никто не пишет стендапы. Можно присоединиться с помощью команды /join")
		_, err := b.tgAPI.Send(msg)
		return err
	}

	list := []string{}
	for _, standuper := range standupers {
		list = append(list, "@"+standuper.Username)
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, fmt.Sprintf("Стендапят: %v", strings.Join(list, ", ")))
	_, err = b.tgAPI.Send(msg)
	return err
}

//LeaveStandupers standupers
func (b *Bot) LeaveStandupers(event tgbotapi.Update) error {
	standuper, err := b.db.FindStanduper(event.Message.From.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
	if err != nil {
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Вы не стендапите")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err := b.tgAPI.Send(msg)
		return err
	}

	err = b.db.DeleteStanduper(standuper.ID)
	if err != nil {
		log.Error("DeleteStanduper failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог убрать из стендап команды")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err := b.tgAPI.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Вы больше не стендапите, спасибо за все ваши сообщения!")
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//EditDeadline modifies standup time
func (b *Bot) EditDeadline(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	deadline := event.Message.CommandArguments()

	if strings.TrimSpace(deadline) == "" {
		return nil
	}

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		log.Error("findTeam failed")
		return fmt.Errorf("failed to find sutable team for edit deadline")
	}

	team.Group.StandupDeadline = deadline

	log.Info(team.Group)

	_, err = b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in EditDeadline failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог установить срок сдачи стендапов")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, fmt.Sprintf("Срок сдачи стендапов обновлён! Писать стендапы до %s", deadline))
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//ShowDeadline shows current standup time
func (b *Bot) ShowDeadline(event tgbotapi.Update) error {
	group, err := b.db.FindGroup(event.Message.Chat.ID)
	if err != nil {
		group, err = b.db.CreateGroup(&model.Group{
			ChatID:          event.Message.Chat.ID,
			Title:           event.Message.Chat.Title,
			Description:     event.Message.Chat.Description,
			StandupDeadline: "10:00",
			TZ:              "Asia/Bishkek", // default value...
		})
		if err != nil {
			return err
		}
	}

	var msg tgbotapi.MessageConfig

	if group.StandupDeadline == "" {
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, "Срок сдачи стендапов не установлен.")
	} else {
		msg = tgbotapi.NewMessage(event.Message.Chat.ID, fmt.Sprintf("Срок сдачи стендапов ежедневно до %s кроме выходных", group.StandupDeadline))
	}

	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//RemoveDeadline sets standup deadline to empty string
func (b *Bot) RemoveDeadline(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		group, err := b.db.CreateGroup(&model.Group{
			ChatID:          event.Message.Chat.ID,
			Title:           event.Message.Chat.Title,
			Description:     event.Message.Chat.Description,
			StandupDeadline: "10:00",
			TZ:              "Asia/Bishkek", // default value...
		})
		if err != nil {
			return err
		}
		b.watchersChan <- group
		team = b.findTeam(event.Message.Chat.ID)
	}

	team.Group.StandupDeadline = ""

	_, err = b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in RemoveDeadline failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не удалось убрать срок сдачи стендапов")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Срок сдачи стендапов успешно удалён")
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//ChangeGroupTimeZone modifies time zone of the group
func (b *Bot) ChangeGroupTimeZone(event tgbotapi.Update) error {
	isAdmin, err := b.senderIsAdminInChannel(event.Message.From.UserName, event.Message.Chat.ID)
	if err != nil {
		log.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	if !isAdmin {
		log.Warn("User not an admin", event.Message.From.UserName)
		return nil
	}

	tz := event.Message.CommandArguments()

	if strings.TrimSpace(tz) == "" {
		return nil
	}

	team := b.findTeam(event.Message.Chat.ID)
	if team == nil {
		group, err := b.db.CreateGroup(&model.Group{
			ChatID:          event.Message.Chat.ID,
			Title:           event.Message.Chat.Title,
			Description:     event.Message.Chat.Description,
			StandupDeadline: "10:00",
			TZ:              "Asia/Bishkek", // default value...
		})
		if err != nil {
			return err
		}
		b.watchersChan <- group
		team = b.findTeam(event.Message.Chat.ID)
	}

	team.Group.TZ = tz

	_, err = time.LoadLocation(tz)
	if err != nil {
		log.Error("UpdateGroup in ChangeTimeZone failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог изменить часовой пояс, пожалуйста проверьте название часовой зоны и повторите")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.UpdateGroup(team.Group)
	if err != nil {
		log.Error("UpdateGroup in ChangeTimeZone failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог изменить часовой пояс")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, fmt.Sprintf("Часовой пояс группы обновлён на: %s", tz))
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//ChangeUserTimeZone assign user a different time zone
func (b *Bot) ChangeUserTimeZone(event tgbotapi.Update) error {
	tz := event.Message.CommandArguments()

	if strings.TrimSpace(tz) == "" {
		return nil
	}

	st, err := b.db.FindStanduper(event.Message.From.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
	if err != nil {
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Вы не стендапите, для начала присоединитесь с помощью /join а потом сможете поменять свой часовой пояс")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err := b.tgAPI.Send(msg)
		return err
	}

	st.TZ = tz

	_, err = time.LoadLocation(tz)
	if err != nil {
		log.Error("LoadLocation in ChangeUserTimeZone failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог изменить часовой пояс, пожалуйста проверьте название часовой зоны и повторите")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	_, err = b.db.UpdateStanduper(st)
	if err != nil {
		log.Error("UpdateStanduper in ChangeUserTimeZone failed: ", err)
		msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Не смог изменить часовой пояс")
		msg.ReplyToMessageID = event.Message.MessageID
		_, err = b.tgAPI.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, fmt.Sprintf("Ваш часовой пояс изменен. Новый: %s", tz))
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

func (b *Bot) senderIsAdminInChannel(sendername string, chatID int64) (bool, error) {
	isAdmin := false
	chat := tgbotapi.ChatConfig{chatID, ""}
	admins, err := b.tgAPI.GetChatAdministrators(chat)
	if err != nil {
		return false, err
	}
	for _, admin := range admins {
		if admin.User.UserName == sendername {
			isAdmin = true
			return true, nil
		}
	}
	return isAdmin, nil
}
