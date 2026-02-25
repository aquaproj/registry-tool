package scaffold

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// RunTests runs aqua install tests on the specified platforms.
func RunTests(ctx context.Context, dm *DockerManager, pkgName string, platforms []Platform) error {
	pkgDir := filepath.Join("pkgs", pkgName)

	// Copy package files to container
	if err := dm.CopyTo(ctx, filepath.Join(pkgDir, "pkg.yaml"), dm.config.WorkingDir+"/pkg.yaml"); err != nil {
		return fmt.Errorf("copy pkg.yaml to container: %w", err)
	}
	if err := dm.CopyTo(ctx, filepath.Join(pkgDir, "registry.yaml"), dm.config.WorkingDir+"/registry.yaml"); err != nil {
		return fmt.Errorf("copy registry.yaml to container: %w", err)
	}

	for _, p := range platforms {
		fmt.Fprintf(os.Stderr, "\n[INFO] Testing %s/%s\n", p.OS, p.Arch)

		// Remove aqua-checksums.json before each test
		_ = dm.ExecBash(ctx, "rm aqua-checksums.json 2>/dev/null || :")

		env := map[string]string{
			"AQUA_GOOS":   p.OS,
			"AQUA_GOARCH": p.Arch,
		}

		if err := dm.Exec(ctx, env, "aqua", "i"); err != nil {
			return fmt.Errorf("test failed for %s/%s: %w", p.OS, p.Arch, err)
		}
	}

	return nil
}

// RunLinuxDarwinTests runs tests for Linux and Darwin platforms.
func RunLinuxDarwinTests(ctx context.Context, dm *DockerManager, pkgName string) error {
	return RunTests(ctx, dm, pkgName, LinuxDarwinPlatforms())
}

// RunWindowsTests runs tests for Windows platforms.
func RunWindowsTests(ctx context.Context, dm *DockerManager, pkgName string) error {
	return RunTests(ctx, dm, pkgName, WindowsPlatforms())
}
