package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
	"github.com/launchdarkly/ld-find-code-refs/internal/version"
	"github.com/launchdarkly/ld-find-code-refs/pkg/coderefs"
)

var prune = &cobra.Command{
	Use:     "prune [flags] branches...",
	Example: "ld-find-code-refs prune \"branch1\" \"branch2\" # prunes branch1 and branch2",
	Short:   "Delete stale code reference data stored in LaunchDarkly. Accepts stale branch names as arguments",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := o.InitYAML()
		if err != nil {
			return err
		}

		opts, err := o.GetOptions()
		if err != nil {
			return err
		}
		err = opts.ValidateRequired()
		if err != nil {
			return err
		}

		log.Init(opts.Debug)
		coderefs.Prune(opts, args)
		return nil
	},
}

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
		err = opts.Validate()
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
	cmd.AddCommand(prune)

	err = cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
