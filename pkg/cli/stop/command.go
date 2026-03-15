package stop

import (
	"context"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/stop"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "stop",
		Usage:     "Stop Docker containers",
		UsageText: "argd stop",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return stop.Stop(ctx, logger)
		},
	}
}
