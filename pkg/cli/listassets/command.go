package listassets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/registry-tool/pkg/naming"
	"github.com/urfave/cli/v3"
)

// assetStateUploaded is the state of a GitHub Release asset whose upload has completed.
// GitHub keeps assets whose upload was interrupted with the state "starter".
// They are hidden from the release page and can't be downloaded,
// so listing them would mislead you into configuring an asset that doesn't exist.
const assetStateUploaded = "uploaded"

type ghClient interface {
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error)
}

func repoFromPkgName(name string) (string, string, error) {
	name = strings.TrimPrefix(name, "https://github.com/")
	name = strings.TrimPrefix(name, "github.com/")
	parts := strings.SplitN(name, "/", 3)                   //nolint:mnd
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" { //nolint:mnd
		return "", "", fmt.Errorf("package name %q does not contain owner/repo", name)
	}
	if strings.Contains(parts[0], ".") {
		return "", "", fmt.Errorf("package %q is not a GitHub repository, pass explicit owner/repo", name)
	}
	return parts[0], parts[1], nil
}

func listAssets(ctx context.Context, stdout io.Writer, client ghClient, owner, name, version string) error {
	release, _, err := client.GetReleaseByTag(ctx, owner, name, version)
	if err != nil {
		return fmt.Errorf("get release by tag %s: %w", version, err)
	}
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100, //nolint:mnd
	}
	for {
		assets, resp, err := client.ListReleaseAssets(ctx, owner, name, release.GetID(), opts)
		if err != nil {
			return fmt.Errorf("list release assets: %w", err)
		}
		for _, asset := range assets {
			if asset.GetState() != assetStateUploaded {
				continue
			}
			fmt.Fprintln(stdout, asset.GetName())
		}
		if resp.NextPage == 0 {
			return nil
		}
		opts.Page = resp.NextPage
	}
}

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "list-assets",
		Aliases:   []string{"lsa"},
		Usage:     "List release assets of a GitHub Release",
		UsageText: "argd list-assets [owner/repo] <version>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			var owner, name, version string
			switch len(args) {
			case 2: //nolint:mnd
				var err error
				owner, name, err = repoFromPkgName(args[0])
				if err != nil {
					return err
				}
				version = args[1]
			case 1:
				pkgName, err := naming.Resolve(ctx, logger, "")
				if err != nil {
					return fmt.Errorf("resolve package name: %w", err)
				}
				owner, name, err = repoFromPkgName(pkgName)
				if err != nil {
					return err
				}
				version = args[0]
			default:
				return errors.New("usage: argd list-assets [owner/repo] <version>")
			}
			gh, err := github.New(ctx, logger)
			if err != nil {
				return fmt.Errorf("create github client: %w", err)
			}
			return listAssets(ctx, os.Stdout, gh, owner, name, version)
		},
	}
}
