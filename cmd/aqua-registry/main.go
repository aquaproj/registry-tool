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
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	rt := runtime.New()
	logE := log.New(rt, version)
	if err := core(logE); err != nil {
		logerr.WithError(logE, err).Fatal("aqua failed")
	}
}

func core(logE *logrus.Entry) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return cli.Run(ctx, logE, &urfave.LDFlags{ //nolint:wrapcheck
		Version: version,
		Commit:  commit,
		Date:    date,
	}, os.Args...)
}
