package osexec

import (
	"log/slog"
	"os"
	"os/exec"
	"time"
)

const defaultWaitDelay = 1000 * time.Hour

func SetCancel(logger *slog.Logger, cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		logger.Warn("SIGINT is sent to cancel the command")
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = defaultWaitDelay
}
