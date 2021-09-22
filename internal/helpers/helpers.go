package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

func Dedupe(s []string) []string {
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

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func FatalServiceError(err error, ignoreServiceErrors bool) {
	if ld.IsTransient(err) {
		if ignoreServiceErrors {
			os.Exit(0)
		}
		err = fmt.Errorf("%w\n Add the --ignoreServiceErrors flag to ignore this error", err)
	}
	log.Error.Fatal(err)
}
