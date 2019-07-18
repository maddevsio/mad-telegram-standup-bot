package bot

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

func (b *Bot) analyzePullRequest(pr github.PullRequest, language string) []string {
	localizer := i18n.NewLocalizer(b.bundle, language)

	errors := []string{}

	if pr.Body == nil || len(*pr.Body) < 50 {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRDescription",
				Other: "- need more words in PR Description field",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	if pr.Assignee == nil || len(pr.Assignees) == 0 {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRAsignee",
				Other: "- need to add who is assigned to implement changes to this PR",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	if pr.RequestedReviewers == nil || len(pr.RequestedReviewers) == 0 {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRReviewer",
				Other: "- no reviewers tagged, please, assign at least one",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	if pr.Mergeable == nil || !*pr.Mergeable {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRConflicts",
				Other: "- no way to merge it. Fix conflicts first",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	if !strings.ContainsAny(*pr.Title, "#") && !strings.ContainsAny(*pr.Body, "#") {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRLinks",
				Other: "- need to include links on tickets which would be closed by this PR",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	if *pr.Additions > 300 {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRSize",
				Other: "- PR contains too much changes. Divide it and resend.",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	if strings.Contains(*pr.Title, "[WIP]") {
		warnings, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "analyzePRWIP",
				Other: "- PR contains unfinished work, please, finish it and resend the link",
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}

	return errors
}
