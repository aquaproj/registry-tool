package fix

import (
	"context"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/fix"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "fix",
		Usage:     "Fix a package",
		UsageText: "argd fix [<package name> or pkgs/**/pkg.yaml or pkgs/**/registry.yaml] ...",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return fix.Fix(ctx, logger, cmd.Args().Slice())
		},
	}
}
