package stop

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/docker"
)

// Stop stops Docker containers for the Linux and Windows environments.
func Stop(ctx context.Context, logger *slog.Logger) error {
	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := linuxDM.StopContainer(ctx, logger); err != nil {
		return fmt.Errorf("stop Linux container: %w", err)
	}
	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := windowsDM.StopContainer(ctx, logger); err != nil {
		return fmt.Errorf("stop Windows container: %w", err)
	}
	return nil
}
