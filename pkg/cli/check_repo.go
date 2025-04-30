package cli

import (
	"context"
	"net/http"

	"github.com/aquaproj/registry-tool/pkg/checkrepo"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

func (r *Runner) newCheckRepoCommand() *cli.Command {
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
		Action: r.checkRepoAction,
		Flags:  []cli.Flag{},
	}
}

func (r *Runner) checkRepoAction(ctx context.Context, cmd *cli.Command) error {
	return checkrepo.CheckRepo( //nolint:wrapcheck
		ctx, afero.NewOsFs(), &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		cmd.Args().First())
}
