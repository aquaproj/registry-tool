package naming

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/scaffold"
)

func Resolve(ctx context.Context, logger *slog.Logger, pkgName string) (string, error) {
	if pkgName != "" {
		return strings.TrimPrefix(pkgName, "https://github.com/"), nil
	}

	branch, err := scaffold.GetCurrentBranch(ctx, logger)
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}

	if !strings.HasPrefix(branch, "feat/") {
		return "", errors.New("current branch must be feat/<package name> or you must give a package name")
	}

	return strings.TrimPrefix(branch, "feat/"), nil
}
