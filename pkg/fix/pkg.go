package fix

import (
	"errors"
	"fmt"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/goccy/go-yaml"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

// pkgYAML is the typed representation of pkg.yaml.
// It embeds Aqua's package model so pkg.yaml uses Aqua's parsing rules.
type pkgYAML struct {
	Packages []*aqua.Package `yaml:"packages"`
}

// fixPkgYAML normalizes pkgFile.
// The latest package uses the short syntax `<name>@<version>`,
// and the other packages use the long syntax to prevent aqua-registry-updater
// from updating them.
func fixPkgYAML(pkgFile string) error {
	src, err := os.ReadFile(pkgFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", pkgFile, err)
	}
	formatted, err := formatPkgYAML(pkgFile, src)
	if err != nil {
		return err
	}
	if string(formatted) == string(src) {
		return nil
	}
	return writeFile(pkgFile, formatted)
}

// formatPkgYAML parses package data and normalizes it.
func formatPkgYAML(pkgFile string, src []byte) ([]byte, error) {
	pkg := &pkgYAML{}
	if err := yaml.UnmarshalWithOptions(src, pkg, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("parse pkg.yaml: %w", err)
	}
	actions, err := pkg.actions()
	if err != nil {
		return nil, err
	}
	s, err := yamledit.EditBytes(pkgFile, src, actions...)
	if err != nil {
		return nil, fmt.Errorf("edit pkg.yaml: %w", err)
	}
	return []byte(s), nil
}

// actions builds the actions normalizing every package.
func (p *pkgYAML) actions() ([]yamledit.Action, error) {
	if len(p.Packages) == 0 {
		return nil, errors.New("packages must not be empty")
	}
	actions := make([]yamledit.Action, 0, len(p.Packages))
	for i, pkg := range p.Packages {
		if pkg == nil || pkg.Name == "" {
			return nil, fmt.Errorf("packages[%d].name must be specified", i)
		}
		if pkg.Version == "" {
			return nil, fmt.Errorf("packages[%d].version must be specified", i)
		}
		actions = append(actions, yamledit.MapAction(
			fmt.Sprintf("$.packages[%d]", i), packageActions(i, pkg)...))
	}
	return actions, nil
}

// packageActions normalizes a package.
// The first package is the latest version, and the others are old versions.
func packageActions(idx int, pkg *aqua.Package) []yamledit.MappingNodeAction {
	var removed []any
	if pkg.Registry == aqua.RegistryTypeStandard {
		// standard is the default registry, so the field is redundant.
		removed = append(removed, "registry")
	}
	if idx != 0 {
		return []yamledit.MappingNodeAction{
			yamledit.RemoveKeys(removed...),
			yamledit.SetKey("name", packageName(pkg.Name), nil),
			yamledit.SetKey("version", pkg.Version, &yamledit.SetKeyOption{
				InsertLocations: []*yamledit.InsertLocation{{AfterKey: "name"}},
			}),
		}
	}
	// The latest package uses the short syntax, so the version field is removed.
	removed = append(removed, "version")
	return []yamledit.MappingNodeAction{
		yamledit.RemoveKeys(removed...),
		yamledit.SetKey("name", packageName(pkg.Name+"@"+pkg.Version), nil),
	}
}

// packageName keeps a package name unquoted if it is valid as a plain scalar.
// go-yaml quotes names including '#' such as `_go/example.com/foo#cmd/bar`,
// but they don't have to be quoted and quoting them causes a lint failure.
func packageName(name string) any {
	var s string
	if err := yaml.Unmarshal([]byte(name), &s); err != nil || s != name {
		return name
	}
	return yamledit.NewBytes([]byte(name))
}
