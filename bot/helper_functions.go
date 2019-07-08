package bot

import (
	"strings"
	"time"

	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

func (b *Bot) isStandup(message, language string) (bool, []string) {
	localizer := i18n.NewLocalizer(b.bundle, language)

	errors := []string{}
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

	if !mentionsYesterdayWork {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noYesterdayMention",
				Other: "- no 'yesterday' keywords detected: {{.Keywords}}",
			},
			TemplateData: map[string]interface{}{
				"Keywords": strings.Join(yesterdayWorkKeywords, ", "),
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		errors = append(errors, warnings)
	}
	if !mentionsTodayPlans {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noTodayMention",
				Other: "- no 'today' keywords detected: {{.Keywords}}",
			},
			TemplateData: map[string]interface{}{
				"Keywords": strings.Join(todayPlansKeywords, ", "),
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		errors = append(errors, warnings)
	}
	if !mentionsProblem {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noProblemsMention",
				Other: "- no 'problems' keywords detected: {{.Keywords}}",
			},
			TemplateData: map[string]interface{}{
				"Keywords": strings.Join(issuesKeywords, ", "),
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		errors = append(errors, warnings)
	}

	return mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans, errors
}

func (b *Bot) analyzeStandup(standup, language string) ([]string, int) {
	localizer := i18n.NewLocalizer(b.bundle, language)

	var advises []string
	ok, pB := containsProblems(standup)
	if !ok {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzeNoBlockers",
				Other: "- seems like standups does not contain any prolem or blocker. Remember that problems help us grow. No problems == no development. Dont hesitate to report them",
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		advises = append(advises, warnings)
	}

	ok, qB := containsQuestions(standup)
	if !ok {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzeNoQuestions",
				Other: "- seems like standups does not contain any questions. Internship is made up of questions, so ask as many as you can!",
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		advises = append(advises, warnings)
	}

	ok, mB := containsMentions(standup)
	if !ok {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzeNoTags",
				Other: "- seems like standups does not contain any tags. If you want your mentors to notice you, tag them right away!",
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		advises = append(advises, warnings)
	}

	ok, lB := containsLinks(standup)
	if !ok {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzeNoLinks",
				Other: "- seems like standups does not contain any links, that means that probably no work was done or no research was conducted. Poor you.",
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		advises = append(advises, warnings)
	}

	ok, sB := hasGoodSize(standup)
	if !ok {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzeNoSize",
				Other: "- seems like standups does is either small or too large. Appropriate size is about 80 - 200 words.",
			},
		})
		if err != nil {
			logrus.Error(err)
		}
		advises = append(advises, warnings)
	}

	return advises, pB + qB + mB + lB + sB
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
	standup = strings.ToLower(standup)
	var wordsAfterProblemsKeyword int
	var positionOfProblemsKeyword int
	var found bool
	words := strings.Fields(standup)

	for i, word := range words {
		for _, problem := range issuesKeywords {
			if strings.Contains(word, problem) {
				positionOfProblemsKeyword = i
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	wordsAfterProblemsKeyword = len(words) - positionOfProblemsKeyword

	if wordsAfterProblemsKeyword > 5 {
		return true, wordsAfterProblemsKeyword
	}
	return false, wordsAfterProblemsKeyword
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

func containsLists(standup string) (bool, int) {
	lists := strings.Count(standup, "-")
	if lists > 1 {
		return true, lists
	}
	return false, lists
}

func hasGoodSize(standup string) (bool, int) {
	words := strings.Fields(standup)
	if len(words) > 80 && len(words) < 200 {
		return true, 1
	}
	return false, 0
}
