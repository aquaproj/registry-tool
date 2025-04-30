package cli

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/scaffold"
	"github.com/urfave/cli/v3"
)

func (r *Runner) newScaffoldCommand() *cli.Command {
	return &cli.Command{
		Name:      "scaffold",
		Usage:     `Scaffold a package`,
		UsageText: `$ aqua-registry scaffold [-limit <the number of versions] [-cmd <command>[,<command>...]] <package name>`,
		Description: `Scaffold a package.

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
		Action: r.scaffoldAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "This flag was deprecated and had no meaning from aqua v2.15.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/2351",
			},
			&cli.StringFlag{
				Name:  "cmd",
				Usage: "A list of commands joined with single quotes ','",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "the maximum number of versions",
			},
		},
	}
}

func (r *Runner) scaffoldAction(ctx context.Context, c *cli.Command) error {
	return scaffold.Scaffold(ctx, c.String("cmd"), c.Int("limit"), c.Args().Slice()...) //nolint:wrapcheck
}
