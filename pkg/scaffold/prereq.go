package scaffold

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/osexec"
)

// CheckPrerequisites checks if required commands are available.
func CheckPrerequisites(ctx context.Context, logger *slog.Logger) error {
	commands := []string{"docker", "git", "aqua"}
	var missing []string
	for _, cmd := range commands {
		if err := checkCommand(ctx, logger, cmd); err != nil {
			missing = append(missing, cmd)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required commands not found: %s", strings.Join(missing, ", "))
	}
	return nil
}

func checkCommand(ctx context.Context, logger *slog.Logger, name string) error {
	cmd := exec.CommandContext(ctx, name, "--version")
	osexec.SetCancel(logger, cmd)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() //nolint:wrapcheck
}

// CheckPkgsDiff checks if the pkgs directory has uncommitted changes.
func CheckPkgsDiff(ctx context.Context, logger *slog.Logger) error {
	// Check for unstaged changes
	if err := gitDiffQuiet(ctx, logger, "pkgs"); err != nil {
		return errors.New("the directory pkgs has unstaged changes")
	}

	// Check for staged changes
	if err := gitDiffCachedQuiet(ctx, logger, "pkgs"); err != nil {
		return errors.New("the directory pkgs has staged changes")
	}

	// Check for untracked files
	untracked, err := gitLsFilesOthers(ctx, logger, "pkgs")
	if err != nil {
		return fmt.Errorf("check untracked files: %w", err)
	}
	if len(untracked) > 0 {
		return fmt.Errorf("the directory pkgs has untracked files:\n%s", strings.Join(untracked, "\n"))
	}

	return nil
}

func gitDiffQuiet(ctx context.Context, logger *slog.Logger, path string) error {
	cmd := exec.CommandContext(ctx, "git", "diff", "--quiet", path)
	cmd.Stdout = nil
	cmd.Stderr = nil
	osexec.SetCancel(logger, cmd)
	return cmd.Run() //nolint:wrapcheck
}

func gitDiffCachedQuiet(ctx context.Context, logger *slog.Logger, path string) error {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet", path)
	cmd.Stdout = nil
	cmd.Stderr = nil
	osexec.SetCancel(logger, cmd)
	return cmd.Run() //nolint:wrapcheck
}

func gitLsFilesOthers(ctx context.Context, logger *slog.Logger, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-files", "--others", "--exclude-standard", path)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, nil
	}
	return strings.Split(output, "\n"), nil
}
