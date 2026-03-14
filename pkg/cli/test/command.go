package test

import (
	"context"
	"log/slog"

	testpkg "github.com/aquaproj/registry-tool/pkg/test"
	"github.com/urfave/cli/v3"
)

func Command(logger *slog.Logger) *cli.Command {
	var recreate bool
	return &cli.Command{
		Name:      "test",
		Aliases:   []string{"t"},
		Usage:     "Test a package in Docker containers",
		UsageText: "aqua-registry test [-r] [<package name>]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "recreate",
				Aliases:     []string{"r"},
				Usage:       "Recreate the containers",
				Destination: &recreate,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return testpkg.Test(ctx, logger, &testpkg.Config{
				PkgName:  cmd.Args().First(),
				Recreate: recreate,
			})
		},
	}
}
