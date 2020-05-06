package main

import (
	"os"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
	"github.com/launchdarkly/ld-find-code-refs/internal/version"
	"github.com/launchdarkly/ld-find-code-refs/pkg/coderefs"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "ld-find-code-refs",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return o.ValidateOptions()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Init(o.Debug)
		coderefs.Scan()
	},
	Version: version.Version,
}

func main() {
	err := o.Init(rootCmd)
	if err != nil {
		panic(err)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
