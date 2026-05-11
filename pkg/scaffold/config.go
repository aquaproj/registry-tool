package scaffold

// Config holds the configuration for the scaffold command.
type Config struct {
	// PkgName is the package name (e.g., "cli/cli")
	PkgName string
	// Cmds is a comma-separated list of commands to test
	Cmds string
	// Limit is the maximum number of versions to generate
	Limit int
	// Recreate forces recreation of Docker containers
	Recreate bool
	// NoCreateBranch skips creating a git branch
	NoCreateBranch bool
	// ConfigPath is the path to scaffold.yaml config file
	ConfigPath string
}

const (
	osLinux   = "linux"
	osDarwin  = "darwin"
	archAmd64 = "amd64"
	archArm64 = "arm64"
)

// Platform represents a target platform for testing.
type Platform struct {
	OS   string
	Arch string
}

// LinuxDarwinPlatforms returns the Linux and Darwin platforms for testing.
func LinuxDarwinPlatforms() []Platform {
	return []Platform{
		{OS: osLinux, Arch: archAmd64},
		{OS: osLinux, Arch: archArm64},
		{OS: osDarwin, Arch: archAmd64},
		{OS: osDarwin, Arch: archArm64},
	}
}

// LinuxPlatforms returns the Linux platforms for testing.
// Used for the Alpine (musl) container path where libc differentiation matters.
func LinuxPlatforms() []Platform {
	return []Platform{
		{OS: osLinux, Arch: archAmd64},
		{OS: osLinux, Arch: archArm64},
	}
}

// WindowsPlatforms returns the Windows platforms for testing.
func WindowsPlatforms() []Platform {
	return []Platform{
		{OS: "windows", Arch: archAmd64},
		{OS: "windows", Arch: archArm64},
	}
}
