package parse

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateReferencesFromGrep(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		grepResult [][]string
		ctxLines   int
		want       []reference
		exclude    string
	}{
		{
			name:  "succeeds",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
			},
			ctxLines: 0,
			want: []reference{
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
			},
		},
		{
			name:  "succeeds with exclude",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
			},
			ctxLines: 0,
			want:     []reference{},
			exclude:  ".*",
		},
		{
			name:  "succeeds with no LineText lines",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
			},
			ctxLines: -1,
			want: []reference{
				{Path: "flags.txt", LineNum: 12, FlagKeys: []string{"someFlag"}},
			},
		},
		{
			name:  "succeeds with multiple references",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
				{"", "path/flags.txt", ":", "12", "someFlag anotherFlag"},
			},
			ctxLines: 0,
			want: []reference{
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
				{Path: "path/flags.txt", LineNum: 12, LineText: "someFlag anotherFlag", FlagKeys: []string{"someFlag", "anotherFlag"}},
			},
		},
		{
			name:  "succeeds with extra LineText lines",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				{"", "flags.txt", ":", "12", "someFlag"},
				{"", "flags.txt", "-", "13", "not a flag key line"},
			},
			ctxLines: 1,
			want: []reference{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex, err := regexp.Compile(tt.exclude)
			require.NoError(t, err)
			got := generateReferencesFromGrep(tt.flags, tt.grepResult, tt.ctxLines, ex)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_findReferencedFlags(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want []string
	}{
		{
			name: "finds a flag",
			ref:  "line contains someFlag",
			want: []string{"someFlag"},
		},
		{
			name: "finds multiple flags",
			ref:  "line contains someFlag and anotherFlag",
			want: []string{"someFlag", "anotherFlag"},
		},
		{
			name: "finds no flags",
			ref:  "line contains no flags",
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findReferencedFlags(tt.ref, []string{"someFlag", "anotherFlag"})
			require.Equal(t, tt.want, got)
		})
	}
}
