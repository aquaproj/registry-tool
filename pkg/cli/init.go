package cli

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/initcmd"
	"github.com/urfave/cli/v3"
)

func (runner *Runner) newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     `Create configuration files`,
		UsageText: `$ aqua-registry init`,
		Description: `Create configuration files.

* aqua.yaml
* aqua-dev.yaml
`,
		Action: runner.initAction,
	}
}

func (runner *Runner) initAction(ctx context.Context, _ *cli.Command) error {
	return initcmd.Init(ctx) //nolint:wrapcheck
}
