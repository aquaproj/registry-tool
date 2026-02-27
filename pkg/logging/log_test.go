package logging_test

import (
	"testing"

	"github.com/aquaproj/registry-tool/pkg/logging"
	"github.com/aquaproj/registry-tool/pkg/osarch"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if logE := logging.New(osarch.New(), "v1.6.0"); logE == nil {
		t.Fatal("logE must not be nil")
	}
}

func TestSetLevel(t *testing.T) {
	t.Parallel()
	logE := logging.New(osarch.New(), "v1.6.0")
	logging.SetLevel("debug", logE)
}
