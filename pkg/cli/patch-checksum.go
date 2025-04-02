package cli

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/patchchecksum"
	"github.com/urfave/cli/v3"
)

func (runner *Runner) newPatchChecksumCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch-checksum",
		Usage:     `Patch a checksum configuration`,
		UsageText: `$ aqua-registry patch-checksum <registry configuration file path>`,
		Description: `Patch a checksum configuration.

e.g.

$ aqua-registry patch-checksum pkgs/suzuki-shunsuke/tfcmt/registry.yaml
`,
		Action: runner.patchChecksumAction,
	}
}

func (runner *Runner) patchChecksumAction(ctx context.Context, cmd *cli.Command) error {
	return patchchecksum.PatchChecksum(ctx, runner.LogE, cmd.Args().First()) //nolint:wrapcheck
}
