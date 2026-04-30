package stop

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/docker"
)

// Stop stops Docker containers for the Linux, Windows and Alpine environments.
// StopContainer is a no-op when the container is not running, so the Alpine
// container is stopped unconditionally to clean up regardless of variant state.
func Stop(ctx context.Context, logger *slog.Logger) error {
	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := linuxDM.StopContainer(ctx, logger); err != nil {
		return fmt.Errorf("stop Linux container: %w", err)
	}
	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := windowsDM.StopContainer(ctx, logger); err != nil {
		return fmt.Errorf("stop Windows container: %w", err)
	}
	alpineDM := docker.NewManager(docker.DefaultAlpineContainer())
	if err := alpineDM.StopContainer(ctx, logger); err != nil {
		return fmt.Errorf("stop Alpine container: %w", err)
	}
	return nil
}
