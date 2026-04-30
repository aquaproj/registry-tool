package docker

import "os"

// Config holds Docker container configuration.
type Config struct {
	Name       string
	Image      string
	WorkingDir string
	// Dockerfile is the file name under docker/ used to build the image.
	// If empty, "Dockerfile" is used.
	Dockerfile string
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

// DefaultAlpineContainer returns the default Alpine (musl) Linux container configuration.
// It is used when registry.yaml contains packages with `key: libc` variants
// to verify both musl and gnu libc paths.
func DefaultAlpineContainer() Config {
	return Config{
		Name:       "aqua-registry-alpine",
		Image:      "aquaproj/aqua-registry-alpine",
		WorkingDir: ContainerWorkingDir,
		Dockerfile: "Dockerfile-alpine",
	}
}
