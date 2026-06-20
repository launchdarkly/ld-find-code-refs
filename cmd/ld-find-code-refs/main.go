package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/launchdarkly/ld-find-code-refs/v2/coderefs"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/version"
	o "github.com/launchdarkly/ld-find-code-refs/v2/options"
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

var aliasCmd = &cobra.Command{
	Use:     "alias",
	Example: "ld-find-code-refs alias --flag-key my-flag",
	Short:   "Generate aliases for feature flags",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := o.InitYAMLForAlias()
		if err != nil {
			return err
		}

		opts, err := o.GetOptions()
		if err != nil {
			return err
		}

		// Get the flag-key value from the command
		flagKey, _ := cmd.Flags().GetString("flag-key")

		// Create AliasOptions from the global options
		aliasOpts := &o.AliasOptions{
			Dir:         opts.Dir,
			ProjKey:     opts.ProjKey,
			Projects:    opts.Projects,
			AccessToken: opts.AccessToken,
			BaseUri:     opts.BaseUri,
			FlagKey:     flagKey,
			Debug:       opts.Debug,
			UserAgent:   opts.UserAgent,
		}

		// Validate the alias options (not the global options)
		err = aliasOpts.ValidateAliasOptions()
		if err != nil {
			return err
		}

		log.Init(opts.Debug)
		return coderefs.GenerateAliases(*aliasOpts)
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
	
	// Add the flag-key flag to the alias command
	aliasCmd.Flags().String("flag-key", "", "Generate aliases for a specific flag key (local mode)")
	
	cmd.AddCommand(prune)
	cmd.AddCommand(extinctions)
	cmd.AddCommand(aliasCmd)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
