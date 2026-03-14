package docker

import "os"

// Config holds Docker container configuration.
type Config struct {
	Name       string
	Image      string
	WorkingDir string
}

const (
	// ContainerWorkingDir is the default working directory inside containers.
	ContainerWorkingDir = "/workspace"
	// DirPermission is the default permission for directories.
	DirPermission os.FileMode = 0o775
	// FilePermission is the default permission for files.
	FilePermission os.FileMode = 0o644
)

// DefaultLinuxContainer returns the default Linux container configuration.
func DefaultLinuxContainer() Config {
	return Config{
		Name:       "aqua-registry",
		Image:      "aquaproj/aqua-registry",
		WorkingDir: ContainerWorkingDir,
	}
}

// DefaultWindowsContainer returns the default Windows container configuration.
func DefaultWindowsContainer() Config {
	return Config{
		Name:       "aqua-registry-windows",
		Image:      "aquaproj/aqua-registry",
		WorkingDir: ContainerWorkingDir,
	}
}
