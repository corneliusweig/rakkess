package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOutputFormat(t *testing.T) {
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
			actual := ValidateOutputFormat(test.format)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
