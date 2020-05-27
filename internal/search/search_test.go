package search

import (
	"os"
	"strings"
	"testing"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Init(true)
}
func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

const (
	testFlagKey     = "someFlag"
	testFlagKey2    = "anotherFlag"
	testFlagAlias   = "some-flag"
	testFlagAlias2  = "some.flag"
	testFlag2Alias  = "another-flag"
	testFlag2Alias2 = "another.flag"
)

var (
	firstFlag            = map[string][]string{testFlagKey: {}}
	firstFlagWithAlias   = map[string][]string{testFlagKey: {testFlagAlias}}
	firstFlagWithAliases = map[string][]string{testFlagKey: {testFlagAlias, testFlagAlias2}}
	secondFlag           = map[string][]string{testFlagKey2: {}}
	twoFlags             = map[string][]string{testFlagKey: {}, testFlagKey2: {}}
	twoFlagsWithAliases  = map[string][]string{testFlagKey: {testFlagAlias, testFlagAlias2}, testFlagKey2: {testFlag2Alias, testFlag2Alias2}}
	noFlags              = map[string][]string{}
)

func Test_truncateLine(t *testing.T) {
	longLine := strings.Repeat("a", maxLineCharCount)

	veryLongLine := strings.Repeat("a", maxLineCharCount+1)

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "empty line",
			line: "",
			want: "",
		},
		{
			name: "line shorter than max length",
			line: "abc efg",
			want: "abc efg",
		},
		{
			name: "long line",
			line: longLine,
			want: longLine,
		},
		{
			name: "very long line",
			line: veryLongLine,
			want: veryLongLine[0:maxLineCharCount] + "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateLine(tt.line)
			require.Equal(t, tt.want, got)
		})
	}
}
