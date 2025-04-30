package patchchecksum

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/patchchecksum"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

type runner struct {
	logE *logrus.Entry
}

func Command(logE *logrus.Entry) *cli.Command {
	return (&runner{logE: logE}).Command()
}

func (r *runner) Command() *cli.Command {
	return &cli.Command{
		Name:      "patch-checksum",
		Usage:     `Patch a checksum configuration`,
		UsageText: `$ aqua-registry patch-checksum <registry configuration file path>`,
		Description: `Patch a checksum configuration.

e.g.

$ aqua-registry patch-checksum pkgs/suzuki-shunsuke/tfcmt/registry.yaml
`,
		Action: r.action,
	}
}

func (r *runner) action(ctx context.Context, cmd *cli.Command) error {
	return patchchecksum.PatchChecksum(ctx, r.logE, cmd.Args().First()) //nolint:wrapcheck
}
