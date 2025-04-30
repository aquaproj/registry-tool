package cli

import (
	"context"
	"errors"

	"github.com/aquaproj/registry-tool/pkg/mv"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

func (r *Runner) newMVCommand() *cli.Command {
	return &cli.Command{
		Name:        "mv",
		Usage:       `Rename a package`,
		UsageText:   `$ aqua-registry mv <old package name> <new package name>`,
		Description: `Rename a package.`,
		Action:      r.moveAction,
	}
}

func (r *Runner) moveAction(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) != 2 { //nolint:mnd
		return errors.New("invalid arguments")
	}
	return mv.Move(ctx, afero.NewOsFs(), args[0], args[1]) //nolint:wrapcheck
}
