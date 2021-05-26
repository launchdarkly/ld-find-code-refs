package coderefs

import (
	"testing"

	o "github.com/launchdarkly/ld-find-code-refs/options"
	"github.com/stretchr/testify/assert"
)

var allNamingConventions = []o.Alias{
	alias(o.CamelCase),
	alias(o.PascalCase),
	alias(o.SnakeCase),
	alias(o.UpperSnakeCase),
	alias(o.KebabCase),
	alias(o.DotCase),
}

var allSomeFlagNamingConventionAliases = slice("anyKindOfKey", "AnyKindOfKey", "any_kind.of_key", "ANY_KIND.OF_KEY", "any-kind.of-key", "any.kind.of.key")

const (
	testFlagKey      = "someFlag"
	testFlagKey2     = "anotherFlag"
	testFlagAliasKey = "AnyKind.of_key"
	testWildFlagKey  = "wildFlag"
)

func Test_GenerateAliases(t *testing.T) {
	specs := []struct {
		name    string
		flags   []string
		aliases []o.Alias
		want    map[string][]string
		wantErr error
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
			name:  "file pattern",
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
			want: map[string][]string{testWildFlagKey: slice("WILD_FLAG"), testFlagKey: slice("SOME_FLAG")},
		},
		// TODO
		// {
		// 	name:    "command",
		// 	flags:   slice(testFlagKey),
		// 	aliases: []o.Alias{cmd(`echo '["SOME_FLAG"]'`, 0)},
		// },
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			aliases, err := GenerateAliases(tt.flags, tt.aliases, "")
			assert.Equal(t, tt.want, aliases)
			assert.Equal(t, tt.wantErr, err)
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
	a.Paths = []string{"testdata/*/*.txt", "testdata/*.txt"}
	a.Patterns = []string{pattern}
	return a
}

func cmd(command string, timeout int64) o.Alias {
	a := alias(o.Command)
	a.Command = &command
	a.Timeout = &timeout
	return a
}
