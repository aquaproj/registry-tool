package createprnewpkg

import (
	"context"
	"log/slog"

	newpkg "github.com/aquaproj/registry-tool/pkg/create-pr-new-pkg"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "create-pr-new-pkg",
		Aliases:   []string{"new"},
		Usage:     `Create a pull request to add new packages`,
		UsageText: `argd new <package name> [<package name> ...]`,
		Description: `Create a pull request to add new packages.

e.g.

$ argd new cli/cli

This tool does the following things.

1. Create a feature branch
2. Create a commit
3. Push the commit to origin
4. Open a web browser to create a request with GitHub CLI
`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return newpkg.CreatePRNewPkgs(ctx, logger, cmd.Args().First())
		},
	}
}
