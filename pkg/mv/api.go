package mv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	dirPermission  os.FileMode = 0o775
	filePermission os.FileMode = 0o644
)

func Move(_ context.Context, afs afero.Fs, oldPackageName, newPackageName string) error {
	// Check if the old package exists
	oldPkgPath := filepath.Join("pkgs", filepath.FromSlash(oldPackageName))
	newPkgPath := filepath.Join("pkgs", filepath.FromSlash(newPackageName))
	// Create directories
	if err := afs.MkdirAll(newPkgPath, dirPermission); err != nil {
		return fmt.Errorf("create directories for new package: %w", err)
	}
	newPkgYAMLPath := filepath.Join(newPkgPath, "pkg.yaml")
	newRegistryYAMLPath := filepath.Join(newPkgPath, "registry.yaml")
	newScaffoldYAMLPath := filepath.Join(newPkgPath, "scaffold.yaml")
	// Move files
	if err := afs.Rename(filepath.Join(oldPkgPath, "pkg.yaml"), newPkgYAMLPath); err != nil {
		return fmt.Errorf("move pkg.yaml: %w", err)
	}
	if err := afs.Rename(filepath.Join(oldPkgPath, "registry.yaml"), newRegistryYAMLPath); err != nil {
		return fmt.Errorf("move registry.yaml: %w", err)
	}
	oldScaffoldYAMLPath := filepath.Join(oldPkgPath, "scaffold.yaml")
	if f, err := afero.Exists(afs, oldScaffoldYAMLPath); err != nil {
		return fmt.Errorf("check if scaffold.yaml exists: %w", err)
	} else if f {
		if err := afs.Rename(oldScaffoldYAMLPath, newScaffoldYAMLPath); err != nil {
			return fmt.Errorf("move scaffold.yaml: %w", err)
		}
	}

	// Fix repo_owner and repo_name in registry.yaml
	// Add aliases in registry.yaml
	if err := editRegistry(afs, newRegistryYAMLPath, oldPackageName, newPackageName); err != nil {
		return err
	}
	// Fix package names in pkg.yaml
	if err := editPackageYAML(afs, newPkgYAMLPath, oldPackageName, newPackageName); err != nil {
		return err
	}
	return nil
}
