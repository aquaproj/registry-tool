package start

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/docker"
	"github.com/aquaproj/registry-tool/pkg/libc"
)

// Start starts Docker containers for the Linux and Windows environments.
// When the aggregated registry.yaml contains any variant with `key: libc`,
// the Alpine (musl) container is also started for libc-aware testing.
func Start(ctx context.Context, logger *slog.Logger, recreate bool) error {
	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := linuxDM.EnsureContainer(ctx, logger, recreate); err != nil {
		return fmt.Errorf("ensure Linux container: %w", err)
	}
	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := windowsDM.EnsureContainer(ctx, logger, recreate); err != nil {
		return fmt.Errorf("ensure Windows container: %w", err)
	}

	hasLibc, err := libc.HasVariant("registry.yaml")
	if err != nil {
		return fmt.Errorf("check libc variant: %w", err)
	}
	if hasLibc {
		alpineDM := docker.NewManager(docker.DefaultAlpineContainer())
		if err := alpineDM.EnsureContainer(ctx, logger, recreate); err != nil {
			return fmt.Errorf("ensure Alpine container: %w", err)
		}
	}
	return nil
}
