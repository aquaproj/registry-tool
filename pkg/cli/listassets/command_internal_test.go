package listassets

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

type mockGH struct {
	release *github.RepositoryRelease
	pages   [][]*github.ReleaseAsset
	relErr  error
	listErr error
}

func (m *mockGH) GetReleaseByTag(context.Context, string, string, string) (*github.RepositoryRelease, *github.Response, error) {
	return m.release, &github.Response{}, m.relErr
}

func (m *mockGH) ListReleaseAssets(_ context.Context, _, _ string, _ int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	idx := opts.Page - 1
	if idx < 0 || idx >= len(m.pages) {
		return nil, &github.Response{}, nil
	}
	var nextPage int
	if idx+1 < len(m.pages) {
		nextPage = opts.Page + 1
	}
	return m.pages[idx], &github.Response{NextPage: nextPage}, nil
}

func Test_listAssets(t *testing.T) {
	t.Parallel()
	id := int64(1)
	rel := &github.RepositoryRelease{ID: &id}

	t.Run("release error", func(t *testing.T) {
		t.Parallel()
		err := listAssets(context.Background(), &mockGH{relErr: errors.New("not found")}, "o", "r", "v1")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("list error", func(t *testing.T) {
		t.Parallel()
		err := listAssets(context.Background(), &mockGH{release: rel, listErr: errors.New("fail")}, "o", "r", "v1")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("pagination", func(t *testing.T) {
		t.Parallel()
		page1 := make([]*github.ReleaseAsset, 100) //nolint:mnd
		for i := range page1 {
			n := fmt.Sprintf("a%d", i)
			page1[i] = &github.ReleaseAsset{Name: &n}
		}
		n := "last"
		page2 := []*github.ReleaseAsset{{Name: &n}}
		err := listAssets(context.Background(), &mockGH{
			release: rel,
			pages:   [][]*github.ReleaseAsset{page1, page2},
		}, "o", "r", "v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func Test_repoFromPkgName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pkg       string
		wantOwner string
		wantRepo  string
		isErr     bool
	}{
		{name: "simple", pkg: "cli/cli", wantOwner: "cli", wantRepo: "cli"},
		{name: "extra segments", pkg: "DeNA/unity-meta-check/gh-action", wantOwner: "DeNA", wantRepo: "unity-meta-check"},
		{name: "github.com prefix", pkg: "github.com/zeromicro/go-zero/tools/goctl", wantOwner: "zeromicro", wantRepo: "go-zero"},
		{name: "https github URL", pkg: "https://github.com/cli/cli", wantOwner: "cli", wantRepo: "cli"},
		{name: "non-github domain", pkg: "gitlab.com/gitlab-org/cli", isErr: true},
		{name: "single segment", pkg: "foo", isErr: true},
		{name: "empty owner", pkg: "/repo", isErr: true},
		{name: "empty name", pkg: "owner/", isErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, err := repoFromPkgName(tt.pkg)
			if (err != nil) != tt.isErr {
				t.Fatalf("repoFromPkgName(%q) error = %v, wantErr %v", tt.pkg, err, tt.isErr)
			}
			if err != nil {
				return
			}
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Fatalf("repoFromPkgName(%q) = (%q, %q), want (%q, %q)", tt.pkg, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}
