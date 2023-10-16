package aliases

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/v2/options"
)

var allNamingConventions = []o.Alias{
	alias(o.CamelCase),
	alias(o.PascalCase),
	alias(o.SnakeCase),
	alias(o.UpperSnakeCase),
	alias(o.KebabCase),
	alias(o.DotCase),
}

var allSomeFlagNamingConventionAliases = slice("anyKindOfKey", "AnyKindOfKey", "any_kind_of_key", "ANY_KIND_OF_KEY", "any-kind-of-key", "any.kind.of.key")

const (
	testFlagAlias    = "test-flag"
	testFlagKey      = "someFlag"
	testFlagKey2     = "anotherFlag"
	testFlagAliasKey = "AnyKind.of_key"
	testWildFlagKey  = "wildFlag"
)

func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func Test_GenerateAliases(t *testing.T) {
	specs := []struct {
		name    string
		flags   []string
		aliases []o.Alias
		want    map[string][]string
	}{
		{
			name:  "literals",
			flags: slice(testFlagAliasKey),
			aliases: []o.Alias{
				literal(slice(testFlagAliasKey)),
			},
			want: makeLiteralAliases(slice(testFlagAliasKey)),
		},
		{
			name:    "naming conventions",
			flags:   slice(testFlagAliasKey),
			aliases: allNamingConventions,
			want:    map[string][]string{testFlagAliasKey: allSomeFlagNamingConventionAliases},
		},
		{
			name:  "two flags",
			flags: slice(testFlagKey, testFlagKey2),
			aliases: []o.Alias{
				alias(o.PascalCase),
			},
			want: map[string][]string{testFlagKey: slice("SomeFlag"), testFlagKey2: slice("AnotherFlag")},
		},
		{
			name:  "duplicate alias types",
			flags: slice(testFlagKey),
			aliases: []o.Alias{
				alias(o.PascalCase),
				alias(o.PascalCase),
			},
			want: map[string][]string{testFlagKey: slice("SomeFlag")},
		},
		{
			name:  "PascalCase alias",
			flags: slice("some_flag"),
			aliases: []o.Alias{
				alias(o.PascalCase),
			},
			want: map[string][]string{"some_flag": slice("SomeFlag")},
		},
		{
			name:  "screaming snake case to camel case",
			flags: slice("SOME_FLAG"),
			aliases: []o.Alias{
				alias(o.CamelCase),
			},
			want: map[string][]string{"SOME_FLAG": slice("someFlag")},
		},
		{
			name:  "file exact pattern",
			flags: slice(testFlagKey),
			aliases: []o.Alias{
				fileExactPattern(testFlagKey),
			},
			want: map[string][]string{testFlagKey: slice("SOME_FLAG")},
		},
		{
			name:  "file wildcard pattern",
			flags: slice(testFlagKey, testWildFlagKey),
			aliases: []o.Alias{
				fileWildPattern(testFlagKey),
			},
			want: map[string][]string{testWildFlagKey: slice("WILD_FLAG", "WILD_FLAG_SECOND_ALIAS"), testFlagKey: slice("SOME_FLAG")},
		},
		{
			name:  "command",
			flags: slice(testFlagKey),
			aliases: []o.Alias{
				cmd(`echo ["SOME_FLAG"]`, 0),
			},
			want: map[string][]string{testFlagKey: slice("SOME_FLAG")},
		},
		{
			name:  "command, custom pascalcase",
			flags: slice("some_flag"),
			aliases: []o.Alias{
				cmd(`python3 ./aliases-test.py`, 5),
			},
			want: map[string][]string{"some_flag": slice("Some_Flag")},
		},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			aliases, err := GenerateAliases(tt.flags, tt.aliases, "")
			require.NoError(t, err)
			assert.Equal(t, tt.want, aliases)
		})
	}
}

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

func slice(args ...string) []string {
	return args
}

var literalAliases = []string{"abc", "def"}

func makeLiteralAliases(flags []string) map[string][]string {
	ret := map[string][]string{}
	for _, f := range flags {
		ret[f] = literalAliases
	}
	return ret
}

func alias(t o.AliasType) o.Alias {
	return o.Alias{Type: t}
}

func literal(flags []string) o.Alias {
	a := alias(o.Literal)
	a.Flags = makeLiteralAliases(flags)
	return a
}

func fileExactPattern(flag string) o.Alias {
	a := alias(o.FilePattern)
	pattern := "(\\w+)\\s= 'FLAG_KEY'"
	a.Paths = []string{"testdata/alias_test.txt"}
	a.Patterns = []string{pattern}
	return a
}

func fileWildPattern(flag string) o.Alias {
	a := alias(o.FilePattern)
	pattern := "(\\w+)\\s= 'FLAG_KEY'"
	a.Paths = []string{"testdata/**/*.txt"}
	a.Patterns = []string{pattern}
	return a
}

func cmd(command string, timeout int64) o.Alias {
	a := alias(o.Command)
	a.Command = &command
	a.Timeout = &timeout
	return a
}
