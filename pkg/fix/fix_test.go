package fix_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/registry-tool/pkg/fix"
)

type pkgYAMLTestCase struct {
	name      string
	source    string
	want      string
	wantError string
}

const (
	shortSyntaxSource = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo@v1.0.0
`
	shortSyntaxWant = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo
    version: v1.0.0
`
	longSyntaxSource = `packages:
  - name: owner/repo
    version: v1.0.1
  - name: owner/repo
    version: v1.0.0
`
	longSyntaxWant = `packages:
  - name: owner/repo@v1.0.1
  - name: owner/repo
    version: v1.0.0
`
	latestVersionCommentsSource = `packages:
  - name: owner/repo
    # Keep the reason for selecting this release.
    # This explanation intentionally spans multiple lines.
    version: v1.0.1 # Keep the latest version note too.
  - name: owner/repo
    version: v1.0.0
`
	latestVersionCommentsWant = `packages:
  # Keep the reason for selecting this release.
  # This explanation intentionally spans multiple lines.
  # Keep the latest version note too.
  - name: owner/repo@v1.0.1
  - name: owner/repo
    version: v1.0.0
`
	defaultRegistryCommentsSource = `packages:
  - name: owner/repo
    # Keep the reason for using the default registry.
    # This explanation intentionally spans multiple lines.
    registry: standard # Keep the registry note too.
    version: v1.0.1
  - name: owner/repo
    registry: standard # Keep the old registry note too.
    version: v1.0.0
`
	defaultRegistryCommentsWant = `packages:
  # Keep the reason for using the default registry.
  # This explanation intentionally spans multiple lines.
  # Keep the registry note too.
  - name: owner/repo@v1.0.1
  # Keep the old registry note too.
  - name: owner/repo
    version: v1.0.0
`
	omittedFieldCommentsSource = `packages:
  - name: owner/repo
    description: "" # Keep the latest description note.
    registry: standard # Keep the latest registry note.
    vars: {} # Keep the empty vars note.
    version: v1.0.1 # Keep the latest version note.
  - name: owner/repo@v1.0.0
    description: "" # Keep the empty description note.
`
	omittedFieldCommentsWant = `packages:
  # Keep the latest description note.
  # Keep the latest registry note.
  # Keep the empty vars note.
  # Keep the latest version note.
  - name: owner/repo@v1.0.1
  # Keep the empty description note.
  - name: owner/repo
    version: v1.0.0
`
	reconstructedSource = `packages:
- name: owner/repo@v2.0.0
- name: owner/repo@v1.0.0 # Keep this comment.
`
	reconstructedWant = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo # Keep this comment.
    version: v1.0.0
`
	standaloneCommentsSource = `packages:
  - name: owner/repo@v2.0.0
  # Keep the reason for testing this old release.
  # This comment intentionally spans multiple lines.
  - name: owner/repo@v1.0.0
    # Keep the variable explanation too.
    # It belongs to the vars field below.
    vars:
      go_version: 1.24.0
`
	standaloneCommentsWant = `packages:
  - name: owner/repo@v2.0.0
  # Keep the reason for testing this old release.
  # This comment intentionally spans multiple lines.
  - name: owner/repo
    version: v1.0.0
    # Keep the variable explanation too.
    # It belongs to the vars field below.
    vars:
      go_version: 1.24.0
`
	preservedDetailsSource = `packages:
  - name: owner/repo@v2.0.0
  # Keep this explanation.
  - name: "owner/repo@v1.0.0" # Keep this inline comment.
    vars:
      go_version: 1.24.0
  - name: owner/repo
    version: v0.9.0
`
	preservedDetailsWant = `packages:
  - name: owner/repo@v2.0.0
  # Keep this explanation.
  - name: owner/repo # Keep this inline comment.
    version: v1.0.0
    vars:
      go_version: 1.24.0
  - name: owner/repo
    version: v0.9.0
`
	ambiguousVersionsSource = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo@2026-08-27
  - name: owner/repo@true
  - name: owner/repo@null
  - name: owner/repo@1.0
  - name: owner/repo@01
  - name: owner/repo@.5
`
	ambiguousVersionsWant = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo
    version: "2026-08-27"
  - name: owner/repo
    version: "true"
  - name: owner/repo
    version: "null"
  - name: owner/repo
    version: "1.0"
  - name: owner/repo
    version: "01"
  - name: owner/repo
    version: ".5"
`
	mixedLineEndingsSource = "packages:\r\n" +
		"  - name: owner/repo@v2.0.0\n" +
		"  - name: owner/repo@v1.0.0\r\n"
	mixedLineEndingsWant = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo
    version: v1.0.0
`
	aquaShortSyntaxSource = `packages:
  - name: owner/repo@v2.0.0
  - name: 'owner/repo@variant@v1.0.0'
`
	aquaShortSyntaxWant = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo
    version: variant@v1.0.0
`
	aquaPackageSemanticsSource = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo@v1.0.0
    version: v0.9.0
  - name: owner/repo@v0.8.0
    version: null
`
	aquaPackageSemanticsWant = `packages:
  - name: owner/repo@v2.0.0
  - name: owner/repo
    version: v1.0.0
  - name: owner/repo
    version: v0.8.0
`
	quotedGoPackageSource = `packages:
  - name: "_go/sigsum.org/sigsum-go#cmd/sigsum-key@v0.9.1"
`
	goPackageWant = `packages:
  - name: _go/sigsum.org/sigsum-go#cmd/sigsum-key@v0.9.1
`
	invalidPkgYAML = `packages:
  - name: [
`
	unsupportedEntrySource = `packages:
  - name: owner/repo@v2.0.0
  - unsupported
  - name: owner/repo@v1.0.0
`
	nonSequencePackages = `packages: unsupported
`
	missingPackageName = `packages:
  - name: owner/repo@v2.0.0
  - version: v1.0.0
`
	missingPackageVersion = `packages:
  - name: owner/repo@
`
	unknownPackageField = `packages:
  - name: owner/repo@v1.0.0
    unsupported: true
`
)

func TestFixPkgYAML(t *testing.T) {
	t.Parallel()

	tests := []pkgYAMLTestCase{
		{name: "expands old package short syntax idempotently", source: shortSyntaxSource, want: shortSyntaxWant},
		{name: "uses short syntax only for the latest package", source: longSyntaxSource, want: longSyntaxWant},
		{name: "preserves comments from the latest version field", source: latestVersionCommentsSource, want: latestVersionCommentsWant},
		{name: "preserves comments from default registry fields", source: defaultRegistryCommentsSource, want: defaultRegistryCommentsWant},
		{name: "preserves comments from omitted empty fields", source: omittedFieldCommentsSource, want: omittedFieldCommentsWant},
		{name: "reconstructs YAML and preserves comments", source: reconstructedSource, want: reconstructedWant},
		{name: "preserves multi-line standalone comments", source: standaloneCommentsSource, want: standaloneCommentsWant},
		{name: "preserves old package entry details", source: preservedDetailsSource, want: preservedDetailsWant},
		{name: "quotes ambiguous versions", source: ambiguousVersionsSource, want: ambiguousVersionsWant},
		{name: "normalizes line endings", source: mixedLineEndingsSource, want: mixedLineEndingsWant},
		{name: "uses Aqua short syntax parsing", source: aquaShortSyntaxSource, want: aquaShortSyntaxWant},
		{name: "uses Aqua package parsing semantics", source: aquaPackageSemanticsSource, want: aquaPackageSemanticsWant},
		{name: "formats a Go package name without quotes", source: quotedGoPackageSource, want: goPackageWant},
		{name: "rejects a non-mapping package entry", source: unsupportedEntrySource, wantError: "parse pkg.yaml"},
		{name: "rejects a non-sequence packages value", source: nonSequencePackages, wantError: "parse pkg.yaml"},
		{name: "rejects a package without a name", source: missingPackageName, wantError: "packages[1].name must be specified"},
		{name: "rejects a package without a version", source: missingPackageVersion, wantError: "packages[0].version must be specified"},
		{name: "rejects an unknown package field", source: unknownPackageField, wantError: "unknown field"},
		{name: "rejects invalid YAML", source: invalidPkgYAML, wantError: "parse pkg.yaml"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := writePkgYAML(t, test.source)
			err := runFix(path)
			if test.wantError != "" {
				if err == nil {
					t.Error("expected an error")
				} else if !strings.Contains(err.Error(), test.wantError) {
					t.Errorf("error %q does not contain %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := readFile(t, path)
			if string(got) != test.want {
				t.Errorf("got:  %q\nwant: %q", got, test.want)
			}

			if err := runFix(path); err != nil {
				t.Fatal(err)
			}
			second := readFile(t, path)
			if string(second) != string(got) {
				t.Errorf("second run changed the file:\nfirst:\n%s\nsecond:\n%s", got, second)
			}
		})
	}
}

func writePkgYAML(t *testing.T, source string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "pkg.yaml")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runFix(path string) error {
	logger := slog.New(slog.DiscardHandler)
	if err := fix.Fix(context.Background(), logger, []string{path}); err != nil {
		return fmt.Errorf("fix pkg.yaml: %w", err)
	}
	return nil
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
