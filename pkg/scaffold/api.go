package scaffold

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/aquaproj/registry-tool/pkg/initcmd"
)

const (
	dirPermission  os.FileMode = 0o775
	filePermission os.FileMode = 0o644
)

func Scaffold(ctx context.Context, pkgNames ...string) error {
	if len(pkgNames) != 1 {
		return errors.New(`usage: $ aqua-registry scaffold <pkgname>
e.g. $ aqua-registry scaffold cli/cli`)
	}
	pkgName := pkgNames[0]
	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)
	pkgFile := filepath.Join(pkgDir, "pkg.yaml")
	rgFile := filepath.Join(pkgDir, "registry.yaml")
	if err := os.MkdirAll(pkgDir, dirPermission); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}
	if err := aquaGR(ctx, pkgName, rgFile); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Update registry.yaml")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
	}
	if err := initcmd.Init(ctx); err != nil {
		return err //nolint:wrapcheck
	}
	if err := aquaG(ctx, "local,"+pkgName); err != nil {
		return err
	}
	if err := createPkgFile(ctx, pkgName, pkgFile); err != nil {
		return err
	}
	if err := aquaI(ctx); err != nil {
		return err
	}
	return nil
}

func aquaGR(ctx context.Context, pkgName, rgFilePath string) error {
	outFile, err := os.Create(rgFilePath)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", rgFilePath, err)
	}
	defer outFile.Close()
	fmt.Fprintf(os.Stderr, "+ aqua gr %s >> registry.yaml\n", pkgName)
	cmd := exec.CommandContext(ctx, "aqua", "gr", pkgName)
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %w", err)
	}
	return nil
}

func aquaG(ctx context.Context, pkgName string) error {
	outFile, err := os.OpenFile("aqua-local.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermission)
	if err != nil {
		return fmt.Errorf("open aqua-local.yaml: %w", err)
	}
	defer outFile.Close()
	fmt.Fprintf(os.Stderr, "+ aqua g %s >> aqua-local.yaml\n", pkgName)
	cmd := exec.CommandContext(ctx, "aqua", "g", pkgName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %w", err)
	}
	return nil
}

func aquaI(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "+ aqua i --test")
	cmd := exec.CommandContext(ctx, "aqua", "i", "--test")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: aqua i: %w", err)
	}
	return nil
}

func createPkgFile(ctx context.Context, pkgName, pkgFilePath string) error {
	outFile, err := os.Create(pkgFilePath)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", pkgFilePath, err)
	}
	defer outFile.Close()
	if _, err := outFile.WriteString("packages:\n"); err != nil {
		return fmt.Errorf("write a string to file %s: %w", pkgFilePath, err)
	}
	buf := &bytes.Buffer{}
	fmt.Fprintf(os.Stderr, "+ aqua -c aqua-all.yaml g %s >> %s\n", pkgName, pkgFilePath)
	cmd := exec.CommandContext(ctx, "aqua", "-c", "aqua-all.yaml", "g", pkgName)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: aqua g %s: %w", pkgName, err)
	}
	txt := ""
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		txt += fmt.Sprintf("  %s\n", line)
	}
	if _, err := outFile.WriteString(txt); err != nil {
		return fmt.Errorf("write a string to file %s: %w", pkgFilePath, err)
	}
	return nil
}
