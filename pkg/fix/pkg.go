package fix

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
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
			fmt.Sprintf("$.packages[%d]", i), packageAction(i, pkg)))
	}
	return actions, nil
}

// packageAction normalizes a package.
// The first package is the latest version, and the others are old versions.
func packageAction(idx int, pkg *aqua.Package) yamledit.MappingNodeAction {
	return yamledit.EditMapAction(func(m *yamledit.Map[any, any]) error {
		version := packageVersion(m, pkg)
		var actions []yamledit.MappingNodeAction
		if pkg.Registry == aqua.RegistryTypeStandard {
			// standard is the default registry, so the field is redundant.
			actions = append(actions, yamledit.RemoveKeys("registry"))
		}
		if idx == 0 {
			// The latest package uses the short syntax `<name>@<version>`.
			actions = append(actions,
				yamledit.RemoveKeys("version"),
				yamledit.SetKey("name", packageName(pkg.Name+"@"+version), nil))
		} else {
			actions = append(actions,
				yamledit.SetKey("name", packageName(pkg.Name), nil),
				yamledit.SetKey("version", version, &yamledit.SetKeyOption{
					InsertLocations: []*yamledit.InsertLocation{{AfterKey: "name"}},
				}))
		}
		for _, action := range actions {
			if err := action.Run(m.Node); err != nil {
				return fmt.Errorf("normalize a package: %w", err)
			}
		}
		return nil
	})
}

// packageVersion returns the version as it is written in the file.
// Aqua decodes the version field into a string, so an unquoted version such as
// 1.10 or 010 would come back as 1.1 or 8. The scalar token keeps the original
// text, and quoting it on output makes Aqua read it correctly too.
// If the name has the short syntax, the version is a part of the name and Aqua
// ignores the version field, so the decoded version is used as it is.
func packageVersion(m *yamledit.Map[any, any], pkg *aqua.Package) string {
	name, ok := m.Map["name"]
	if !ok {
		return pkg.Version
	}
	if strings.Contains(scalarText(name.Node.Value), "@") {
		return pkg.Version
	}
	version, ok := m.Map["version"]
	if !ok {
		return pkg.Version
	}
	if text := scalarText(version.Node.Value); text != "" {
		return text
	}
	return pkg.Version
}

// scalarText returns the text of a scalar node without quotes.
func scalarText(node ast.Node) string {
	if node == nil {
		return ""
	}
	tk := node.GetToken()
	if tk == nil {
		return ""
	}
	return tk.Value
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
