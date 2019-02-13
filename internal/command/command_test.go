package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultDelims = []rune{'"', '\'', '`'}

func TestGenerateFlagRegex(t *testing.T) {
	specs := []struct {
		name     string
		flags    []string
		expected string
	}{
		{
			name:     "succeeds for single flag",
			flags:    []string{"flag"},
			expected: "flag"},
		{
			name:     "succeeds for single flag with escape char",
			flags:    []string{"^flag"},
			expected: `\^flag`},
		{
			name:     "succeeds for multiple flags",
			flags:    []string{"flag1", "flag2"},
			expected: "flag1|flag2"},
		{
			name:     "succeeds for multiple flags with escape characters",
			flags:    []string{"^flag1", ".flag2", "*flag3"},
			expected: `\^flag1|\.flag2|\*flag3`},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateFlagRegex(tt.flags))
		})
	}
}

func TestGenerateDelimiterRegex(t *testing.T) {
	specs := []struct {
		name               string
		delimiters         []rune
		expectedLookBehind string
		expectedLookAhead  string
	}{
		{
			name:               "succeeds for default delimiters",
			delimiters:         defaultDelims,
			expectedLookBehind: "(?<=[\"'`])",
			expectedLookAhead:  "(?=[\"'`])",
		},
		{
			name:               "succeeds for extra delimiter",
			delimiters:         append(defaultDelims, 'a'),
			expectedLookBehind: "(?<=[\"'`a])",
			expectedLookAhead:  "(?=[\"'`a])",
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			lb, la := generateDelimiterRegex(tt.delimiters)
			assert.Equal(t, tt.expectedLookBehind, lb)
			assert.Equal(t, tt.expectedLookAhead, la)
		})
	}
}

func TestGenerateSearchPattern(t *testing.T) {
	specs := []struct {
		name       string
		padPattern bool
		expected   string
	}{
		{
			name:       "correctly pads flag pattern",
			padPattern: true,
			expected:   "(?<=[\"'`])(a^|flag|a^)(?=[\"'`])",
		},
		{
			name:       "correctly doesn't pad pattern",
			padPattern: false,
			expected:   "(?<=[\"'`])(flag)(?=[\"'`])",
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateSearchPattern([]string{"flag"}, defaultDelims, tt.padPattern))
		})
	}
}
