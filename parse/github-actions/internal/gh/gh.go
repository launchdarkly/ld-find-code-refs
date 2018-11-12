package gh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	jsonpointer "github.com/mattn/go-jsonpointer"
)

func ParseEvent(path string) map[string]interface{} {
	eventJsonFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	eventJsonBytes, err := ioutil.ReadAll(eventJsonFile)
	if err != nil {
		panic(err)
	}

	var eventMap map[string]interface{}
	if err = json.Unmarshal(eventJsonBytes, &eventMap); err != nil {
		panic(err)
	}

	// List of fields to pull out into the top level
	fields := []string{
		"repository/url",
		"repository/default_branch",
		"repository/pushed_at",
		"sender/login",
	}
	ret := map[string]interface{}{}

Parse:
	for _, field := range fields {
		val, err := jsonpointer.Get(eventMap, "/"+field)
		if err != nil {
			fmt.Printf("Error parsing field: %s\n%+v\n", field, err)
			continue Parse
		}
		pointer := strings.Replace(field, "/", "_", -1)
		ret[pointer] = val
	}

	return ret
}
