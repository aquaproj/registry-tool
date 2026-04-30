package libc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/registry-tool/pkg/libc"
	"github.com/google/go-cmp/cmp"
)

type hasVariantCase struct {
	name    string
	yaml    string
	want    bool
	wantErr bool
}

var hasVariantCases = []hasVariantCase{ //nolint:gochecknoglobals
	{
		name: "key libc inside overrides variants",
		yaml: `packages:
  - type: github_release
    repo_owner: foo
    repo_name: bar
    overrides:
      - goos: linux
        variants:
          - key: libc
            value: musl
`,
		want: true,
	},
	{
		name: "key libc inside version_overrides overrides variants",
		yaml: `packages:
  - type: github_release
    repo_owner: foo
    repo_name: bar
    version_overrides:
      - version_constraint: "true"
        overrides:
          - goos: linux
            variants:
              - key: libc
                value: musl
`,
		want: true,
	},
	{
		name: "variants without libc key",
		yaml: `packages:
  - type: github_release
    repo_owner: foo
    repo_name: bar
    overrides:
      - goos: linux
        variants:
          - key: glibc
            value: x
`,
	},
	{
		name: "no variants at all",
		yaml: `packages:
  - type: github_release
    repo_owner: foo
    repo_name: bar
`,
	},
	{
		name: "empty packages",
		yaml: "packages: []\n",
	},
	{
		name:    "invalid yaml",
		yaml:    "packages: [not a list\n",
		wantErr: true,
	},
}

func TestHasVariant(t *testing.T) {
	t.Parallel()
	for _, tt := range hasVariantCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "registry.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := libc.HasVariant(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("HasVariant(-want +got):\n%s", diff)
			}
		})
	}
}

func TestHasVariant_FileNotExist(t *testing.T) {
	t.Parallel()
	got, err := libc.HasVariant(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("want false, got true")
	}
}
