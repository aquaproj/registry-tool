package scaffold

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/aquaproj/registry-tool/pkg/initcmd"
)

const (
	dirPermission  os.FileMode = 0o775
	filePermission os.FileMode = 0o644
)

// Scaffold is the main entry point for the scaffold command.
// It supports both local mode (simple, no Docker) and full mode (Docker-based with tests).
func Scaffold(ctx context.Context, cfg *Config) error {
	if cfg.PkgName == "" {
		return errors.New(`usage: $ aqua-registry scaffold <pkgname>
e.g. $ aqua-registry scaffold cli/cli`)
	}

	// Strip https://github.com/ prefix if present
	cfg.PkgName = strings.TrimPrefix(cfg.PkgName, "https://github.com/")

	if cfg.Local {
		return scaffoldLocal(ctx, cfg)
	}
	return scaffoldFull(ctx, cfg)
}

// ScaffoldLegacy is the legacy entry point for backward compatibility.
func ScaffoldLegacy(ctx context.Context, cmds string, limit int, pkgNames ...string) error {
	if len(pkgNames) != 1 {
		return errors.New(`usage: $ aqua-registry scaffold <pkgname>
e.g. $ aqua-registry scaffold cli/cli`)
	}
	cfg := &Config{
		PkgName: pkgNames[0],
		Cmds:    cmds,
		Limit:   limit,
		Local:   true,
	}
	return Scaffold(ctx, cfg)
}

// scaffoldLocal runs the simple local scaffold without Docker.
func scaffoldLocal(ctx context.Context, cfg *Config) error {
	pkgName := cfg.PkgName
	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)
	pkgFile := filepath.Join(pkgDir, "pkg.yaml")
	rgFile := filepath.Join(pkgDir, "registry.yaml")

	if err := os.MkdirAll(pkgDir, dirPermission); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	if err := aquaGR(ctx, pkgName, pkgFile, rgFile, cfg.Cmds, cfg.Limit, cfg.ConfigPath); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Update registry.yaml")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
	}

	if err := initcmd.Init(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// scaffoldFull runs the full scaffold workflow with Docker containers and tests.
func scaffoldFull(ctx context.Context, cfg *Config) error {
	pkgName := cfg.PkgName

	// Step 1: Check prerequisites
	fmt.Fprintln(os.Stderr, "[Step 1/10] Checking prerequisites...")
	if err := CheckPrerequisites(ctx); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// Step 2: Check pkgs directory for uncommitted changes
	fmt.Fprintln(os.Stderr, "[Step 2/10] Checking for uncommitted changes in pkgs directory...")
	if err := CheckPkgsDiff(ctx); err != nil {
		return fmt.Errorf("pkgs directory check failed: %w", err)
	}

	// Step 3: Create/switch to feature branch
	fmt.Fprintln(os.Stderr, "[Step 3/10] Setting up git branch...")
	if err := GitCheckout(ctx, pkgName, cfg.NoCreateBranch); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// Step 4: Ensure Linux container is running
	fmt.Fprintln(os.Stderr, "[Step 4/10] Starting Linux container...")
	linuxContainer := DefaultLinuxContainer()
	linuxDM := NewDockerManager(linuxContainer)
	if err := linuxDM.EnsureContainer(ctx, cfg.Recreate); err != nil {
		return fmt.Errorf("failed to ensure Linux container: %w", err)
	}

	// Step 5: Run aqua gr in container
	fmt.Fprintln(os.Stderr, "[Step 5/10] Running scaffold in container...")
	if err := scaffoldInContainer(ctx, linuxDM, cfg); err != nil {
		return fmt.Errorf("scaffold in container failed: %w", err)
	}

	// Step 6: Update registry.yaml
	fmt.Fprintln(os.Stderr, "[Step 6/10] Updating registry.yaml...")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
	}

	// Step 7: Git commit
	fmt.Fprintln(os.Stderr, "[Step 7/10] Committing changes...")
	if err := GitCommit(ctx, pkgName); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Step 8: Run Linux/Darwin tests
	fmt.Fprintln(os.Stderr, "[Step 8/10] Running Linux/Darwin tests...")
	if err := RunLinuxDarwinTests(ctx, linuxDM, pkgName); err != nil {
		return fmt.Errorf("Linux/Darwin tests failed: %w", err)
	}

	// Step 9: Ensure Windows container is running
	fmt.Fprintln(os.Stderr, "[Step 9/10] Starting Windows container...")
	windowsContainer := DefaultWindowsContainer()
	windowsDM := NewDockerManager(windowsContainer)
	if err := windowsDM.EnsureContainer(ctx, cfg.Recreate); err != nil {
		return fmt.Errorf("failed to ensure Windows container: %w", err)
	}

	// Step 10: Run Windows tests
	fmt.Fprintln(os.Stderr, "[Step 10/10] Running Windows tests...")
	if err := RunWindowsTests(ctx, windowsDM, pkgName); err != nil {
		return fmt.Errorf("Windows tests failed: %w", err)
	}

	fmt.Fprintln(os.Stderr, "\n[SUCCESS] Scaffold completed successfully!")
	return nil
}

// scaffoldInContainer runs aqua gr inside the Docker container.
func scaffoldInContainer(ctx context.Context, dm *DockerManager, cfg *Config) error {
	pkgName := cfg.PkgName
	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)

	// Create local package directory
	if err := os.MkdirAll(pkgDir, dirPermission); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	// Copy scaffold config if provided
	if cfg.ConfigPath != "" {
		if err := dm.CopyTo(ctx, cfg.ConfigPath, dm.config.WorkingDir+"/scaffold.yaml"); err != nil {
			return fmt.Errorf("copy scaffold config to container: %w", err)
		}
	} else {
		// Check if pkgs/{pkg}/scaffold.yaml exists
		scaffoldConfig := filepath.Join(pkgDir, "scaffold.yaml")
		if _, err := os.Stat(scaffoldConfig); err == nil {
			fmt.Fprintf(os.Stderr, "[INFO] Using %s\n", scaffoldConfig)
			if err := dm.CopyTo(ctx, scaffoldConfig, dm.config.WorkingDir+"/scaffold.yaml"); err != nil {
				return fmt.Errorf("copy scaffold config to container: %w", err)
			}
		}
	}

	// Remove old pkg.yaml if exists
	_ = dm.ExecBash(ctx, "rm pkg.yaml 2>/dev/null || :")

	// Create registry.yaml with schema comment
	if err := dm.ExecBash(ctx, `echo '# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/registry.json' > registry.yaml`); err != nil {
		return fmt.Errorf("create registry.yaml header: %w", err)
	}

	// Build aqua gr command
	grCmd := "aqua gr"
	if cfg.Cmds != "" {
		grCmd += " -cmd " + cfg.Cmds
	}
	if cfg.Limit != 0 {
		grCmd += " -limit " + strconv.Itoa(cfg.Limit)
	}
	if cfg.ConfigPath != "" || fileExists(filepath.Join(pkgDir, "scaffold.yaml")) {
		grCmd += " -c scaffold.yaml"
	}
	grCmd += " --out-testdata pkg.yaml"
	grCmd += fmt.Sprintf(" %q >> registry.yaml", pkgName)

	// Run aqua gr in container
	if err := dm.ExecBash(ctx, grCmd); err != nil {
		return fmt.Errorf("aqua gr in container: %w", err)
	}

	// Copy results back from container
	if err := dm.CopyFrom(ctx, dm.config.WorkingDir+"/pkg.yaml", filepath.Join(pkgDir, "pkg.yaml")); err != nil {
		return fmt.Errorf("copy pkg.yaml from container: %w", err)
	}
	if err := dm.CopyFrom(ctx, dm.config.WorkingDir+"/registry.yaml", filepath.Join(pkgDir, "registry.yaml")); err != nil {
		return fmt.Errorf("copy registry.yaml from container: %w", err)
	}

	// Copy scaffold config to package directory if provided
	if cfg.ConfigPath != "" {
		if err := copyFile(cfg.ConfigPath, filepath.Join(pkgDir, "scaffold.yaml")); err != nil {
			return fmt.Errorf("copy scaffold config to pkgs: %w", err)
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func aquaGR(ctx context.Context, pkgName, pkgFilePath, rgFilePath, cmds string, limit int, configPath string) error {
	outFile, err := os.Create(rgFilePath)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", rgFilePath, err)
	}
	defer outFile.Close()

	if _, err := outFile.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/registry.json\n"); err != nil {
		return fmt.Errorf("write a code comment for yaml-language-server: %w", err)
	}

	command := "+ aqua gr --out-testdata " + pkgFilePath
	args := []string{"gr", "-out-testdata", pkgFilePath}

	if cmds != "" {
		args = append(args, "-cmd", cmds)
		command += " -cmd " + cmds
	}
	if limit != 0 {
		s := strconv.Itoa(limit)
		args = append(args, "-limit", s)
		command += " -limit " + s
	}
	if configPath != "" {
		args = append(args, "-c", configPath)
		command += " -c " + configPath
	}

	fmt.Fprintf(os.Stderr, "%s %s > %s\n", command, pkgName, rgFilePath)

	cmd := exec.CommandContext(ctx, "aqua", append(args, pkgName)...) //nolint:gosec
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %w", err)
	}

	return nil
}
