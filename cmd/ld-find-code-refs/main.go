package main

import (
	"os"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
	"github.com/launchdarkly/ld-find-code-refs/pkg/coderefs"
)

func main() {
	err, cb := o.Init()
	if err != nil {
		log.Init(false)
		log.Error.Printf("could not validate configuration: %s", err)
		if cb != nil {
			cb()
		}
		os.Exit(1)
	}
	log.Init(o.Debug.Value())
	coderefs.Scan()
}
