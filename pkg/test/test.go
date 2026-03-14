package test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/docker"
	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/aquaproj/registry-tool/pkg/scaffold"
)

// Config holds configuration for the test command.
type Config struct {
	PkgName  string
	Recreate bool
}

// Test tests a package in Docker containers across all platforms.
func Test(ctx context.Context, logger *slog.Logger, cfg *Config) error {
	pkgName, err := resolvePkgName(ctx, logger, cfg.PkgName)
	if err != nil {
		return err
	}

	// Ensure Linux container
	linuxDM := docker.NewManager(docker.DefaultLinuxContainer())
	if err := linuxDM.EnsureContainer(ctx, logger, cfg.Recreate); err != nil {
		return fmt.Errorf("ensure Linux container: %w", err)
	}

	// Run Linux/Darwin tests
	logger.Info("Running Linux/Darwin tests")
	if err := scaffold.RunLinuxDarwinTests(ctx, logger, linuxDM, pkgName); err != nil {
		return fmt.Errorf("Linux/Darwin tests failed: %w", err)
	}

	// Ensure Windows container
	windowsDM := docker.NewManager(docker.DefaultWindowsContainer())
	if err := windowsDM.EnsureContainer(ctx, logger, cfg.Recreate); err != nil {
		return fmt.Errorf("ensure Windows container: %w", err)
	}

	// Run Windows tests
	logger.Info("Running Windows tests")
	if err := scaffold.RunWindowsTests(ctx, logger, windowsDM, pkgName); err != nil {
		return fmt.Errorf("windows tests failed: %w", err)
	}

	// Update registry.yaml
	logger.Info("Updating registry.yaml")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
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
