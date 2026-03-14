package remove

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/docker"
)

// Remove removes Docker containers for the Linux and Windows environments.
func Remove(ctx context.Context, logger *slog.Logger) error {
	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := linuxDM.RemoveContainer(ctx, logger); err != nil {
		return fmt.Errorf("remove Linux container: %w", err)
	}
	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := windowsDM.RemoveContainer(ctx, logger); err != nil {
		return fmt.Errorf("remove Windows container: %w", err)
	}
	return nil
}
