package lint

import (
	"context"
	"log/slog"

	lintpkg "github.com/aquaproj/registry-tool/pkg/lint"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "lint",
		Aliases:   []string{"l"},
		Usage:     "Lint a package",
		UsageText: "aqua-registry lint [<package name>]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return lintpkg.Lint(ctx, logger, cmd.Args().First())
		},
	}
}
