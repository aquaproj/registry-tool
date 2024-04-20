package newpkg

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/initcmd"
	"gopkg.in/yaml.v3"
)

//go:embed pr_template.md
var bodyTemplate []byte

func CreatePRNewPkgs(ctx context.Context, pkgNames ...string) error { //nolint:cyclop,funlen
	if len(pkgNames) == 0 {
		return errors.New(`usage: $ aqua-registry create-pr-new-pkg <pkgname>...
e.g. $ aqua-registry create-pr-new-pkg cli/cli`)
	}
	bodies := make([]string, len(pkgNames))
	for i, pkgName := range pkgNames {
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
		bodies[i] = getBody(pkgName, desc)
		if err := command(ctx, "git", "add", "pkgs/"+pkgName, "registry.yaml"); err != nil {
			return err
		}
	}
	commitMsg := strings.Join(bodies, "\n")
	body := strings.Join(append(bodies, []string{ //nolint:makezero
		"",
		"```console",
		"$ aqua g -i " + strings.Join(pkgNames, " "),
		"```",
		string(bodyTemplate),
	}...), "\n")
	pkgName := pkgNames[0]
	branch := "feat/" + pkgName
	if err := initcmd.Init(ctx); err != nil {
		return err //nolint:wrapcheck
	}
	if err := command(ctx, "git", "checkout", "-b", branch); err != nil {
		return err
	}
	if err := command(ctx, "git", "commit", "-m", strings.Join([]string{
		"feat: add " + pkgName,
		"",
		commitMsg,
	}, "\n")); err != nil {
		return err
	}
	if err := command(ctx, "git", "push", "origin", branch); err != nil {
		return err
	}
	if err := command(ctx, "aqua", "-c", "aqua/dev.yaml", "exec", "--", "gh", "pr", "create", "-w", "-t", "feat: add "+pkgName, "-b", body); err != nil {
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
	if len(pkg) >= 2 { //nolint:gomnd
		repo := pkg[0] + "/" + pkg[1]
		return fmt.Sprintf(`[%s](https://github.com/%s): %s`, pkgName, repo, desc)
	}
	return fmt.Sprintf(`%s: %s`, pkgName, desc)
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
