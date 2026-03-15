package remove

import (
	"context"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/remove"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Remove Docker containers",
		UsageText: "argd remove",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return remove.Remove(ctx, logger)
		},
	}
}
