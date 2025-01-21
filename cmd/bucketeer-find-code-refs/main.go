package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bucketeer-io/code-refs/coderefs"
	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/internal/version"
	o "github.com/bucketeer-io/code-refs/options"
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

var extinctions = &cobra.Command{
	Use:     "extinctions",
	Example: "ld-find-code-refs extinctions",
	Short:   "Find and Post extinctions for branch",
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
		coderefs.Run(opts, false)
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
		coderefs.Run(opts, true)
		return nil
	},
	Version: version.Version,
}

func main() {
	if err := o.Init(cmd.PersistentFlags()); err != nil {
		panic(err)
	}
	cmd.AddCommand(prune)
	cmd.AddCommand(extinctions)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
