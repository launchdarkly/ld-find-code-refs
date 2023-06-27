package helpers

import (
	"fmt"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/version"
)

func GetUserAgent(u string) string {
	userAgent := "LDFindCodeRefs/" + version.Version
	if u != "" {
		userAgent = fmt.Sprintf("%s (CI %s)", userAgent, u)
	}

	return userAgent
}
