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
