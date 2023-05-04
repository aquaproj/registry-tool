package cli

import (
	"context"
	"io"
	"time"

	"github.com/aquaproj/registry-tool/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
	compiledDate, err := time.Parse(time.RFC3339, runner.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:     "aqua-registry",
		Usage:    "CLI to develop aqua Registry. https://github.com/aquaproj/registry-tool",
		Version:  runner.LDFlags.Version + " (" + runner.LDFlags.Commit + ")",
		Compiled: compiledDate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				EnvVars: []string{"AQUA_LOG_LEVEL"},
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			runner.newScaffoldCommand(),
			runner.newCreatePRNewPkgCommand(),
			runner.newGenerateRegistryCommand(),
			runner.newCompletionCommand(),
			runner.newVersionCommand(),
			runner.newInitCommand(),
			runner.newPatchChecksumCommand(),
			runner.newCheckRepoCommand(),
		},
	}

	return app.RunContext(ctx, args) //nolint:wrapcheck
}
