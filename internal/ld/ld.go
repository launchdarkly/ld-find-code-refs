package ld

import (
	"context"

	ldapi "github.com/launchdarkly/api-client-go"
	log "github.com/sirupsen/logrus"
)

type ApiClient struct {
	client *ldapi.APIClient
	apiKey string
}

type ApiOptions struct {
	ApiKey  string
	BaseUri string
}

func InitApiClient(options ApiOptions) ApiClient {
	if options.BaseUri == "" {
		options.BaseUri = "https://app.launchdarkly.com"
	}
	return ApiClient{
		client: ldapi.NewAPIClient(&ldapi.Configuration{
			BasePath:  options.BaseUri + "/api/v2",
			UserAgent: "github-actor",
		}),
		apiKey: options.ApiKey,
	}
}

func (service ApiClient) GetFlagKeyList(projectKey string) ([]string, error) {
	ctx := context.WithValue(context.Background(), ldapi.ContextAPIKey, ldapi.APIKey{Key: service.apiKey})
	flags, _, err := service.client.FeatureFlagsApi.GetFeatureFlags(ctx, projectKey, nil)
	if err != nil {
		return nil, err
	}
	flagKeys := make([]string, 0, len(flags.Items))
	for _, flag := range flags.Items {
		flagKeys = append(flagKeys, flag.Key)
	}
	return flagKeys, nil
}

func (service ApiClient) PutCodeReferenceBranch([]byte) error {
	log.Debug("STUBBED PutCodeReferenceBranch")
	return nil
}
