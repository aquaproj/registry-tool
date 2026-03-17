package listassets

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

func Test_parseRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		repo  string
		isErr bool
	}{
		{name: "valid", repo: "owner/repo"},
		{name: "no slash", repo: "repo", isErr: true},
		{name: "too many slashes", repo: "a/b/c", isErr: true},
		{name: "empty owner", repo: "/repo", isErr: true},
		{name: "empty name", repo: "owner/", isErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := parseRepo(tt.repo)
			if (err != nil) != tt.isErr {
				t.Fatalf("parseRepo(%q) error = %v, wantErr %v", tt.repo, err, tt.isErr)
			}
		})
	}
}

type mockGH struct {
	release *github.RepositoryRelease
	pages   [][]*github.ReleaseAsset
	relErr  error
	listErr error
}

func (m *mockGH) GetReleaseByTag(context.Context, string, string, string) (*github.RepositoryRelease, *github.Response, error) {
	return m.release, nil, m.relErr
}

func (m *mockGH) ListReleaseAssets(_ context.Context, _, _ string, _ int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	idx := opts.Page - 1
	if idx < 0 || idx >= len(m.pages) {
		return nil, nil, nil
	}
	return m.pages[idx], nil, nil
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
