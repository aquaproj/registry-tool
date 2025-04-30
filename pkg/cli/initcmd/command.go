package initcmd

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/initcmd"
	"github.com/urfave/cli/v3"
)

type runner struct{}

func Command() *cli.Command {
	return (&runner{}).Command()
}

func (r *runner) Command() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     `Create configuration files`,
		UsageText: `$ aqua-registry init`,
		Description: `Create configuration files.

* aqua.yaml
* aqua-dev.yaml
`,
		Action: r.action,
	}
}

func (r *runner) action(ctx context.Context, _ *cli.Command) error {
	return initcmd.Init(ctx) //nolint:wrapcheck
}
