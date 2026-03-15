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
	if os.Getenv("AQUA_GHTKN_ENABLED") != "true" {
		return "", nil
	}
	client := ghtkn.New()
	token, _, err := client.Get(ctx, logger, &ghtkn.InputGet{})
	if err != nil {
		return "", fmt.Errorf("get a github access token by ghtkn SDK: %w", err)
	}
	return token.AccessToken, nil
}
