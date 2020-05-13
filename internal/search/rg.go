package search

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
)

// Code reference search using rg (ripgrep): https://github.com/BurntSushi/ripgrep
// We configure ag to run with the following flags:
// --case-sensitive as feature flags and aliases are case sensitive
// --pcre2          as the search pattern is a regular expression
// --json           to receive matches in jsonlines format
// -C               for providing context lines
// --sort="path"    to ensure files are sorted by path, as expected by search.go#aggregateByPath

/* example
$ rg --case-sensitive --pcre2 --json  -C1 ${search.go#generateSearchPattern} ${dir}
> {"type":"begin","data":{"path":{"text":"./test/index.ts"}}}
> {"type":"context","data":{"path":{"text":"./test/index.ts"},"lines":{"text":"// export my flag\n"},"line_number":2,"absolute_offset":3,"submatches":[]}}
> {"type":"match","data":{"path":{"text":"./test/index.ts"},"lines":{"text":"export const isMyFlagEnabled = flagPredicate('my-flag');\n"},"line_number":3,"absolute_offset":21,"submatches":[{"match":{"text":"my-flag"},"start":46,"end":53}]}}
> {"type":"context","data":{"path":{"text":"./test/index.ts"},"lines":{"text":"console.log('exported my flag');\n"},"line_number":4,"absolute_offset":78,"submatches":[]}}
> {"type":"end","data":{"path":{"text":"./test/index.ts"},"binary_offset":null,"stats":{"elapsed":{"secs":0,"nanos":73500,"human":"0.000073s"},"searches":1,"searches_with_match":1,"bytes_searched":111,"bytes_printed":621,"matched_lines":1,"matches":1}}}
> {"type":"begin","data":{"path":{"text":"./test/flags.go"}}}
> {"type":"context","data":{"path":{"text":"./test/flags.go"},"lines":{"text":"\tAnotherFlag = \"another-flag\"\n"},"line_number":5,"absolute_offset":4,"submatches":[]}}
> {"type":"match","data":{"path":{"text":"./test/flags.go"},"lines":{"text":"\tMyFlag = \"my-flag\"\n"},"line_number":6,"absolute_offset":34,"submatches":[{"match":{"text":"my-flag"},"start":11,"end":18}]}}
> {"type":"context","data":{"path":{"text":"./test/flags.go"},"lines":{"text":"\tYetAnotherFlag = \"yet-another-flag\"\n"},"line_number":7,"absolute_offset":54,"submatches":[]}}
> {"type":"end","data":{"path":{"text":"./test/flags.go"},"binary_offset":null,"stats":{"elapsed":{"secs":0,"nanos":48251,"human":"0.000048s"},"searches":1,"searches_with_match":1,"bytes_searched":91,"bytes_printed":609,"matched_lines":1,"matches":1}}}
> {"data":{"elapsed_total":{"human":"0.022576s","nanos":22576397,"secs":0},"stats":{"bytes_printed":1230,"bytes_searched":202,"elapsed":{"human":"0.000122s","nanos":121751,"secs":0},"matched_lines":2,"matches":2,"searches":2,"searches_with_match":2}},"type":"summary"}
*/

type RgClient struct {
	workspace string
}

func NewRgClient(path string) (*RgClient, error) {
	if !filepath.IsAbs(path) {
		log.Error.Fatalf("expected an absolute path but received a relative path: %s", path)
	}
	_, err := exec.LookPath("rg")
	if err != nil {
		return nil, errors.New("rg (ripgrep) is a required dependency, but was not found in the system PATH")
	}

	return &RgClient{workspace: path}, nil
}

func (c *RgClient) searchForRefs(flags []string, aliases map[string][]string, ctxLines int, delimiters []byte) (SearchResultLines, error) {
	args := []string{"--case-sensitive", "--json", "--pcre2", "--sort=path"}
	pathToIgnore := filepath.Join(c.workspace, ignoreFileName)
	if validation.FileExists(pathToIgnore) {
		log.Debug.Printf("excluding files matched in %s", ignoreFileName)
		args = append(args, fmt.Sprintf("--ignore-file=%s", pathToIgnore))
	}
	if ctxLines > 0 {
		args = append(args, fmt.Sprintf("-C%d", ctxLines))
	}

	searchPattern := generateSearchPattern(flags, delimiters, runtime.GOOS == windows)
	/* #nosec */
	cmd := exec.Command("rg", args...)
	cmd.Args = append(cmd.Args, searchPattern, c.workspace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	jsonLines := rgLines{}
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		jsonLine := rgLine{}
		err := dec.Decode(&jsonLine)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		jsonLines = append(jsonLines, jsonLine)
	}
	return jsonLines.toSearchResultLines(aliases, c.workspace+"/")
}

type rgText struct {
	Text *string `json:"text,omitempty"`
}

type rgLines []rgLine

// toSearchResultLines converts ripgrep output to the standardized format used for search results
func (l rgLines) toSearchResultLines(aliases map[string][]string, pathPrefix string) (SearchResultLines, error) {
	ret := make(SearchResultLines, 0, len(l))

	// invert the map of flag keys to aliases for use when associating a matched alias to a flag key
	aliasesToFlagKeys := map[string]string{}
	for flagKey, flagAliases := range aliases {
		for _, alias := range flagAliases {
			aliasesToFlagKeys[alias] = flagKey
		}
	}

	for _, line := range l {
		switch line.Type {
		// other types of lines are not useful for our purposes
		case "match", "context":
			srl, err := line.toSearchResultLine(aliases, aliasesToFlagKeys, pathPrefix)
			if err != nil {
				return ret, err
			}
			ret = append(ret, *srl)
		}
	}

	return ret, nil
}

type rgLine struct {
	Type string `json:"type"`
	Data struct {
		Path       *rgText `json:"path,omitempty"`
		Lines      *rgText `json:"lines,omitempty"`
		LineNumber *int    `json:"line_number,omitempty"`
		Submatches []struct {
			Match *rgText `json:"match,omitempty"`
		} `json:"submatches,omitempty"`
	} `json:"data,omitempty"`
}

func (j rgLine) err() error {
	return fmt.Errorf("error converting ripgrep output to code reference:\n%#v", j)
}

func (j rgLine) toSearchResultLine(aliases map[string][]string, aliasesToFlagKeys map[string]string, pathPrefix string) (*SearchResultLine, error) {
	data := j.Data

	// unexpected ripgrep output
	if data.Path == nil || data.Path.Text == nil || data.LineNumber == nil || data.Lines == nil || data.Lines.Text == nil {
		return nil, j.err()
	}

	line := SearchResultLine{
		Path:     strings.TrimPrefix(*data.Path.Text, pathPrefix),
		LineNum:  *data.LineNumber,
		LineText: *data.Lines.Text,
	}

	flagKeys := map[string][]string{}

	for _, match := range data.Submatches {
		matchedTerm := match.Match.Text
		_, isFlagKey := aliases[*matchedTerm]

		key := *matchedTerm
		if !isFlagKey {
			key = aliasesToFlagKeys[key]
		}

		if flagKeys[key] == nil {
			flagKeys[key] = []string{}
		}

		if !isFlagKey {
			flagKeys[key] = append(flagKeys[key], *matchedTerm)
		}
	}
	line.FlagKeys = flagKeys
	return &line, nil
}
