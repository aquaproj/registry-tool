package fix

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func Fix(ctx context.Context, logger *slog.Logger, args []string) error {
	for _, arg := range args {
		if err := fix(ctx, logger, arg); err != nil {
			return slogerr.With(err, "arg", arg) //nolint:wrapcheck
		}
	}
	return nil
}

func fix(ctx context.Context, logger *slog.Logger, targetFile string) error {
	switch filepath.Base(targetFile) {
	case "registry.yaml":
		pkgName, err := getPkgName(targetFile)
		if err != nil {
			return err
		}
		return fixRegistryYAML(pkgName, targetFile)
	case "pkg.yaml":
		return fixPkgYAML(targetFile)
	case "scaffold.yaml":
		return nil
	default:
		return fixPackage(ctx, logger, targetFile)
	}
}

func getPkgName(targetFile string) (string, error) {
	dir := filepath.ToSlash(filepath.Dir(targetFile))
	pkgName, ok := strings.CutPrefix(dir, "pkgs/")
	if !ok {
		return "", fmt.Errorf("the parent directory of %s must be pkgs/<package name>", targetFile)
	}
	return pkgName, nil
}

func writeFile(path string, data []byte) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.WriteFile(filepath.Clean(path), data, stat.Mode()); err != nil { //nolint:gosec
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
