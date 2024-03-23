package cli

import (
	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newGenerateRegistryCommand() *cli.Command {
	return &cli.Command{
		Name:      "generate-registry",
		Aliases:   []string{"gr"},
		Usage:     `Update registry.yaml`,
		UsageText: `aqua-registry gr`,
		Description: `Update registry.yaml

This command updates registry.yaml on the repository root directory.
Don't edit it manually, and if you update registry.yaml in the pkgs directory, don't forget to run this command.

No argument is needed.
`,
		Action: runner.generateRegistryAction,
	}
}

func (runner *Runner) generateRegistryAction(_ *cli.Context) error {
	return genrg.GenerateRegistry() //nolint:wrapcheck
}
