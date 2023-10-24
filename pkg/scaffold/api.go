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

func Scaffold(ctx context.Context, cmds string, limit int, pkgNames ...string) error {
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
	if err := aquaGR(ctx, pkgName, pkgFile, rgFile, cmds, limit); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Update registry.yaml")
	if err := genrg.GenerateRegistry(); err != nil {
		return fmt.Errorf("update registry.yaml: %w", err)
	}
	if err := initcmd.Init(ctx); err != nil {
		return err //nolint:wrapcheck
	}
	if err := aquaG(pkgFile); err != nil {
		return err
	}
	return nil
}

func aquaGR(ctx context.Context, pkgName, pkgFilePath, rgFilePath string, cmds string, limit int) error {
	outFile, err := os.Create(rgFilePath)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", rgFilePath, err)
	}
	defer outFile.Close()
	var cmd *exec.Cmd
	command := fmt.Sprintf("+ aqua gr --out-testdata %s", pkgFilePath)
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
	fmt.Fprintf(os.Stderr, "%s %s > %s\n", command, pkgName, rgFilePath)
	cmd = exec.CommandContext(ctx, "aqua", append(args, pkgName)...) //nolint:gosec
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %w", err)
	}
	return nil
}

func aquaG(pkgFilePath string) error {
	fmt.Fprintf(os.Stderr, "appending '- import: %s' to aqua-local.yaml\n", pkgFilePath)
	if err := generateInsert("aqua-local.yaml", pkgFilePath); err != nil {
		return err
	}
	return nil
}
