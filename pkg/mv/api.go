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
	// Move files
	if err := afs.Rename(filepath.Join(oldPkgPath, "pkg.yaml"), newPkgYAMLPath); err != nil {
		return fmt.Errorf("rename directories: %w", err)
	}
	if err := afs.Rename(filepath.Join(oldPkgPath, "registry.yaml"), newRegistryYAMLPath); err != nil {
		return fmt.Errorf("rename directories: %w", err)
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
