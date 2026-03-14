package scaffold

import (
	"context"
	"log/slog"

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

const scaffoldDescription = `Scaffold a package.

e.g.

$ argd scaffold cli/cli
`

func Command(logger *slog.Logger, gFlags *gflag.Flags) *cli.Command {
	flags := &scaffoldFlags{
		Flags: gFlags,
	}
	return &cli.Command{
		Name:        "scaffold",
		Aliases:     []string{"s"},
		Usage:       `Scaffold a package`,
		UsageText:   `$ argd scaffold [options] <package name>`,
		Description: scaffoldDescription,
		Flags:       scaffoldCLIFlags(flags),
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

			return scaffold.Scaffold(ctx, logger, cfg)
		},
	}
}

func scaffoldCLIFlags(flags *scaffoldFlags) []cli.Flag {
	return []cli.Flag{
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
	}
}
