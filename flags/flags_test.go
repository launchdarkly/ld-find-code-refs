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

func Test_addFlagKeys(t *testing.T) {
	// Note: these specs assume minFlagKeyLen is 3
	tests := []struct {
		name     string
		flagKeys map[string][]string
		flags    []string
		projKey  string
		want     map[string][]string
	}{
		{
			name:     "With a project that contains no flags",
			flags:    []string{},
			projKey:  "test_project_1",
			flagKeys: map[string][]string{},
			want:     map[string][]string{},
		},
		{
			name:     "With a project that contains no flags and already existing keys",
			flags:    []string{},
			projKey:  "test_project_1",
			flagKeys: map[string][]string{"test_project_2": {"test_key"}},
			want:     map[string][]string{"test_project_2": {"test_key"}},
		},
		{
			name:     "With a project that contains a flag and already existing keys",
			flags:    []string{"test_project_key"},
			projKey:  "test_project_1",
			flagKeys: map[string][]string{"test_project_2": {"test_key"}},
			want:     map[string][]string{"test_project_2": {"test_key"}, "test_project_1": {"test_project_key"}},
		},
		{
			name:     "With a project that contains multiple flags and already existing keys",
			flags:    []string{"test_project_key", "test_project_key_2"},
			projKey:  "test_project_1",
			flagKeys: map[string][]string{"test_project_2": {"test_key"}},
			want:     map[string][]string{"test_project_2": {"test_key"}, "test_project_1": {"test_project_key", "test_project_key_2"}},
		},
		{
			name:     "With a project that contains a short and log flags",
			flags:    []string{"test_project_key", "t"},
			projKey:  "test_project_1",
			flagKeys: map[string][]string{"test_project_2": {"test_key"}},
			want:     map[string][]string{"test_project_2": {"test_key"}, "test_project_1": {"test_project_key"}},
		},
		{
			name:     "With a project that contains a short flag",
			flags:    []string{"k"},
			projKey:  "test_project_1",
			flagKeys: map[string][]string{"test_project_2": {"test_key"}},
			want:     map[string][]string{"test_project_2": {"test_key"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addFlagKeys(tt.flagKeys, tt.flags, tt.projKey)
			require.Equal(t, tt.want, tt.flagKeys)
		})
	}
}
