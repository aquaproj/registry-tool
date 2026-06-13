package github

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/suzuki-shunsuke/ghtkn-go-sdk/ghtkn"
)

// GetAccessToken retrieves the GitHub token from environment or gh CLI.
func GetAccessToken(ctx context.Context, logger *slog.Logger) (string, error) {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	ghtknEnabled, err := ghtkn.Enabled(&ghtkn.InputEnabled{
		Envs: []string{
			"AQUA_GHTKN_ENABLED",
		},
	})
	if err != nil {
		return "", fmt.Errorf("check ghtkn enabled: %w", err)
	}
	if !ghtknEnabled {
		return "", nil
	}
	client, err := ghtkn.New()
	if err != nil {
		return "", fmt.Errorf("create ghtkn client: %w", err)
	}
	token, _, err := client.Get(ctx, logger, &ghtkn.InputGet{})
	if err != nil {
		return "", fmt.Errorf("get a github access token by ghtkn SDK: %w", err)
	}
	return token.AccessToken, nil
}
