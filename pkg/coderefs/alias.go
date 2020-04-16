package coderefs

import (
	"github.com/launchdarkly/ld-find-code-refs/internal/options"
)

func generateAliases(flags []string, aliases []options.Alias) (map[string][]string, error) {
	ret := make(map[string][]string, len(flags))
	for _, a := range aliases {
		for _, flag := range flags {
			flagAliases, err := a.Generate(flag)
			if err != nil {
				return nil, err
			}
			ret[flag] = append(ret[flag], flagAliases...)
		}
	}
	return ret, nil
}
