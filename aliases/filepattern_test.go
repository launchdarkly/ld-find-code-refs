package aliases

import (
	"os"
	"testing"

	o "github.com/launchdarkly/ld-find-code-refs/v2/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_processFileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	f, err := os.MkdirTemp(tmpDir, "testalias")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpDir)

	emptyMap := make(map[string][]byte)
	tests := []struct {
		name    string
		dir     string
		aliases []o.Alias
		want    map[string][]byte
		wantErr bool
	}{
		{
			name: "Existing directory and file",
			aliases: []o.Alias{
				{
					Paths: []string{f},
				},
			},
			dir:     tmpDir,
			want:    emptyMap,
			wantErr: false,
		},
		{
			name: "Non-existent directory",
			aliases: []o.Alias{
				{
					Type:  "filepattern",
					Paths: []string{"test"},
				},
			},
			dir:     "dirDoesNotExist",
			want:    emptyMap,
			wantErr: false,
		},
		{
			name: "Non-existent file",
			aliases: []o.Alias{
				{
					Type:  "filepattern",
					Paths: []string{"fileDoesNotExist"},
				},
			},
			dir:     tmpDir,
			want:    emptyMap,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliases, err := processFileContent(tt.aliases, tt.dir)
			assert.Equal(t, tt.want, aliases)
			if (err != nil) != tt.wantErr {
				t.Errorf("processFileContent error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_GenerateAliasesFromFilePattern(t *testing.T) {
	expectedAliases := []string{"WILD_FLAG", "WILD_FLAG_SECOND_ALIAS"}

	fileContents := map[string][]byte{
		"testdata/alias_test.txt":                                  []byte("SOME_FLAG = 'someFlag'"),
		"testdata/wild/alias_test.txt":                             []byte("WILD_FLAG = 'wildFlag'"),
		"testdata/wild/nested-wild/alias_test.txt":                 []byte("WILD_FLAG_SECOND_ALIAS = 'wildFlag'"),
		"testdata/wild/nested-wild/another/another/alias_test.txt": []byte("EVEN_WILDER = 'someFlag'"),
	}

	alias := fileWildPattern()

	aliases, err := GenerateAliasesFromFilePattern(alias, "wildFlag", "", fileContents)
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedAliases, aliases)
}
