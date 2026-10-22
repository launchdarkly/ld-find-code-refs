package options

import (
	"fmt"
	"os"
)

// AliasOptions contains the options specific to the alias command
type AliasOptions struct {
	Dir         string    `mapstructure:"dir"`
	ProjKey     string    `mapstructure:"projkey"`
	Projects    []Project `mapstructure:"projects"`
	AccessToken string    `mapstructure:"accessToken"`
	BaseUri     string    `mapstructure:"baseUri"`
	FlagKey     string    `mapstructure:"flagKey"`
	Debug       bool      `mapstructure:"debug"`
	UserAgent   string    `mapstructure:"userAgent"`
}

// ValidateAliasOptions validates the alias options based on the mode of operation
func (o *AliasOptions) ValidateAliasOptions() error {
	missingRequiredOptions := []string{}

	// Default Dir to current working directory if not provided
	if o.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current working directory: %w", err)
		}
		o.Dir = cwd
	}

	// If FlagKey is provided (Local Mode), AccessToken and ProjKey/Projects are not required
	if o.FlagKey != "" {
		// Local mode - dir is defaulted above if needed
		return nil
	}

	// If FlagKey is not provided (API-Connected Mode), AccessToken and ProjKey/Projects are required
	if o.AccessToken == "" {
		missingRequiredOptions = append(missingRequiredOptions, "accessToken")
	}
	if o.ProjKey == "" && len(o.Projects) == 0 {
		missingRequiredOptions = append(missingRequiredOptions, "projKey/projects")
	}

	if len(missingRequiredOptions) > 0 {
		return fmt.Errorf("missing required option(s): %v", missingRequiredOptions)
	}

	return nil
}