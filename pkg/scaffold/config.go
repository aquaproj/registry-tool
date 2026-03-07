package scaffold

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/suzuki-shunsuke/ghtkn-go-sdk/ghtkn"
)

// Config holds the configuration for the scaffold command.
type Config struct {
	// PkgName is the package name (e.g., "cli/cli")
	PkgName string
	// Cmds is a comma-separated list of commands to test
	Cmds string
	// Limit is the maximum number of versions to generate
	Limit int
	// Local runs in local mode without Docker
	Local bool
	// Recreate forces recreation of Docker containers
	Recreate bool
	// NoCreateBranch skips creating a git branch
	NoCreateBranch bool
	// ConfigPath is the path to scaffold.yaml config file
	ConfigPath string
}

// ContainerConfig holds Docker container configuration.
type ContainerConfig struct {
	Name       string
	Image      string
	WorkingDir string
}

const containerWorkingDir = "/workspace"

// DefaultLinuxContainer returns the default Linux container configuration.
func DefaultLinuxContainer() ContainerConfig {
	return ContainerConfig{
		Name:       "aqua-registry",
		Image:      "aquaproj/aqua-registry",
		WorkingDir: containerWorkingDir,
	}
}

// DefaultWindowsContainer returns the default Windows container configuration.
func DefaultWindowsContainer() ContainerConfig {
	return ContainerConfig{
		Name:       "aqua-registry-windows",
		Image:      "aquaproj/aqua-registry",
		WorkingDir: containerWorkingDir,
	}
}

// Platform represents a target platform for testing.
type Platform struct {
	OS   string
	Arch string
}

// LinuxDarwinPlatforms returns the Linux and Darwin platforms for testing.
func LinuxDarwinPlatforms() []Platform {
	return []Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
	}
}

// WindowsPlatforms returns the Windows platforms for testing.
func WindowsPlatforms() []Platform {
	return []Platform{
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "arm64"},
	}
}

// getGitHubToken retrieves the GitHub token from environment or gh CLI.
func getGitHubToken(ctx context.Context, logger *slog.Logger) (string, error) {
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
