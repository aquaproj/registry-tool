package scaffold

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/cli/gflag"
	"github.com/aquaproj/registry-tool/pkg/scaffold"
	"github.com/urfave/cli/v3"
)

type scaffoldFlags struct {
	*gflag.Flags

	Cmd            string
	Config         string
	Limit          int
	Deep           bool
	Local          bool
	Recreate       bool
	NoCreateBranch bool
}

func Command(gFlags *gflag.Flags) *cli.Command {
	flags := &scaffoldFlags{
		Flags: gFlags,
	}
	return &cli.Command{
		Name:      "scaffold",
		Usage:     `Scaffold a package`,
		UsageText: `$ aqua-registry scaffold [options] <package name>`,
		Description: `Scaffold a package.

e.g.

$ aqua-registry scaffold cli/cli

This tool does the following things.

## Full mode (default)

1. Check prerequisites (docker, git, aqua)
2. Check for uncommitted changes in pkgs directory
3. Create/switch to feature branch (feat/{pkg})
4. Start Linux container
5. Scaffold configuration files in container
6. Update registry.yaml
7. Commit changes
8. Run Linux/Darwin tests
9. Start Windows container
10. Run Windows tests

## Local mode (--local)

1. Scaffold configuration files
2. Install packages for testing

--

1. Create directories pkgs/<package name>
2. Create pkgs/<package name>/pkg.yaml and pkgs/<package name>/registry.yaml
3. Update registry.yaml
4. Create or update aqua.yaml
5. aqua g -i <package name>
6. aqua i
`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "deep",
				Usage:       "This flag was deprecated and had no meaning from aqua v2.15.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/2351",
				Destination: &flags.Deep,
			},
			&cli.StringFlag{
				Name:        "cmd",
				Usage:       "A list of commands joined with single quotes ','",
				Destination: &flags.Cmd,
			},
			&cli.IntFlag{
				Name:        "limit",
				Aliases:     []string{"l"},
				Usage:       "the maximum number of versions",
				Destination: &flags.Limit,
			},
			&cli.BoolFlag{
				Name:        "local",
				Usage:       "Run in local mode without Docker (simple scaffold only)",
				Destination: &flags.Local,
			},
			&cli.BoolFlag{
				Name:        "recreate",
				Aliases:     []string{"r"},
				Usage:       "Recreate Docker containers",
				Destination: &flags.Recreate,
			},
			&cli.BoolFlag{
				Name:        "no-create-branch",
				Aliases:     []string{"B"},
				Usage:       "Don't create a git branch",
				Destination: &flags.NoCreateBranch,
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Path to scaffold.yaml configuration file",
				Destination: &flags.Config,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			args := c.Args().Slice()
			pkgName := ""
			if len(args) > 0 {
				pkgName = args[0]
			}

			cfg := &scaffold.Config{
				PkgName:        pkgName,
				Cmds:           flags.Cmd,
				Limit:          flags.Limit,
				Local:          flags.Local,
				Recreate:       flags.Recreate,
				NoCreateBranch: flags.NoCreateBranch,
				ConfigPath:     flags.Config,
			}

			return scaffold.Scaffold(ctx, cfg)
		},
	}
}
