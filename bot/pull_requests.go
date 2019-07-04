package bot

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

func containsPullRequests(message string) (bool, []github.PullRequest) {
	var prs []github.PullRequest
	var linkList []string
	words := strings.Fields(message)

	for _, word := range words {
		if strings.Contains(word, "pull") && strings.Contains(word, "github.com") {
			linkList = append(linkList, word)
		}
	}

	if len(linkList) == 0 {
		return false, prs
	}

	for _, link := range linkList {
		var pr github.PullRequest
		endpoint := convertToAPIEndpoint(link)
		r, err := http.Get(endpoint)
		if err != nil {
			log.Error("Could not get info about PR: ", err)
			continue
		}

		defer r.Body.Close()

		err = json.NewDecoder(r.Body).Decode(&pr)
		if err != nil {
			log.Error("Could not decode info about PR: ", err)
			continue
		}
		prs = append(prs, pr)
	}

	if len(prs) == 0 {
		return false, prs
	}

	return true, prs
}

func convertToAPIEndpoint(link string) string {
	link = strings.Replace(link, "https://github.com", "https://api.github.com/repos", -1)
	link = strings.Replace(link, "/pull/", "/pulls/", -1)
	return link
}

func analyzePullRequest(pr github.PullRequest) []string {
	errors := []string{}

	if len(*pr.Body) < 50 {
		errors = append(errors, "- нужно больше описания PRа")
	}

	if pr.Assignee == nil || len(pr.Assignees) == 0 {
		errors = append(errors, "- не назначен проверяющий, который должен посмотреть PR")
	}

	if !*pr.Mergeable {
		errors = append(errors, "- нельза смержить, нужно перепроверить код на конфликты")
	}

	if !strings.ContainsAny(*pr.Title, "#") && !strings.ContainsAny(*pr.Body, "#") {
		errors = append(errors, "- нет ссылок на тикеты которые закроет этот PR. необходимо всегда писать код по тикетам!")
	}

	if *pr.Additions > 300 {
		errors = append(errors, "- слишком много проверять, надо раздробить PR на части")
	}

	if strings.Contains(*pr.Title, "[WIP]") {
		errors = append(errors, "- PR содержит незаконченную работу. переотправьте как будет всё готово")
	}

	return errors
}
