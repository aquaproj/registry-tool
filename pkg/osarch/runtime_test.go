package osarch_test

import (
	"testing"

	"github.com/aquaproj/registry-tool/pkg/osarch"
)

func TestNew(t *testing.T) {
	t.Parallel()
	rt := osarch.New()
	if rt == nil {
		t.Fatal("runtime must not be nil")
	}
	if rt.GOOS == "" {
		t.Fatal("rt.GOOS is empty")
	}
	if rt.GOARCH == "" {
		t.Fatal("rt.GOARCH is empty")
	}
}
