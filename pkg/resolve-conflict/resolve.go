package resolveconflict

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/aquaproj/registry-tool/pkg/osexec"
)

func ResolveConflict(ctx context.Context, logger *slog.Logger, prNumber string) error {
	if prNumber == "" {
		return errors.New("PR number is required")
	}

	// Fetch origin main
	if err := run(ctx, logger, "git", "fetch", "origin", "main"); err != nil {
		return err
	}

	// Checkout the PR
	if err := run(ctx, logger, "aqua", "-c", "aqua/dev.yaml", "exec", "--", "gh", "pr", "checkout", prNumber); err != nil {
		return err
	}

	// Merge main with registry.yaml backup/restore
	if err := mergeMainWithBackup(ctx, logger); err != nil {
		return err
	}

	// Stage registry.yaml
	if err := run(ctx, logger, "git", "add", "registry.yaml"); err != nil {
		return err
	}

	// Interactive commit
	commitCmd := exec.CommandContext(ctx, "git", "commit")
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	commitCmd.Stdin = os.Stdin
	osexec.SetCancel(logger, commitCmd)
	logger.Info("+ " + commitCmd.String())
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

func mergeMainWithBackup(ctx context.Context, logger *slog.Logger) error {
	// Copy registry.yaml to a temp file
	tmpFile, err := os.CreateTemp("", "registry-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := copyFile("registry.yaml", tmpFile.Name()); err != nil {
		return fmt.Errorf("backup registry.yaml: %w", err)
	}

	// Merge origin/main — conflict is expected, so ignore the error
	cmd := exec.CommandContext(ctx, "git", "merge", "origin/main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	logger.Info("+ " + cmd.String())
	_ = cmd.Run()

	// Restore registry.yaml from temp file
	if err := copyFile(tmpFile.Name(), "registry.yaml"); err != nil {
		return fmt.Errorf("restore registry.yaml: %w", err)
	}

	// Regenerate registry.yaml
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("generate registry: %w", err)
	}

	return nil
}

func run(ctx context.Context, logger *slog.Logger, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	logger.Info("+ " + cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmd.String(), err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	return nil
}
