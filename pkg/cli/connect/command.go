package connect

import (
	"context"
	"log/slog"

	connectpkg "github.com/aquaproj/registry-tool/pkg/connect"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "connect",
		Aliases:   []string{"con"},
		Usage:     "Connect to a Docker container with an interactive shell",
		UsageText: "aqua-registry connect [<os>] [<arch>]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return connectpkg.Connect(ctx, logger, cmd.Args().Get(0), cmd.Args().Get(1))
		},
	}
}
