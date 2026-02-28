package scaffold

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GitCheckout creates or switches to a feature branch for the package.
// If noCreateBranch is true, this function does nothing.
func GitCheckout(ctx context.Context, pkgName string, noCreateBranch bool) error {
	if noCreateBranch {
		return nil
	}

	branch := "feat/" + pkgName

	// Check if branch already exists locally
	if branchExists(ctx, branch) {
		return gitCheckoutBranch(ctx, branch)
	}

	// Create a new branch from upstream main
	return createBranchFromUpstream(ctx, branch)
}

func branchExists(ctx context.Context, branch string) bool {
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--quiet", "refs/heads/"+branch) //nolint:gosec
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func gitCheckoutBranch(ctx context.Context, branch string) error {
	fmt.Fprintf(os.Stderr, "+ git checkout %s\n", branch)
	cmd := exec.CommandContext(ctx, "git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout %s: %w", branch, err)
	}
	return nil
}

func createBranchFromUpstream(ctx context.Context, branch string) error {
	// Create a temporary remote to fetch from upstream
	tempRemote := "temp-remote-" + time.Now().Format("20060102150405")

	// Add temporary remote
	fmt.Fprintf(os.Stderr, "+ git remote add %s https://github.com/aquaproj/aqua-registry\n", tempRemote)
	cmd := exec.CommandContext(ctx, "git", "remote", "add", tempRemote, "https://github.com/aquaproj/aqua-registry")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}

	// Ensure we remove the temporary remote even if something fails
	defer func() {
		fmt.Fprintf(os.Stderr, "+ git remote remove %s\n", tempRemote)
		rmCmd := exec.CommandContext(ctx, "git", "remote", "remove", tempRemote)
		rmCmd.Stdout = os.Stdout
		rmCmd.Stderr = os.Stderr
		_ = rmCmd.Run()
	}()

	// Fetch main from upstream
	fmt.Fprintf(os.Stderr, "+ git fetch %s main\n", tempRemote)
	cmd = exec.CommandContext(ctx, "git", "fetch", tempRemote, "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch: %w", err)
	}

	// Create and checkout new branch
	fmt.Fprintf(os.Stderr, "+ git checkout -b %s %s/main\n", branch, tempRemote)
	cmd = exec.CommandContext(ctx, "git", "checkout", "-b", branch, tempRemote+"/main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout -b: %w", err)
	}

	return nil
}

// GitCommit commits the scaffold changes.
func GitCommit(ctx context.Context, pkgName string) error {
	// Stage the registry.yaml and package files
	pkgPattern := "pkgs/" + pkgName + "/*.yaml"

	fmt.Fprintf(os.Stderr, "+ git add registry.yaml %s\n", pkgPattern)
	cmd := exec.CommandContext(ctx, "git", "add", "registry.yaml", pkgPattern)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Create commit with conventional commit message
	commitMsg := fmt.Sprintf("feat(%s): scaffold %s", pkgName, pkgName)

	fmt.Fprintf(os.Stderr, "+ git commit -m \"%s\"\n", commitMsg)
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
