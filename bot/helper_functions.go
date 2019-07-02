package bot

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/maddevsio/mad-internship-bot/model"
)

func isStandup(message string) bool {
	message = strings.ToLower(message)

	var mentionsYesterdayWork, mentionsTodayPlans, mentionsProblem bool

	for _, work := range yesterdayWorkKeywords {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	for _, plan := range todayPlansKeywords {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}

	for _, problem := range issuesKeywords {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}

	return mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans
}

func analyzeStandup(standup string) (qualityPoints int) {

	/* a good standup contains:

	- structure (yesterday, today, blockers)
	- points (- * or any other indicator)
	- questions (? questions marks or words like why, how, etc)
	- tags of people
	- links to sources
	- amount of text in each block should be fine to avoid yesterday today blockers none
	*/

	// count how many questions inter has
	qualityPoints += strings.Count(standup, "?")
	// count how many mentors and other interns was tagged
	qualityPoints += strings.Count(standup, "@") - 1 // -1 since it always had bot tag in it

	return qualityPoints
}

func (b *Bot) submittedStandupToday(standuper *model.Standuper) bool {
	standup, err := b.db.LastStandupFor(standuper.Username, standuper.ChatID)
	if err != nil {
		return false
	}
	loc, err := time.LoadLocation(standuper.TZ)
	if err != nil {
		logrus.Error("failed to load location for ", standuper)
		return true
	}
	if standup.Created.In(loc).Day() == time.Now().Day() {
		return true
	}
	return false
}
