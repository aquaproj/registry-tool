package removepackage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/docker"
	"github.com/aquaproj/registry-tool/pkg/scaffold"
)

// RemovePackage removes a package from Docker containers.
func RemovePackage(ctx context.Context, logger *slog.Logger, pkgName string) error {
	pkg, err := resolvePkgName(ctx, logger, pkgName)
	if err != nil {
		return err
	}

	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := removeFromContainer(ctx, logger, linuxDM, pkg); err != nil {
		return fmt.Errorf("remove package from Linux container: %w", err)
	}

	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := removeFromContainer(ctx, logger, windowsDM, pkg); err != nil {
		return fmt.Errorf("remove package from Windows container: %w", err)
	}

	return nil
}

func removeFromContainer(ctx context.Context, logger *slog.Logger, dm *docker.Manager, pkg string) error {
	if err := dm.Exec(ctx, logger, nil, "aqua", "rm", pkg); err != nil {
		return fmt.Errorf("aqua rm: %w", err)
	}
	if err := dm.ExecBash(ctx, logger, "! test -f aqua-checksums.json || rm aqua-checksums.json"); err != nil {
		return fmt.Errorf("remove aqua-checksums.json: %w", err)
	}
	return nil
}

func resolvePkgName(ctx context.Context, logger *slog.Logger, pkgName string) (string, error) {
	if pkgName != "" {
		return strings.TrimPrefix(pkgName, "https://github.com/"), nil
	}

	branch, err := scaffold.GetCurrentBranch(ctx, logger)
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}

	if !strings.HasPrefix(branch, "feat/") {
		return "", errors.New("current branch must be feat/<package name> or you must give a package name")
	}

	return strings.TrimPrefix(branch, "feat/"), nil
}
