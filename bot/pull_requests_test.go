package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToAPIEndpoint(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"https://github.com/foo/bar/pull/123", "https://api.github.com/repos/foo/bar/pulls/123"},
	}

	for _, tc := range testCases {
		actual := convertToAPIEndpoint(tc.input)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestContainsPullRequests(t *testing.T) {
	testCases := []struct {
		input           string
		contains        bool
		pullRequestsLen int
	}{
		{"Hey bot, check https://github.com/triggermesh/knative-lambda-sources/pull/26 this PR", true, 1},
		{"Hey bot, check https://gitlab.com/triggermesh/knative-lambda-sources/pull/26 this PR", false, 0},
	}

	for _, tc := range testCases {
		contains, prs := containsPullRequests(tc.input)
		assert.Equal(t, tc.contains, contains)
		assert.Equal(t, tc.pullRequestsLen, len(prs))
	}
}
