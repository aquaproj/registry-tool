package gengr

import (
	"context"

	genrg "github.com/aquaproj/registry-tool/pkg/generate-registry"
	"github.com/urfave/cli/v3"
)

type runner struct{}

func Command() *cli.Command {
	return (&runner{}).Command()
}

func (r *runner) Command() *cli.Command {
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
		Action: r.action,
	}
}

func (r *runner) action(context.Context, *cli.Command) error {
	return genrg.GenerateRegistry() //nolint:wrapcheck
}
