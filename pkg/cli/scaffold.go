package cli

import (
	"github.com/aquaproj/registry-tool/pkg/scaffold"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newScaffoldCommand() *cli.Command {
	return &cli.Command{
		Name:      "scaffold",
		Usage:     `Scaffold a package`,
		UsageText: `$ aqua-registry scaffold [--deep] <package name>`,
		Description: `Scaffold a package.

aqua >= v1.14.0 is required.

e.g.

$ aqua-registry scaffold cli/cli

This tool does the following things.

1. Scaffold configuration files.
1. Install packages for testing

--

1. Create directories pkgs/<package name>
2. Create pkgs/<package name>/pkg.yaml and pkgs/<package name>/registry.yaml
3. Update registry.yaml
4. Create or update aqua.yaml
5. aqua g -i <package name>
6. aqua i
`,
		Action: runner.scaffoldAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "Resolve version_overrides. aqua >= v1.34.0 is required",
			},
		},
	}
}

func (runner *Runner) scaffoldAction(c *cli.Context) error {
	return scaffold.Scaffold(c.Context, c.Bool("deep"), c.Args().Slice()...) //nolint:wrapcheck
}
