package cli

import (
	"github.com/aquaproj/registry-tool/pkg/checkrepo"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newCheckRepoCommand() *cli.Command {
	return &cli.Command{
		Name:      "check-repo",
		Usage:     `Check if GitHub Repository was transferred`,
		UsageText: `$ aqua-registry check-repo <package name>`,
		Description: `Check if GitHub Repository is transferred.
This command succeeds if the repository isn't transferred.

e.g.

$ aqua-registry check-repo Azure/aztfy
Azure/aztfexport
`,
		Action: runner.checkRepoAction,
		Flags:  []cli.Flag{},
	}
}

func (runner *Runner) checkRepoAction(c *cli.Context) error {
	return checkrepo.CheckRepo(c.Context, c.Bool("fix"), c.Args().First()) //nolint:wrapcheck
}
