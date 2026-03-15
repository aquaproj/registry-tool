package scaffold

// Config holds the configuration for the scaffold command.
type Config struct {
	// PkgName is the package name (e.g., "cli/cli")
	PkgName string
	// Cmds is a comma-separated list of commands to test
	Cmds string
	// Limit is the maximum number of versions to generate
	Limit int
	// Local runs in local mode without Docker
	Local bool
	// Recreate forces recreation of Docker containers
	Recreate bool
	// NoCreateBranch skips creating a git branch
	NoCreateBranch bool
	// ConfigPath is the path to scaffold.yaml config file
	ConfigPath string
}

// Platform represents a target platform for testing.
type Platform struct {
	OS   string
	Arch string
}

// LinuxDarwinPlatforms returns the Linux and Darwin platforms for testing.
func LinuxDarwinPlatforms() []Platform {
	return []Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
	}
}

// WindowsPlatforms returns the Windows platforms for testing.
func WindowsPlatforms() []Platform {
	return []Platform{
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "arm64"},
	}
}
