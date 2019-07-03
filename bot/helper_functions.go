package bot

import (
	"strings"
	"time"

	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/sirupsen/logrus"
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

func analyzeStandup(standup string) []string {
	var advises []string
	return advises
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

func containsProblems(standup string) (bool, int) {
	var wordsAfterProblemsKeyword int
	var positionOfProblemsKeyword int
	words := strings.Fields(standup)
	for i, word := range words {
		for _, problem := range issuesKeywords {
			if strings.Contains(word, problem) {
				positionOfProblemsKeyword = i
				break
			}
		}
	}
	wordsAfterProblemsKeyword = len(words) - positionOfProblemsKeyword

	if wordsAfterProblemsKeyword > 5 {
		return true, wordsAfterProblemsKeyword
	}
	return false, wordsAfterProblemsKeyword
}

func containsLists(standup string) (bool, int) {
	lists := strings.Count(standup, "-")
	if lists > 1 {
		return true, lists
	}
	return false, lists
}

func containsQuestions(standup string) (bool, int) {
	questions := strings.Count(standup, "?")
	if questions != 0 {
		return true, questions
	}
	return false, questions
}

func containsMentions(standup string) (bool, int) {
	tags := strings.Count(standup, "@") - 1
	if tags != 0 {
		return true, tags
	}
	return false, tags
}

func containsLinks(standup string) (bool, int) {
	links := strings.Count(standup, "http")
	if links != 0 {
		return true, links
	}
	return false, links
}
