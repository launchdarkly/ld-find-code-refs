package main

import (
	"os"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
	"github.com/launchdarkly/ld-find-code-refs/internal/version"
	"github.com/launchdarkly/ld-find-code-refs/pkg/coderefs"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use: "ld-find-code-refs",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := o.InitYAML()
		if err != nil {
			return err
		}

		opts, err := o.GetOptions()
		if err != nil {
			return err
		}
		log.Init(opts.Debug)
		coderefs.Scan(opts)
		return nil
	},
	Version: version.Version,
}

func main() {
	err := o.Init(cmd.PersistentFlags())
	if err != nil {
		panic(err)
	}
	err = cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
