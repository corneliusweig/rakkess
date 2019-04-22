package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "valid format",
			format: "icon-table",
		},
		{
			name:     "invalid format",
			format:   "cassowary",
			expected: "unexpected output format: cassowary",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := outputFormat(test.format)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}

func TestVerbs(t *testing.T) {
	tests := []struct {
		name     string
		verbs    []string
		expected string
	}{
		{
			name:  "only valid verbs",
			verbs: []string{"list", "get", "proxy"},
		},
		{
			name:     "only invalid verbs",
			verbs:    []string{"lust", "git", "poxy"},
			expected: "unexpected verbs: [git lust poxy]",
		},
		{
			name:     "valid and invalid verbs",
			verbs:    []string{"list", "git", "proxy"},
			expected: "unexpected verbs: [git]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := verbs(test.verbs)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
