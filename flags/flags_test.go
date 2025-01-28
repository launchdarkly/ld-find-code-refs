package flags

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bucketeer-io/code-refs/internal/log"
)

func init() {
	log.Init(true)
}
func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func Test_filterShortFlags(t *testing.T) {
	// Note: these specs assume minFlagKeyLen is 3
	tests := []struct {
		name  string
		flags []string
		want  []string
	}{
		{
			name:  "Empty input/output",
			flags: []string{},
			want:  []string{},
		},
		{
			name:  "all flags are too short",
			flags: []string{"a", "b"},
			want:  []string{},
		},
		{
			name:  "some flags are too short",
			flags: []string{"abcdefg", "b", "ab", "abc"},
			want:  []string{"abcdefg", "abc"},
		},
		{
			name:  "no flags are too short",
			flags: []string{"catsarecool", "dogsarecool"},
			want:  []string{"catsarecool", "dogsarecool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filterShortFlagKeys(tt.flags)
			require.Equal(t, tt.want, got)
		})
	}
}
