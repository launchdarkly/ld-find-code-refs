package coderefs

import (
	"github.com/launchdarkly/ld-find-code-refs/internal/options"
)

func generateAliases(flags []string, aliases []options.Alias) (map[string][]string, error) {
	ret := make(map[string][]string, len(flags))
	for _, flag := range flags {
		for _, a := range aliases {
			flagAliases, err := a.Generate(flag)
			if err != nil {
				return nil, err
			}
			ret[flag] = append(ret[flag], flagAliases...)
		}
		ret[flag] = dedupe(ret[flag])
	}
	return ret, nil
}

func dedupe(s []string) []string {
	keys := make(map[string]struct{}, len(s))
	ret := make([]string, 0, len(s))
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = struct{}{}
			ret = append(ret, entry)
		}
	}
	return ret
}
