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

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	return helpall.With(&cli.Command{ //nolint:wrapcheck
		Name:    "aqua-registry",
		Usage:   "CLI to develop aqua Registry. https://github.com/aquaproj/registry-tool",
		Version: runner.LDFlags.Version + " (" + runner.LDFlags.Commit + ")",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				Sources: cli.EnvVars("AQUA_LOG_LEVEL"),
			},
		},
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			runner.newScaffoldCommand(),
			runner.newCreatePRNewPkgCommand(),
			runner.newGenerateRegistryCommand(),
			runner.newCompletionCommand(),
			runner.newInitCommand(),
			runner.newPatchChecksumCommand(),
			runner.newCheckRepoCommand(),
			runner.newMVCommand(),
			vcmd.New(&vcmd.Command{
				Name:    "aqua-registry",
				Version: runner.LDFlags.Version,
				SHA:     runner.LDFlags.Commit,
			}),
		},
	}, nil).Run(ctx, args)
}
