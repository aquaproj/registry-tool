package start

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/docker"
)

// Start starts Docker containers for the Linux and Windows environments.
func Start(ctx context.Context, logger *slog.Logger, recreate bool) error {
	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := linuxDM.EnsureContainer(ctx, logger, recreate); err != nil {
		return fmt.Errorf("ensure Linux container: %w", err)
	}
	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := windowsDM.EnsureContainer(ctx, logger, recreate); err != nil {
		return fmt.Errorf("ensure Windows container: %w", err)
	}
	return nil
}
