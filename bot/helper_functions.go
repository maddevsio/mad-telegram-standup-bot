package bot

import (
	"strings"
	"time"

	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/sirupsen/logrus"
)

func isStandup(message string) (bool, []string) {
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
		errors = append(errors, "- не увидел ключевых слов блока 'вчера': "+strings.Join(yesterdayWorkKeywords, ", "))
	}
	if !mentionsTodayPlans {
		errors = append(errors, "- не увидел ключевых слов блока 'сегодня': "+strings.Join(todayPlansKeywords, ", "))
	}
	if !mentionsYesterdayWork {
		errors = append(errors, "- не увидел ключевых слов блока 'проблемы': "+strings.Join(issuesKeywords, ", "))
	}

	return mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans, errors
}

func analyzeStandup(standup string) ([]string, int) {
	var advises []string
	ok, pB := containsProblems(standup)
	if !ok {
		advises = append(advises, "- Кажется в стендапе нет или мало проблем. Проблемы и всё, что мешает это показатель роста. Если их нет, это плохо. не бойся об этом говорить")
	}

	ok, qB := containsQuestions(standup)
	if !ok {
		advises = append(advises, "- Кажется в стендапе не задано вопросов. На стажировке надо задавать вопросы, самое лучше место это общий чат и стендапы. Без вопросов нет развития")
	}

	ok, mB := containsMentions(standup)
	if !ok {
		advises = append(advises, "- Кажется в стендапе нет тегов. Тегай менторов, чтобы получить их опыт, иначе прогресс будет медленный")
	}

	ok, lB := containsLinks(standup)
	if !ok {
		advises = append(advises, "- Кажется в стендапе нет ни одной ссылки. Желательно отправлять ссылки на PRы либо на изученные ресурсы")
	}

	ok, sB := hasGoodSize(standup)
	if !ok {
		advises = append(advises, "- Подумай над размером стендапа. Напишешь мало - непонятно, грач, без уважения. Много - тяжело читать. Оптимально от 80 до 200 слов")
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
