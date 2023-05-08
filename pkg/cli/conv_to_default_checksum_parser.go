package cli

import (
	"github.com/aquaproj/registry-tool/pkg/convtodefaultchecksumparser"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newConvToDefaultChecksumParserCommand() *cli.Command {
	return &cli.Command{
		Name:      "conv-to-default-checksum-parser",
		Usage:     `Convert the checksum parser to the default checksum parser`,
		UsageText: `$ aqua-registry conv-to-default-checksum-parser <registry file path>`,
		Description: `Convert the checksum parser to the default checksum parser

e.g.

$ aqua-registry conv-to-default-checksum-parser cli/cli
`,
		Action: runner.convToDefaultChecksumParserAction,
	}
}

func (runner *Runner) convToDefaultChecksumParserAction(c *cli.Context) error {
	return convtodefaultchecksumparser.Convert(c.Args().Slice()...) //nolint:wrapcheck
}
