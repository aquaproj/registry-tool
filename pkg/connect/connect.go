package connect

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/aquaproj/registry-tool/pkg/docker"
)

func Connect(ctx context.Context, logger *slog.Logger, osName, arch string) error {
	if osName == "" {
		osName = "linux"
	}
	if arch == "" {
		arch = runtime.GOARCH
	}

	var config docker.Config
	if osName == "windows" {
		config = docker.DefaultWindowsContainer()
	} else {
		config = docker.DefaultLinuxContainer()
	}

	dm := docker.NewManager(config)

	env := map[string]string{
		"AQUA_GOOS":   osName,
		"AQUA_GOARCH": arch,
	}

	// Run aqua i -l as a workaround
	if err := dm.Command(ctx, logger, env, "aqua", "i", "-l").Run(); err != nil {
		return fmt.Errorf("run aqua i -l: %w", err)
	}

	if err := dm.ExecInteractive(ctx, logger, env, "bash"); err != nil {
		return fmt.Errorf("run interactive bash: %w", err)
	}
	return nil
}
