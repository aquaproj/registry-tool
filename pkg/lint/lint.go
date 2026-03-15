package lint

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/registry-tool/pkg/naming"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"gopkg.in/yaml.v3"
)

func Lint(ctx context.Context, logger *slog.Logger, args []string) error {
	for _, arg := range args {
		if err := lint(ctx, logger, arg); err != nil {
			return slogerr.With(err, "arg", arg) //nolint:wrapcheck
		}
	}
	return nil
}

func lint(ctx context.Context, logger *slog.Logger, arg string) error {
	base := filepath.Base(arg)
	switch base {
	case "registry.yaml":
		dir := filepath.ToSlash(filepath.Dir(arg))
		pkgName, ok := strings.CutPrefix(dir, "pkgs/")
		if !ok {
			return nil
		}
		if err := lintRegistryYAML(pkgName, arg); err != nil {
			return err
		}
	case "pkg.yaml":
		dir := filepath.ToSlash(filepath.Dir(arg))
		pkgName, ok := strings.CutPrefix(dir, "pkgs/")
		if !ok {
			return nil
		}
		if err := lintPkgYAML(pkgName, arg); err != nil {
			return err
		}
	case "scaffold.yaml":
		return nil
	default:
		if err := lintPackage(ctx, logger, arg); err != nil {
			return err
		}
	}
	return nil
}

func lintPackage(ctx context.Context, logger *slog.Logger, pkgName string) error {
	pkgName, err := naming.Resolve(ctx, logger, pkgName)
	if err != nil {
		return fmt.Errorf("resolve package name: %w", err)
	}

	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)
	pkgFile := filepath.Join(pkgDir, "pkg.yaml")

	if err := lintPkgYAML(pkgName, pkgFile); err != nil {
		return err
	}

	registryFile := filepath.Join(pkgDir, "registry.yaml")
	return lintRegistryYAML(pkgName, registryFile)
}

func lintPkgYAML(pkgName, pkgFile string) error {
	b, err := os.ReadFile(pkgFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", pkgFile, err)
	}

	var cfg aqua.Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("parse %s: %w", pkgFile, err)
	}

	if len(cfg.Packages) == 0 {
		return slogerr.With(errors.New("packages is empty"), "file", pkgFile) //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		if pkg.Name != pkgName {
			return slogerr.With( //nolint:wrapcheck
				errors.New("package name mismatch"),
				"file", pkgFile,
				"package_name", pkgName,
				"package_name_in_pkg_yaml", pkg.Name,
			)
		}
	}
	return nil
}

func lintRegistryYAML(pkgName, registryFile string) error {
	rb, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", registryFile, err)
	}

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
	if pkgName != pkgInfo.GetName() {
		return slogerr.With( //nolint:wrapcheck
			errors.New("the package name in registry.yaml does not match the given package name"),
			"file", registryFile,
			"package_name", pkgName,
			"package_name_in_registry", pkgInfo.GetName())
	}
	return validatePkgInfo(pkgName, registryFile, pkgInfo)
}

func validatePkgInfo(pkgName, registryFile string, pkgInfo *registry.PackageInfo) error {
	if pkgInfo.VersionConstraints != "" && pkgInfo.VersionConstraints != "false" {
		return slogerr.With( //nolint:wrapcheck
			errors.New(`the top level version_constraints must be either empty or "false"`),
			"file", registryFile,
			"version_constraints", pkgInfo.VersionConstraints,
		)
	}
	if err := validateFiles(pkgInfo); err != nil {
		return slogerr.With(err, "file", registryFile) //nolint:wrapcheck
	}

	repoFullName := pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	if !strings.Contains(pkgName, ".") {
		if pkgInfo.RepoOwner == "" {
			return errors.New("repo_owner must be specified if package name doesn't include period")
		}
		if pkgInfo.RepoName == "" {
			return errors.New("if package name doesn't include period, repo_name must be specified")
		}
		if pkgName != repoFullName && !strings.HasPrefix(pkgName, repoFullName+"/") {
			return slogerr.With( //nolint:wrapcheck
				errors.New("package name must start with repository full name if package name doesn't include period"),
				"package_name", pkgName,
				"repo_owner", pkgInfo.RepoOwner,
				"repo_name", pkgInfo.RepoName,
				"file", registryFile,
			)
		}
	}
	if pkgInfo.Name == repoFullName {
		return slogerr.With( //nolint:wrapcheck
			errors.New("omit .name if it's same with repo_owner/repo_name"),
			"repo_owner", pkgInfo.RepoOwner,
			"repo_name", pkgInfo.RepoName,
			"file", registryFile,
		)
	}

	return nil
}

func validateFiles(pkgInfo *registry.PackageInfo) error {
	files := pkgInfo.Files
	for _, ov := range pkgInfo.Overrides {
		files = append(files, ov.Files...)
	}
	for _, vov := range pkgInfo.VersionOverrides {
		files = append(files, vov.Files...)
		for _, ov := range vov.Overrides {
			files = append(files, ov.Files...)
		}
	}
	for _, file := range files {
		if err := validateFile(file); err != nil {
			return err
		}
	}
	return nil
}

func validateFile(file *registry.File) error {
	if strings.HasSuffix(file.Name, ".exe") {
		return errors.New(".files[].name must not end with .exe. Remove .exe from name")
	}
	if file.Name == file.Src {
		return errors.New("omit .files[].src if it's same with .files[].name")
	}
	if strings.HasSuffix(file.Src, ".exe") {
		return errors.New(".files[].src must not end with .exe. Remove .exe from src")
	}
	return nil
}
