package cli

import (
	"context"
	"io"

	"github.com/aquaproj/registry-tool/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/helpall"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/vcmd"
	"github.com/urfave/cli/v3"
)

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
	LogE    *logrus.Entry
	Runtime *runtime.Runtime
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (r *Runner) Run(ctx context.Context, args ...string) error {
	return helpall.With(&cli.Command{ //nolint:wrapcheck
		Name:    "aqua-registry",
		Usage:   "CLI to develop aqua Registry. https://github.com/aquaproj/registry-tool",
		Version: r.LDFlags.Version + " (" + r.LDFlags.Commit + ")",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				Sources: cli.EnvVars("AQUA_LOG_LEVEL"),
			},
		},
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			r.newScaffoldCommand(),
			r.newCreatePRNewPkgCommand(),
			r.newGenerateRegistryCommand(),
			r.newCompletionCommand(),
			r.newInitCommand(),
			r.newPatchChecksumCommand(),
			r.newCheckRepoCommand(),
			r.newMVCommand(),
			vcmd.New(&vcmd.Command{
				Name:    "aqua-registry",
				Version: r.LDFlags.Version,
				SHA:     r.LDFlags.Commit,
			}),
		},
	}, nil).Run(ctx, args)
}
