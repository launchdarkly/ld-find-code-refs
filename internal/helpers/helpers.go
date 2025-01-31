package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/bucketeer-io/code-refs/internal/log"
)

func Dedupe(s []string) []string {
	if len(s) <= 1 {
		return s
	}
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
	if ignoreServiceErrors {
		log.Error.Fatal(fmt.Errorf("%w\n Ignoring error and exiting", err))
		os.Exit(0)
	}
	log.Error.Fatal(err)
}
