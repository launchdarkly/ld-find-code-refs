package search

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateReferences(t *testing.T) {
	testResult := []string{"", "flags.txt", ":", "12", delimit(testFlagKey, `"`)}
	testWant := SearchResultLine{Path: "flags.txt", LineNum: 12, LineText: delimit(testFlagKey, `"`), FlagKeys: firstFlag}

	tests := []struct {
		name         string
		flags        map[string][]string
		searchResult [][]string
		ctxLines     int
		want         []SearchResultLine
	}{
		{
			name:         "succeeds",
			flags:        twoFlags,
			searchResult: [][]string{testResult},
			ctxLines:     0,
			want:         []SearchResultLine{testWant},
		},
		{
			name:         "succeeds with no LineText lines",
			flags:        twoFlags,
			searchResult: [][]string{testResult},
			ctxLines:     -1,
			want: []SearchResultLine{
				{Path: "flags.txt", LineNum: 12, FlagKeys: firstFlag},
			},
		},
		{
			name:  "succeeds with multiple references",
			flags: twoFlags,
			searchResult: [][]string{
				testResult,
				{"", "path/flags.txt", ":", "12", `"someFlag" "anotherFlag"`},
			},
			ctxLines: 0,
			want: []SearchResultLine{
				testWant,
				{Path: "path/flags.txt", LineNum: 12, LineText: `"someFlag" "anotherFlag"`, FlagKeys: twoFlags},
			},
		},
		{
			name:  "succeeds with aliases",
			flags: firstFlagWithAliases,
			searchResult: [][]string{
				{"", "path/flags.txt", ":", "12", testFlagAlias},
			},
			ctxLines: 0,
			want: []SearchResultLine{
				{Path: "path/flags.txt", LineNum: 12, LineText: testFlagAlias, FlagKeys: firstFlagWithAlias},
			},
		},
		{
			name:  "succeeds with alias and flag key",
			flags: firstFlagWithAliases,
			searchResult: [][]string{
				{"", "path/flags.txt", ":", "12", delimit(testFlagKey, "'") + " " + testFlagAlias},
			},
			ctxLines: 0,
			want: []SearchResultLine{
				{Path: "path/flags.txt", LineNum: 12, LineText: delimit(testFlagKey, "'") + " " + testFlagAlias, FlagKeys: firstFlagWithAlias},
			},
		},
		{
			name:  "succeeds with extra LineText lines",
			flags: twoFlags,
			searchResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				testResult,
				{"", "flags.txt", "-", "13", "not a flag key line"},
			},
			ctxLines: 1,
			want: []SearchResultLine{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				testWant,
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
			},
		},
		{
			name:  "succeeds with extra LineText lines and multiple flags",
			flags: twoFlags,
			searchResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				testResult,
				{"", "flags.txt", "-", "13", "not a flag key line"},
				{"", "flags.txt", ":", "14", delimit(testFlagKey2, `"`)},
				{"", "flags.txt", "-", "15", "not a flag key line"},
			},
			ctxLines: 1,
			want: []SearchResultLine{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				testWant,
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
				{Path: "flags.txt", LineNum: 14, LineText: delimit(testFlagKey2, `"`), FlagKeys: secondFlag},
				{Path: "flags.txt", LineNum: 15, LineText: "not a flag key line"},
			},
		},
		{
			name:         "does not match substring flag key",
			flags:        map[string][]string{testFlagKey: {}, testFlagKey2[:4]: {}},
			searchResult: [][]string{testResult},
			ctxLines:     0,
			want:         []SearchResultLine{testWant},
		},
		{
			// delimeters don't have to match on both sides
			name:  "succeeds with multiple delimiters",
			flags: map[string][]string{testFlagKey: {}, "some": {}},
			searchResult: [][]string{
				{"", "flags.txt", ":", "12", `"` + testFlagKey + "'"},
			},
			ctxLines: 0,
			want: []SearchResultLine{
				{Path: "flags.txt", LineNum: 12, LineText: `"` + testFlagKey + "'", FlagKeys: firstFlag},
			},
		},
	}

	client := AgClient{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.generateReferences(tt.flags, tt.searchResult, tt.ctxLines, `"'`)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_findReferencedFlags(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want map[string][]string
	}{
		{
			name: "finds a flag",
			ref:  "line contains " + delimit(testFlagKey, `"`),
			want: firstFlag,
		},
		{
			name: "finds multiple flags",
			ref:  "line contains " + delimit(testFlagKey, `"`) + " " + delimit(testFlagKey2, `"`),
			want: twoFlags,
		},
		{
			name: "finds no flags",
			ref:  "line contains no flags",
			want: noFlags,
		},
		{
			name: "finds one alias",
			ref:  fmt.Sprintf("line contains %s", testFlagAlias),
			want: firstFlagWithAlias,
		},
		{
			name: "finds all aliases",
			ref:  fmt.Sprintf("line contains %s %s %s %s %s %s", delimit(testFlagKey, `"`), delimit(testFlagKey2, `"`), testFlagAlias, testFlagAlias2, testFlag2Alias, testFlag2Alias2),
			want: twoFlagsWithAliases,
		},
	}

	client := AgClient{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.findReferencedFlags(tt.ref, twoFlagsWithAliases, `"`)
			require.Equal(t, tt.want, got)
		})
	}
}
