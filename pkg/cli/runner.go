package cli

import (
	"context"

	"github.com/aquaproj/registry-tool/pkg/cli/checkrepo"
	"github.com/aquaproj/registry-tool/pkg/cli/createprnewpkg"
	"github.com/aquaproj/registry-tool/pkg/cli/gengr"
	"github.com/aquaproj/registry-tool/pkg/cli/gflag"
	"github.com/aquaproj/registry-tool/pkg/cli/initcmd"
	"github.com/aquaproj/registry-tool/pkg/cli/mv"
	"github.com/aquaproj/registry-tool/pkg/cli/patchchecksum"
	"github.com/aquaproj/registry-tool/pkg/cli/scaffold"
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
			scaffold.Command(flags),
			createprnewpkg.Command(logger.Logger),
			gengr.Command(),
			initcmd.Command(),
			patchchecksum.Command(logger.Logger),
			checkrepo.Command(),
			mv.Command(),
		},
	}).Run(ctx, env.Args)
}
