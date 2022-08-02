package cli

import (
	newpkg "github.com/aquaproj/registry-tool/pkg/create-pr-new-pkg"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newCreatePRNewPkgCommand() *cli.Command {
	return &cli.Command{
		Name:      "create-pr-new-pkg",
		Usage:     `Create a pull request to add new packages`,
		UsageText: `aqua-registry create-pr-new-pkg <package name> [<package name> ...]`,
		Description: `Create a pull request to add new packages.

e.g.

$ aqua-registry create-pr-new-pkg cli/cli

This tool does the following things.

1. Create a feature branch
2. Create a commit
3. Push the commit to origin
4. Open a web browser to create a request with GitHub CLI
`,
		Action: runner.createPRNewPkgAction,
	}
}

func (runner *Runner) createPRNewPkgAction(c *cli.Context) error {
	return newpkg.CreatePRNewPkgs(c.Context, c.Args().Slice()...) //nolint:wrapcheck
}
