package cli

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/cli/checkrepo"
	"github.com/aquaproj/registry-tool/pkg/cli/createprnewpkg"
	"github.com/aquaproj/registry-tool/pkg/cli/gengr"
	"github.com/aquaproj/registry-tool/pkg/cli/initcmd"
	"github.com/aquaproj/registry-tool/pkg/cli/mv"
	"github.com/aquaproj/registry-tool/pkg/cli/patchchecksum"
	"github.com/aquaproj/registry-tool/pkg/cli/scaffold"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, logE *logrus.Entry, ldFlags *urfave.LDFlags, args ...string) error {
	return urfave.Command(logE, ldFlags, &cli.Command{ //nolint:wrapcheck
		Name:  "aqua-registry",
		Usage: "CLI to develop aqua Registry. https://github.com/aquaproj/registry-tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				Sources: cli.EnvVars("AQUA_LOG_LEVEL"),
			},
		},
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			scaffold.Command(),
			createprnewpkg.Command(logE),
			gengr.Command(),
			initcmd.Command(),
			patchchecksum.Command(logE),
			checkrepo.Command(),
			mv.Command(),
		},
	}).Run(ctx, args)
}
