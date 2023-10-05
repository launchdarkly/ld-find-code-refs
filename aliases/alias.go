package aliases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/iancoleman/strcase"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/validation"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

// GenerateAliases returns a map of flag keys to aliases based on config.
func GenerateAliases(flags []string, aliases []options.Alias, dir string) (map[string][]string, error) {
	allFileContents, err := processFileContent(aliases, dir)
	if err != nil {
		return nil, err
	}

	ret := make(map[string][]string, len(flags))
	for _, flag := range flags {
		for i, a := range aliases {
			if a.Name == "" {
				a.Name = strconv.Itoa(i)
			}
			flagAliases, err := generateAlias(a, flag, dir, allFileContents)
			if err != nil {
				return nil, err
			}
			ret[flag] = append(ret[flag], flagAliases...)
		}
		ret[flag] = helpers.Dedupe(ret[flag])
	}
	return ret, nil
}

func generateAlias(a options.Alias, flag, dir string, allFileContents map[string][]byte) (ret []string, err error) {
	switch a.Type.Canonical() {
	case options.Literal:
		ret = a.Flags[flag]
	case options.CamelCase:
		ret = []string{strcase.ToLowerCamel(flag)}
	case options.PascalCase:
		ret = []string{strcase.ToCamel(flag)}
	case options.SnakeCase:
		ret = []string{strcase.ToSnake(flag)}
	case options.UpperSnakeCase:
		ret = []string{strcase.ToScreamingSnake(flag)}
	case options.KebabCase:
		ret = []string{strcase.ToKebab(flag)}
	case options.DotCase:
		ret = []string{strcase.ToDelimited(flag, '.')}
	case options.FilePattern:
		ret, err = generateAliasesFromFilePattern(a, flag, dir, allFileContents)
	case options.Command:
		ret, err = generateAliasesFromCommand(a, flag, dir)
	}

	return ret, err
}

func generateAliasesFromFilePattern(a options.Alias, flag, dir string, allFileContents map[string][]byte) ([]string, error) {
	ret := []string{}
	// Concatenate the contents of all files into a single byte array to be matched by specified patterns
	fileContents := []byte{}
	for _, path := range a.Paths {
		absGlob := filepath.Join(dir, path)
		matches, err := doublestar.FilepathGlob(absGlob)
		if err != nil {
			return nil, fmt.Errorf("filepattern '%s': could not process path glob '%s'", a.Name, absGlob)
		}
		for _, match := range matches {
			if pathFileContents := allFileContents[match]; len(pathFileContents) > 0 {
				fileContents = append(fileContents, pathFileContents...)
			}
		}
	}

	for _, p := range a.Patterns {
		pattern := regexp.MustCompile(strings.ReplaceAll(p, "FLAG_KEY", flag))
		results := pattern.FindAllStringSubmatch(string(fileContents), -1)
		for _, res := range results {
			if len(res) > 1 {
				ret = append(ret, res[1:]...)
			}
		}
	}

	return ret, nil
}

func generateAliasesFromCommand(a options.Alias, flag, dir string) ([]string, error) {
	ret := []string{}
	ctx := context.Background()
	if a.Timeout != nil && *a.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*time.Duration(*a.Timeout)))
		defer cancel()
	}
	tokens := strings.Split(*a.Command, " ")
	name := tokens[0]
	args := []string{}
	if len(tokens) > 1 {
		args = tokens[1:]
	}
	/* #nosec */
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(flag)
	cmd.Dir = dir
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command '%s': failed to execute alias command: %w", a.Name, err)
	}
	if err := json.Unmarshal(stdout, &ret); err != nil {
		return nil, fmt.Errorf("command '%s': could not unmarshal json output of alias command: %w", a.Name, err)
	}

	return ret, err
}

// processFileContent reads and stores the content of files specified by filePattern alias matchers to be matched for aliases
func processFileContent(aliases []options.Alias, dir string) (map[string][]byte, error) {
	allFileContents := map[string][]byte{}
	for idx, a := range aliases {
		if a.Type.Canonical() != options.FilePattern {
			continue
		}

		aliasId := strconv.Itoa(idx)
		if a.Name != "" {
			aliasId = a.Name
		}

		paths := []string{}
		for _, glob := range a.Paths {
			absGlob := filepath.Join(dir, glob)
			matches, err := doublestar.FilepathGlob(absGlob)
			if err != nil {
				return nil, fmt.Errorf("filepattern '%s': could not process path glob '%s'", aliasId, absGlob)
			}
			if matches == nil {
				log.Info.Printf("filepattern '%s': no matching files found for alias path glob '%s'", aliasId, absGlob)
			}
			paths = append(paths, matches...)
		}
		paths = helpers.Dedupe(paths)

		for _, path := range paths {
			_, pathAlreadyProcessed := allFileContents[path]
			if pathAlreadyProcessed {
				continue
			}

			if !validation.FileExists(path) {
				return nil, fmt.Errorf("filepattern '%s': could not find file at path '%s'", aliasId, path)
			}
			/* #nosec */
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("filepattern '%s': could not process file at path '%s': %v", aliasId, path, err)
			}
			allFileContents[path] = data
		}
	}
	return allFileContents, nil
}
