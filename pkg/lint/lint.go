package lint

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/registry-tool/pkg/naming"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"gopkg.in/yaml.v3"
)

func Lint(ctx context.Context, logger *slog.Logger, pkgName string) error {
	pkgName, err := naming.Resolve(ctx, logger, pkgName)
	if err != nil {
		return fmt.Errorf("resolve package name: %w", err)
	}

	pkgDir := filepath.Join(append([]string{"pkgs"}, strings.Split(pkgName, "/")...)...)
	pkgFile := filepath.Join(pkgDir, "pkg.yaml")

	b, err := os.ReadFile(pkgFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", pkgFile, err)
	}

	var cfg aqua.Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("parse %s: %w", pkgFile, err)
	}

	if len(cfg.Packages) == 0 {
		return slogerr.With(errors.New("packages is empty"), "file", pkgFile)
	}

	return nil
}
