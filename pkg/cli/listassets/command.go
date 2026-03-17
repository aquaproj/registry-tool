package listassets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/urfave/cli/v3"
)

type ghClient interface {
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error)
}

func parseRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 { //nolint:mnd
		return "", "", fmt.Errorf("invalid repo format %q, expected <owner>/<repo>", repo)
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo format %q, expected <owner>/<repo>", repo)
	}
	return parts[0], parts[1], nil
}

func listAssets(ctx context.Context, client ghClient, owner, name, version string) error {
	release, _, err := client.GetReleaseByTag(ctx, owner, name, version)
	if err != nil {
		return fmt.Errorf("get release by tag %s: %w", version, err)
	}
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100, //nolint:mnd
	}
	for {
		assets, _, err := client.ListReleaseAssets(ctx, owner, name, release.GetID(), opts)
		if err != nil {
			return fmt.Errorf("list release assets: %w", err)
		}
		for _, asset := range assets {
			fmt.Println(asset.GetName()) //nolint:forbidigo
		}
		if len(assets) < opts.PerPage {
			return nil
		}
		opts.Page++
	}
}

func Command(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:      "list-assets",
		Aliases:   []string{"lsa"},
		Usage:     "List release assets of a GitHub Release",
		UsageText: "argd list-assets <owner/repo> <version>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) != 2 { //nolint:mnd
				return errors.New("usage: argd list-assets <owner/repo> <version>")
			}
			owner, name, err := parseRepo(args[0])
			if err != nil {
				return err
			}
			return listAssets(ctx, github.New(ctx, logger), owner, name, args[1])
		},
	}
}
