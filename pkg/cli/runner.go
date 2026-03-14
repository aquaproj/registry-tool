package cli

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/cli/checkrepo"
	connectcmd "github.com/aquaproj/registry-tool/pkg/cli/connect"
	"github.com/aquaproj/registry-tool/pkg/cli/createprnewpkg"
	"github.com/aquaproj/registry-tool/pkg/cli/gengr"
	"github.com/aquaproj/registry-tool/pkg/cli/gflag"
	"github.com/aquaproj/registry-tool/pkg/cli/initcmd"
	"github.com/aquaproj/registry-tool/pkg/cli/mv"
	"github.com/aquaproj/registry-tool/pkg/cli/patchchecksum"
	removecmd "github.com/aquaproj/registry-tool/pkg/cli/remove"
	removepackagecmd "github.com/aquaproj/registry-tool/pkg/cli/removepackage"
	"github.com/aquaproj/registry-tool/pkg/cli/resolveconflict"
	"github.com/aquaproj/registry-tool/pkg/cli/scaffold"
	startcmd "github.com/aquaproj/registry-tool/pkg/cli/start"
	stopcmd "github.com/aquaproj/registry-tool/pkg/cli/stop"
	testcmd "github.com/aquaproj/registry-tool/pkg/cli/test"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, logger *slogutil.Logger, env *urfave.Env) error {
	flags := &gflag.Flags{}
	return urfave.Command(env, &cli.Command{ //nolint:wrapcheck
		Name:  "aqua-registry",
		Usage: "CLI to develop aqua Registry. https://github.com/aquaproj/registry-tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "log level",
				Sources:     cli.EnvVars("AQUA_LOG_LEVEL"),
				Local:       true,
				Destination: &flags.LogLevel,
			},
		},
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			scaffold.Command(logger.Logger, flags),
			createprnewpkg.Command(logger.Logger),
			gengr.Command(),
			initcmd.Command(),
			patchchecksum.Command(logger.Logger),
			checkrepo.Command(),
			mv.Command(),
			connectcmd.Command(logger.Logger),
			removecmd.Command(logger.Logger),
			removepackagecmd.Command(logger.Logger),
			resolveconflict.Command(logger.Logger),
			startcmd.Command(logger.Logger),
			stopcmd.Command(logger.Logger),
			testcmd.Command(logger.Logger),
		},
	}).Run(ctx, env.Args)
}
