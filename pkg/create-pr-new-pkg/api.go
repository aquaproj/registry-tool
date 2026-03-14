package newpkg

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/initcmd"
	"gopkg.in/yaml.v3"
)

//go:embed pr_template.md
var bodyTemplate []byte

func CreatePRNewPkgs(ctx context.Context, logger *slog.Logger, pkgName string) error {
	pkgName = strings.TrimPrefix(pkgName, "https://github.com/")
	if err := checkDiffPackage(ctx); err != nil {
		return err
	}
	var err error
	pkgName, err = getPkgFromBranch(ctx, pkgName)
	if err != nil {
		return err
	}
	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)
	rgFilePath := filepath.Join(pkgDir, "registry.yaml")
	rgFile, err := os.Open(rgFilePath)
	if err != nil {
		return fmt.Errorf("open a file %s: %w", rgFilePath, err)
	}
	desc, err := func() (string, error) {
		defer rgFile.Close()
		return getDesc(rgFile)
	}()
	if err != nil {
		return err
	}
	if err := command(ctx, "git", "add", "pkgs/"+pkgName, "registry.yaml"); err != nil {
		return err
	}
	prBody := strings.Join([]string{
		getBody(pkgName, desc),
		"",
		"```sh",
		"aqua g -i " + pkgName,
		"```",
		"",
		string(bodyTemplate),
	}, "\n")
	branch := "feat/" + pkgName
	if err := initcmd.Init(ctx); err != nil {
		return err //nolint:wrapcheck
	}
	stderr := &bytes.Buffer{}
	if err := commandStderr(ctx, io.MultiWriter(os.Stderr, stderr), "git", "push", "origin", branch); err != nil {
		if strings.Contains(stderr.String(), "returned error: 403") {
			logger.With(
				"doc", "https://aquaproj.github.io/docs/products/aqua-registry/contributing#cmdx-new-fails-to-push-a-commit-to-the-origin",
			).Warn(`you don't have the permission to push commits to the origin.
Please fork aquaproj/aqua-registry and fix the origin url to your fork repository.
For details, please see the document`)
		} else {
			return err
		}
	}
	if err := command(ctx, "aqua", "-c", "aqua/dev.yaml", "exec", "--", "gh", "pr", "create", "-w", "-t", "feat: add "+pkgName, "-b", prBody); err != nil {
		return err
	}
	return nil
}

type Packages struct {
	Packages []struct {
		Description string
	}
}

func getDesc(rgFile io.Reader) (string, error) {
	pkgs := &Packages{}
	if err := yaml.NewDecoder(rgFile).Decode(pkgs); err != nil {
		return "", fmt.Errorf("decore registry.yaml as YAML: %w", err)
	}
	if len(pkgs.Packages) != 1 {
		return "", errors.New("the number of packages in registry.yaml must be 1")
	}
	return pkgs.Packages[0].Description, nil
}

func getBody(pkgName, desc string) string {
	pkg := strings.Split(pkgName, "/")
	if len(pkg) >= 2 { //nolint:mnd
		repo := pkg[0] + "/" + pkg[1]
		return fmt.Sprintf(`[%s](https://github.com/%s): %s`, pkgName, repo, desc)
	}
	return fmt.Sprintf(`%s: %s`, pkgName, desc)
}

func checkDiffPackage(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "diff", "--quiet", "pkgs", "registry.yaml")
	if err := cmd.Run(); err != nil {
		return errors.New("there are unstaged changes in pkgs or registry.yaml")
	}
	cmd = exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet", "pkgs", "registry.yaml")
	if err := cmd.Run(); err != nil {
		return errors.New("there are staged changes in pkgs or registry.yaml")
	}
	out, err := exec.CommandContext(ctx, "git", "ls-files", "--others", "--exclude-standard", "pkgs").Output()
	if err != nil {
		return fmt.Errorf("check untracked files in pkgs: %w", err)
	}
	if len(bytes.TrimSpace(out)) > 0 {
		return errors.New("there are untracked files in pkgs")
	}
	return nil
}

func getPkgFromBranch(ctx context.Context, pkgName string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if pkgName != "" {
		if branch != "feat/"+pkgName {
			return "", fmt.Errorf("branch %q doesn't match the package name %q (expected branch feat/%s)", branch, pkgName, pkgName)
		}
		return pkgName, nil
	}
	if !strings.HasPrefix(branch, "feat/") {
		return "", fmt.Errorf("branch %q doesn't have the prefix \"feat/\"", branch)
	}
	return strings.TrimPrefix(branch, "feat/"), nil
}

func command(ctx context.Context, cmdName string, args ...string) error {
	s := cmdName + " " + strings.Join(args, " ")
	fmt.Fprintln(os.Stderr, "+ "+s)
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %s: %w", s, err)
	}
	return nil
}

func commandStderr(ctx context.Context, stderr io.Writer, cmdName string, args ...string) error {
	s := cmdName + " " + strings.Join(args, " ")
	fmt.Fprintln(os.Stderr, "+ "+s)
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %s: %w", s, err)
	}
	return nil
}
