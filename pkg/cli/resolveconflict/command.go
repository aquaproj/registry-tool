package resolveconflict

import (
	"context"
	"log/slog"

	rc "github.com/aquaproj/registry-tool/pkg/resolve-conflict"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "resolve-conflict",
		Usage:     "Resolve registry.yaml merge conflict with main",
		UsageText: "argd resolve-conflict <PR number>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return rc.ResolveConflict(ctx, logger, cmd.Args().First())
		},
	}
}
