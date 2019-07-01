package bot

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
)

const allowedSkips = 3
const warnPeriod = 10 // 10 minutes before the deadline

//StartWatchers looks for new gropus from the channel and start watching it
func (b *Bot) StartWatchers() {
	for group := range b.watchersChan {
		log.Info("New group to track: ", group)
		team := &model.Team{
			Group:    group,
			QuitChan: make(chan struct{}),
		}
		b.teams = append(b.teams, team)
		b.wg.Add(1)
		go b.trackStandupersIn(team)
		b.wg.Done()
	}
}

func (b *Bot) trackStandupersIn(team *model.Team) {
	ticker := time.NewTicker(time.Second * 60).C
	for {
		select {
		case <-ticker:
			b.WarnGroup(team.Group, time.Now())
			b.NotifyGroup(team.Group, time.Now())
		case <-team.QuitChan:
			log.Info("Finish working with the group: ", team.QuitChan)
			return
		}
	}
}

//WarnGroup launches go routines that warns standupers
//about upcoming deadlines
func (b *Bot) WarnGroup(group *model.Group, t time.Time) {
	if int(t.Weekday()) == 6 || int(t.Weekday()) == 0 {
		return
	}

	if group.StandupDeadline == "" {
		return
	}
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, err := w.Parse(group.StandupDeadline, time.Now())
	if err != nil {
		log.Errorf("Unable to parse channel standup time [%v]: [%v]", group.StandupDeadline, err)
		return
	}

	if r == nil {
		log.Errorf("Could not find matches. Channel standup time: [%v]", group.StandupDeadline)
		return
	}

	t = t.Add(warnPeriod * time.Minute)

	if t.Hour() != r.Time.Hour() || t.Minute() != r.Time.Minute() {
		return
	}

	standupers, err := b.db.ListChatStandupers(group.ChatID)
	if err != nil {
		log.Error(err)
		return
	}

	if len(standupers) == 0 {
		return
	}

	stillDidNotSubmit := map[string]int{}

	for _, standuper := range standupers {
		if b.submittedStandupToday(standuper) {
			continue
		}
		stillDidNotSubmit["@"+standuper.Username] = standuper.Warnings
	}

	//? if everything is fine, should not bother the team...
	if len(stillDidNotSubmit) == 0 {
		return
	}

	var text string

	for key, value := range stillDidNotSubmit {
		text += fmt.Sprintf("Внимание, %v, до дедлайна осталось %v минут! Срочно пишите стендап и не подводите команду! Осталось пропусков: %v \n\n", key, warnPeriod, allowedSkips-value)
	}

	msg := tgbotapi.NewMessage(group.ChatID, text)
	_, err = b.tgAPI.Send(msg)
	if err != nil {
		log.Error(err)
	}
}

//NotifyGroup launches go routines that notify standupers
//about upcoming deadlines
func (b *Bot) NotifyGroup(group *model.Group, t time.Time) {
	if int(t.Weekday()) == 6 || int(t.Weekday()) == 0 {
		return
	}

	if group.StandupDeadline == "" {
		return
	}
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, err := w.Parse(group.StandupDeadline, time.Now())
	if err != nil {
		log.Errorf("Unable to parse channel standup time [%v]: [%v]", group.StandupDeadline, err)
		return
	}

	if r == nil {
		log.Errorf("Could not find matches. Channel standup time: [%v]", group.StandupDeadline)
		return
	}

	if t.Hour() != r.Time.Hour() || t.Minute() != r.Time.Minute() {
		return
	}

	standupers, err := b.db.ListChatStandupers(group.ChatID)
	if err != nil {
		log.Error(err)
		return
	}

	if len(standupers) == 0 {
		return
	}

	missed := map[string]int{}

	for _, standuper := range standupers {
		if b.submittedStandupToday(standuper) {
			continue
		}
		if standuper.Warnings >= allowedSkips {
			log.Infof("Missed %v standups! Should kick member!", allowedSkips)
			resp, err := b.tgAPI.KickChatMember(tgbotapi.KickChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID:             standuper.ChatID,
					SuperGroupUsername: group.Username,
					ChannelUsername:    standuper.Username,
					UserID:             standuper.UserID,
				},
				UntilDate: time.Now().Unix(),
			})
			if err != nil {
				log.Error("Failed to kick user: ", err)
			}
			log.Info(resp)
			continue
		}
		standuper.Warnings++
		missed["@"+standuper.Username] = standuper.Warnings
		b.db.UpdateStanduper(standuper)
	}

	//? if everything is fine, should not bother the team...
	if len(missed) == 0 {
		return
	}

	var text string

	for key, value := range missed {
		text += fmt.Sprintf("Внимание, %v, вы пропустили крайний срок сдачи стендапов! Срочно пишите стендап и не подводите команду! Осталось пропусков: %v \n\n", key, allowedSkips-value)
	}

	msg := tgbotapi.NewMessage(group.ChatID, text)
	_, err = b.tgAPI.Send(msg)
	if err != nil {
		log.Error(err)
	}
}
