package helpers

import (
	"fmt"

	"github.com/bucketeer-io/code-refs/internal/version"
)

func GetUserAgent(u string) string {
	userAgent := "BucketeerFindCodeRefs/" + version.Version
	if u != "" {
		userAgent = fmt.Sprintf("%s (CI %s)", userAgent, u)
	}

	return userAgent
}
