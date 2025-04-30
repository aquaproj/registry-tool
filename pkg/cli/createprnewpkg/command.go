package createprnewpkg

import (
	"context"

	newpkg "github.com/aquaproj/registry-tool/pkg/create-pr-new-pkg"
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
		Action: r.action,
	}
}

func (r *runner) action(ctx context.Context, cmd *cli.Command) error {
	return newpkg.CreatePRNewPkgs(ctx, r.logE, cmd.Args().Slice()...) //nolint:wrapcheck
}
