package scaffold

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/docker"
	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/aquaproj/registry-tool/pkg/github"
	"github.com/aquaproj/registry-tool/pkg/initcmd"
	"github.com/aquaproj/registry-tool/pkg/osexec"
)

// Scaffold is the main entry point for the scaffold command.
// It supports both local mode (simple, no Docker) and full mode (Docker-based with tests).
func Scaffold(ctx context.Context, logger *slog.Logger, cfg *Config) error {
	if cfg.PkgName == "" {
		return errors.New(`usage: $ argd scaffold <pkgname>
e.g. $ argd scaffold cli/cli`)
	}

	// Strip https://github.com/ prefix if present
	cfg.PkgName = strings.TrimPrefix(cfg.PkgName, "https://github.com/")

	githubToken, err := github.GetAccessToken(ctx, logger)
	if err != nil {
		return err
	}

	if cfg.Local {
		return scaffoldLocal(ctx, logger, cfg)
	}
	return scaffoldFull(ctx, logger, cfg, githubToken)
}

// scaffoldLocal runs the simple local scaffold without Docker.
func scaffoldLocal(ctx context.Context, logger *slog.Logger, cfg *Config) error {
	pkgName := cfg.PkgName
	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)
	pkgFile := filepath.Join(pkgDir, "pkg.yaml")
	rgFile := filepath.Join(pkgDir, "registry.yaml")

	if err := os.MkdirAll(pkgDir, docker.DirPermission); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	if err := aquaGR(ctx, logger, pkgName, pkgFile, rgFile, cfg.Cmds, cfg.Limit, cfg.ConfigPath); err != nil {
		return err
	}

	logger.Info("updating registry.yaml")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
	}

	if err := initcmd.Init(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// scaffoldFull runs the full scaffold workflow with Docker containers and tests.
func scaffoldFull(ctx context.Context, logger *slog.Logger, cfg *Config, githubToken string) error {
	pkgName := cfg.PkgName

	if err := CheckPrerequisites(ctx, logger); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	if err := CheckPkgsDiff(ctx, logger); err != nil {
		return fmt.Errorf("pkgs directory check failed: %w", err)
	}

	if !cfg.NoCreateBranch {
		logger.Info("Setting up git branch")
		if err := GitCheckout(ctx, logger, pkgName); err != nil {
			return fmt.Errorf("git checkout failed: %w", err)
		}
	}

	logger.Info("Starting Linux container")
	linuxContainer := docker.DefaultLinuxContainer()
	linuxDM := docker.NewManager(linuxContainer)
	if err := linuxDM.EnsureContainer(ctx, logger, cfg.Recreate); err != nil {
		return fmt.Errorf("failed to ensure Linux container: %w", err)
	}

	logger.Info("Running scaffold in container")
	if err := scaffoldInContainer(ctx, logger, linuxDM, cfg, githubToken); err != nil {
		return fmt.Errorf("scaffold in container failed: %w", err)
	}

	logger.Info("Updating registry.yaml")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
	}

	logger.Info("Committing changes")
	if err := GitCommit(ctx, logger, pkgName); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	if err := runTests(ctx, logger, cfg, linuxDM, pkgName, githubToken); err != nil {
		return err
	}
	return nil
}

func runTests(ctx context.Context, logger *slog.Logger, cfg *Config, linuxDM *docker.Manager, pkgName, githubToken string) error {
	logger.Info("Running Linux/Darwin tests")
	if err := RunLinuxDarwinTests(ctx, logger, linuxDM, pkgName, githubToken); err != nil {
		return fmt.Errorf("Linux/Darwin tests failed: %w", err)
	}

	logger.Info("Starting a container for Windows")
	windowsContainer := docker.DefaultWindowsContainer()
	windowsDM := docker.NewManager(windowsContainer)
	if err := windowsDM.EnsureContainer(ctx, logger, cfg.Recreate); err != nil {
		return fmt.Errorf("failed to ensure Windows container: %w", err)
	}

	logger.Info("Running Windows tests")
	if err := RunWindowsTests(ctx, logger, windowsDM, pkgName, githubToken); err != nil {
		return fmt.Errorf("windows tests failed: %w", err)
	}

	return nil
}

// scaffoldInContainer runs aqua gr inside the Docker container.
func scaffoldInContainer(ctx context.Context, logger *slog.Logger, dm *docker.Manager, cfg *Config, githubToken string) error {
	pkgName := cfg.PkgName
	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)

	// Create local package directory
	if err := os.MkdirAll(pkgDir, docker.DirPermission); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	if err := copyScaffoldConfig(ctx, logger, dm, cfg, pkgDir); err != nil {
		return err
	}

	// Remove old pkg.yaml if exists
	if err := dm.ExecBash(ctx, logger, "rm pkg.yaml 2>/dev/null || :"); err != nil {
		return fmt.Errorf("remove existing pkg.yaml in the container if it exists: %w", err)
	}

	if err := runAquaGRInContainer(ctx, logger, dm, cfg, pkgDir, githubToken); err != nil {
		return err
	}

	// Copy results back from container
	if err := dm.CopyFrom(ctx, logger, dm.Config().WorkingDir+"/pkg.yaml", filepath.Join(pkgDir, "pkg.yaml")); err != nil {
		return fmt.Errorf("copy pkg.yaml from container: %w", err)
	}
	if err := dm.CopyTo(ctx, logger, filepath.Join(pkgDir, "registry.yaml"), dm.Config().WorkingDir+"/registry.yaml"); err != nil {
		return fmt.Errorf("copy registry.yaml from container: %w", err)
	}

	// Copy scaffold config to package directory if provided
	if cfg.ConfigPath != "" {
		if err := docker.CopyFile(cfg.ConfigPath, filepath.Join(pkgDir, "scaffold.yaml")); err != nil {
			return fmt.Errorf("copy scaffold config to pkgs: %w", err)
		}
	}

	return nil
}

func copyScaffoldConfig(ctx context.Context, logger *slog.Logger, dm *docker.Manager, cfg *Config, pkgDir string) error {
	if cfg.ConfigPath != "" {
		if err := dm.CopyTo(ctx, logger, cfg.ConfigPath, dm.Config().WorkingDir+"/scaffold.yaml"); err != nil {
			return fmt.Errorf("copy scaffold config to container: %w", err)
		}
		return nil
	}
	// Check if pkgs/{pkg}/scaffold.yaml exists
	scaffoldConfig := filepath.Join(pkgDir, "scaffold.yaml")
	if _, err := os.Stat(scaffoldConfig); err == nil {
		logger.Info("using scaffold config", "path", scaffoldConfig)
		if err := dm.CopyTo(ctx, logger, scaffoldConfig, dm.Config().WorkingDir+"/scaffold.yaml"); err != nil {
			return fmt.Errorf("copy scaffold config to container: %w", err)
		}
	}
	return nil
}

func runAquaGRInContainer(ctx context.Context, logger *slog.Logger, dm *docker.Manager, cfg *Config, pkgDir, githubToken string) error {
	var env map[string]string
	if githubToken != "" {
		env = map[string]string{
			"AQUA_GITHUB_TOKEN": githubToken,
		}
	}
	grCmd := []string{"aqua", "gr", "--out-testdata", "pkg.yaml"}
	if cfg.Cmds != "" {
		grCmd = append(grCmd, "-cmd", cfg.Cmds)
	}
	if cfg.Limit != 0 {
		grCmd = append(grCmd, "-limit", strconv.Itoa(cfg.Limit))
	}
	if cfg.ConfigPath != "" || fileExists(filepath.Join(pkgDir, "scaffold.yaml")) {
		grCmd = append(grCmd, "-c", "scaffold.yaml")
	}
	grCmd = append(grCmd, cfg.PkgName)

	args := make([]string, 0, 3+2*len(env)+1+len(grCmd))
	args = append(args, "exec", "-w", docker.ContainerWorkingDir)
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, dm.Config().Name)
	args = append(args, grCmd...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	logger.Info("+ " + docker.RedactSecrets(cmd.String(), env))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker exec: %w", err)
	}
	return writeRegistryYAML(filepath.Join(pkgDir, "registry.yaml"), buf.Bytes())
}

func writeRegistryYAML(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open registry.yaml: %w", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/registry.json\n"); err != nil {
		return fmt.Errorf("write yaml-language-server comment to registry.yaml: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write registry.yaml: %w", err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("write registry.yaml: %w", err)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func aquaGR(ctx context.Context, logger *slog.Logger, pkgName, pkgFilePath, rgFilePath, cmds string, limit int, configPath string) error {
	outFile, err := os.Create(rgFilePath)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", rgFilePath, err)
	}
	defer outFile.Close()

	if _, err := outFile.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/registry.json\n"); err != nil {
		return fmt.Errorf("write a code comment for yaml-language-server: %w", err)
	}

	args := []string{"gr", "-out-testdata", pkgFilePath}

	if cmds != "" {
		args = append(args, "-cmd", cmds)
	}
	if limit != 0 {
		s := strconv.Itoa(limit)
		args = append(args, "-limit", s)
	}
	if configPath != "" {
		args = append(args, "-c", configPath)
	}

	cmd := exec.CommandContext(ctx, "aqua", append(args, pkgName)...) //nolint:gosec
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	logger.Info("+ " + cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %w", err)
	}

	return nil
}
