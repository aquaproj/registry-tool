package scaffold

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquaproj/registry-tool/pkg/osexec"
)

// GitCheckout creates or switches to a feature branch for the package.
func GitCheckout(ctx context.Context, logger *slog.Logger, pkgName string) error {
	branch := "feat/" + pkgName

	// Check if branch already exists locally
	if branchExists(ctx, logger, branch) {
		return gitCheckoutBranch(ctx, logger, branch)
	}

	// Create a new branch from upstream main
	return createBranchFromUpstream(ctx, logger, branch)
}

func branchExists(ctx context.Context, logger *slog.Logger, branch string) bool {
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--quiet", "refs/heads/"+branch) //nolint:gosec
	cmd.Stdout = nil
	cmd.Stderr = nil
	osexec.SetCancel(logger, cmd)
	return cmd.Run() == nil
}

func gitCheckoutBranch(ctx context.Context, logger *slog.Logger, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	logger.Info("+ " + cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout %s: %w", branch, err)
	}
	return nil
}

func createBranchFromUpstream(ctx context.Context, logger *slog.Logger, branch string) error {
	// Create a temporary remote to fetch from upstream
	tempRemote := "temp-remote-" + time.Now().Format("20060102150405")

	// Add temporary remote
	cmd := exec.CommandContext(ctx, "git", "remote", "add", tempRemote, "https://github.com/aquaproj/aqua-registry")
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}

	// Ensure we remove the temporary remote even if something fails
	defer func() {
		rmCmd := exec.CommandContext(ctx, "git", "remote", "remove", tempRemote)
		logger.Info("+ " + rmCmd.String())
		rmCmd.Stdout = os.Stdout
		rmCmd.Stderr = os.Stderr
		osexec.SetCancel(logger, rmCmd)
		_ = rmCmd.Run()
	}()

	// Fetch main from upstream
	cmd = exec.CommandContext(ctx, "git", "fetch", tempRemote, "main")
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch: %w", err)
	}

	// Create and checkout new branch
	cmd = exec.CommandContext(ctx, "git", "checkout", "-b", branch, tempRemote+"/main")
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout -b: %w", err)
	}

	return nil
}

// GitCommit commits the scaffold changes.
func GitCommit(ctx context.Context, logger *slog.Logger, pkgName string) error {
	// Stage the registry.yaml and package files
	pkgDir := filepath.Join("pkgs", filepath.FromSlash(pkgName))

	cmd := exec.CommandContext(ctx, "git", "add", "registry.yaml", pkgDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Create commit with conventional commit message
	commitMsg := fmt.Sprintf("feat(%s): scaffold %s", pkgName, pkgName)

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch(ctx context.Context, logger *slog.Logger) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
