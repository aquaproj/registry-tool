package removepackage

import (
	"context"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/removepackage"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "remove-package",
		Aliases:   []string{"rmp"},
		Usage:     "Remove a package from Docker containers",
		UsageText: "argd remove-package [<package name>]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return removepackage.RemovePackage(ctx, logger, cmd.Args().First())
		},
	}
}
