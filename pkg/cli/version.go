package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (runner *Runner) newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Show version",
		Action: runner.versionAction,
	}
}

func (runner *Runner) versionAction(_ context.Context, c *cli.Command) error {
	cli.ShowVersion(c)
	return nil
}
