package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
)

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
			b.NotifyGroup(team.Group, time.Now())
		case <-team.QuitChan:
			log.Info("Finish working with the group: ", team.QuitChan)
			return
		}
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

	missed := []string{}

	for _, standuper := range standupers {
		if !b.submittedStandupToday(standuper) {
			if standuper.Warnings >= 1 {
				log.Info("Missed more than 1 standup! Should kick member!")
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
			missed = append(missed, "@"+standuper.Username)
			standuper.Warnings++
			b.db.UpdateStanduper(standuper)
		}
	}

	//? if everything is fine, should not bother the team...
	if len(missed) == 0 {
		return
	}

	msg := tgbotapi.NewMessage(group.ChatID, fmt.Sprintf("Внимание, %v, вы пропустили крайний срок сдачи стендапов! Срочно пишите стендапы, не подводите команду. Еще один пропуск и я удалю вас из группы", strings.Join(missed, ", ")))
	_, err = b.tgAPI.Send(msg)
	if err != nil {
		log.Error(err)
	}
}
