package fix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/registry-tool/pkg/naming"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func Fix(ctx context.Context, logger *slog.Logger, args []string) error {
	for _, arg := range args {
		if err := fix(ctx, logger, arg); err != nil {
			return slogerr.With(err, "arg", arg) //nolint:wrapcheck
		}
	}
	return nil
}

func fix(ctx context.Context, logger *slog.Logger, arg string) error {
	base := filepath.Base(arg)
	switch base {
	case "registry.yaml":
		dir := filepath.ToSlash(filepath.Dir(arg))
		pkgName, ok := strings.CutPrefix(dir, "pkgs/")
		if !ok {
			return nil
		}
		if err := fixRegistryYAML(pkgName, arg); err != nil {
			return err
		}
	case "pkg.yaml":
		return nil
	case "scaffold.yaml":
		return nil
	default:
		if err := fixPackage(ctx, logger, arg); err != nil {
			return err
		}
	}
	return nil
}

func fixPackage(ctx context.Context, logger *slog.Logger, pkgName string) error {
	pkgName, err := naming.Resolve(ctx, logger, pkgName)
	if err != nil {
		return fmt.Errorf("resolve package name: %w", err)
	}

	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)

	registryFile := filepath.Join(pkgDir, "registry.yaml")
	return fixRegistryYAML(pkgName, registryFile)
}

func fixRegistryYAML(pkgName, registryFile string) error {
	rb, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", registryFile, err)
	}

	file, err := parser.ParseBytes(rb, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse registry YAML: %w", err)
	}
	body := file.Docs[0].Body

	var registryCfg registry.Config
	if err := yaml.Unmarshal(rb, &registryCfg); err != nil {
		return fmt.Errorf("parse %s: %w", registryFile, err)
	}

	if len(registryCfg.PackageInfos) == 0 {
		return slogerr.With(errors.New("packages is empty"), "file", registryFile) //nolint:wrapcheck
	}
	if len(registryCfg.PackageInfos) > 1 {
		return slogerr.With(errors.New("packages must include only one package"), "file", registryFile) //nolint:wrapcheck
	}
	pkgInfo := registryCfg.PackageInfos[0]
	actions, err := getActionsForPkgInfo(pkgName, pkgInfo)
	if err != nil {
		return err
	}
	for _, act := range actions {
		if err := act.Run(body); err != nil {
			return fmt.Errorf("run fix action: %w", err)
		}
	}
	return writeFile(registryFile, []byte(file.String()))
}

func writeFile(path string, data []byte) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.WriteFile(filepath.Clean(path), data, stat.Mode()); err != nil { //nolint:gosec
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func getActionsForPkgInfo(pkgName string, pkgInfo *registry.PackageInfo) ([]yamledit.Action, error) {
	actions := fixFiles(pkgInfo)
	if strings.Contains(pkgName, ".") {
		return actions, nil
	}

	repoFullName := pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	if pkgInfo.RepoOwner == "" {
		return nil, errors.New("repo_owner must be specified if package name doesn't include period")
	}
	if pkgInfo.RepoName == "" {
		return nil, errors.New("if package name doesn't include period, repo_name must be specified")
	}
	if pkgInfo.Name == repoFullName {
		actions = append(actions, yamledit.MapAction(
			"$.packages[0]",
			yamledit.RemoveKeys("name"),
		))
	}

	return actions, nil
}

func fixFiles(pkgInfo *registry.PackageInfo) []yamledit.Action {
	var actions []yamledit.Action
	for i, file := range pkgInfo.Files {
		actions = append(
			actions,
			yamledit.MapAction(
				fmt.Sprintf("$.packages[0].files[%d]", i),
				fixFile(file)...))
	}
	for i, ov := range pkgInfo.Overrides {
		for j, file := range ov.Files {
			actions = append(
				actions,
				yamledit.MapAction(
					fmt.Sprintf("$.packages[0].overrides[%d].files[%d]", i, j),
					fixFile(file)...))
		}
	}
	for i, vov := range pkgInfo.VersionOverrides {
		for j, file := range vov.Files {
			actions = append(
				actions,
				yamledit.MapAction(
					fmt.Sprintf("$.packages[0].version_overrides[%d].files[%d]", i, j),
					fixFile(file)...))
		}
		for j, ov := range vov.Overrides {
			for k, file := range ov.Files {
				actions = append(
					actions,
					yamledit.MapAction(
						fmt.Sprintf("$.packages[0].version_overrides[%d].overrides[%d].files[%d]", i, j, k),
						fixFile(file)...))
			}
		}
	}
	return actions
}

// .files
// .overrides[].files
// .version_overrides[].files
// .version_overrides[].overrides[].files
// .name, .src

func fixFile(file *registry.File) []yamledit.MappingNodeAction {
	var actions []yamledit.MappingNodeAction
	name, ok := strings.CutSuffix(file.Name, ".exe")
	if ok {
		actions = append(actions, yamledit.SetKey("name", name, nil))
	}
	src, ok := strings.CutSuffix(file.Src, ".exe")
	if ok {
		actions = append(actions, yamledit.SetKey("src", src, nil))
	}
	if name == src {
		actions = append(actions, yamledit.RemoveKeys("src"))
	}
	return actions
}
