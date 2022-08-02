package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aquaproj/registry-tool/pkg/cli"
	"github.com/aquaproj/registry-tool/pkg/log"
	"github.com/aquaproj/registry-tool/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	rt := runtime.New()
	logE := log.New(rt, version)
	if err := core(logE, rt); err != nil {
		logerr.WithError(logE, err).Fatal("aqua failed")
	}
}

func core(logE *logrus.Entry, rt *runtime.Runtime) error {
	runner := cli.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &cli.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
		LogE:    logE,
		Runtime: rt,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runner.Run(ctx, os.Args...) //nolint:wrapcheck
}
