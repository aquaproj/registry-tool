package start

import (
	"context"
	"log/slog"

	"github.com/aquaproj/registry-tool/pkg/start"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	var recreate bool
	return &cli.Command{
		Name:      "start",
		Usage:     "Start Docker containers",
		UsageText: "argd start [-r]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "recreate",
				Aliases:     []string{"r"},
				Usage:       "Recreate the containers",
				Destination: &recreate,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return start.Start(ctx, logger, recreate)
		},
	}
}
