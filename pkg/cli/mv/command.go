package mv

import (
	"context"
	"errors"

	"github.com/aquaproj/registry-tool/pkg/mv"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

type runner struct{}

func Command() *cli.Command {
	return (&runner{}).Command()
}

func (r *runner) Command() *cli.Command {
	return &cli.Command{
		Name:        "mv",
		Usage:       `Rename a package`,
		UsageText:   `$ aqua-registry mv <old package name> <new package name>`,
		Description: `Rename a package.`,
		Action:      r.action,
	}
}

func (r *runner) action(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) != 2 { //nolint:mnd
		return errors.New("invalid arguments")
	}
	return mv.Move(ctx, afero.NewOsFs(), args[0], args[1]) //nolint:wrapcheck
}
