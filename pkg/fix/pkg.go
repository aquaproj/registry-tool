package fix

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

// pkgYAML is the typed representation used to parse and render pkg.yaml.
type pkgYAML struct {
	Packages []*pkgYAMLPackage `yaml:"packages"`
	comments yaml.CommentMap
}

// pkgYAMLPackage embeds Aqua's package model so pkg.yaml uses Aqua's parsing rules.
type pkgYAMLPackage struct {
	aqua.Package `yaml:",inline"`
}

// pkgYAMLPackageOutput overrides name rendering and inlines every other Aqua field.
type pkgYAMLPackageOutput struct {
	Name    pkgYAMLPackageName `yaml:"name"`
	Package aqua.Package       `yaml:",inline"`
}

// pkgYAMLPackageName controls the YAML scalar style used for package names.
type pkgYAMLPackageName string

// fixPkgYAML formats pkgFile and writes it only when the rendered content changes.
func fixPkgYAML(pkgFile string) error {
	source, err := os.ReadFile(pkgFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", pkgFile, err)
	}

	formatted, err := formatPkgYAML(source)
	if err != nil {
		return err
	}
	if bytes.Equal(formatted, source) {
		return nil
	}
	return writeFile(pkgFile, formatted)
}

// formatPkgYAML parses package data, normalizes it, and renders the typed result.
func formatPkgYAML(source []byte) ([]byte, error) {
	pkg, err := parsePkgYAML(source)
	if err != nil {
		return nil, err
	}
	pkg.normalize()
	return pkg.marshal()
}

// parsePkgYAML decodes the complete document into package structs and captures comments.
func parsePkgYAML(source []byte) (*pkgYAML, error) {
	pkg := &pkgYAML{
		comments: yaml.CommentMap{},
	}
	if err := yaml.UnmarshalWithOptions(
		source,
		pkg,
		yaml.Strict(),
		yaml.CommentToMap(pkg.comments),
	); err != nil {
		return nil, fmt.Errorf("parse pkg.yaml: %w", err)
	}
	if err := pkg.validate(); err != nil {
		return nil, err
	}
	return pkg, nil
}

// validate rejects package data that cannot be represented in the uniform format.
func (p *pkgYAML) validate() error {
	if len(p.Packages) == 0 {
		return errors.New("packages must not be empty")
	}
	for i, pkg := range p.Packages {
		if pkg == nil || pkg.Name == "" {
			return fmt.Errorf("packages[%d].name must be specified", i)
		}
		if strings.Contains(pkg.Name, "@") {
			return fmt.Errorf("packages[%d].name is invalid", i)
		}
		if pkg.Version == "" {
			return fmt.Errorf("packages[%d].version must be specified", i)
		}
	}
	return nil
}

// normalize makes the latest package short form and every older package long form.
func (p *pkgYAML) normalize() {
	for _, pkg := range p.Packages {
		if pkg.Registry == aqua.RegistryTypeStandard {
			pkg.Registry = ""
		}
	}

	latest := p.Packages[0]
	latest.Name += "@" + latest.Version
	latest.Version = ""
}

// moveOrphanedComments relocates comments whose fields were omitted from the typed output.
func (p *pkgYAML) moveOrphanedComments(formatted []byte) error {
	if len(p.comments) == 0 {
		return nil
	}
	file, err := parser.ParseBytes(formatted, 0)
	if err != nil {
		return fmt.Errorf("parse formatted pkg.yaml: %w", err)
	}
	// Take a sorted snapshot because the loop relocates entries within p.comments.
	paths := slices.Sorted(maps.Keys(p.comments))
	for _, commentPath := range paths {
		cPath, err := yaml.PathString(commentPath)
		if err != nil {
			return fmt.Errorf("parse comment path %s: %w", commentPath, err)
		}
		// Check whether the comment's original path survived struct marshaling.
		if _, err := cPath.FilterFile(file); err == nil {
			continue
		} else if !yaml.IsNotFoundNodeError(err) {
			return fmt.Errorf("find comment path %s: %w", commentPath, err)
		}
		packageIndex, ok := packageIndexFromCommentPath(commentPath)
		if !ok {
			return fmt.Errorf("comment path %s has no package index", commentPath)
		}
		p.moveComments(commentPath, packageIndex)
	}
	return nil
}

// packageIndexFromCommentPath extracts the package index from a comment path.
func packageIndexFromCommentPath(path string) (int, bool) {
	rest, ok := strings.CutPrefix(path, "$.packages[")
	if !ok {
		return 0, false
	}
	indexText, _, ok := strings.Cut(rest, "]")
	if !ok {
		return 0, false
	}
	index, err := strconv.Atoi(indexText)
	return index, err == nil
}

// moveComments makes comments from an omitted field standalone above its package.
func (p *pkgYAML) moveComments(commentPath string, packageIndex int) {
	packagePath := fmt.Sprintf("$.packages[%d]", packageIndex)
	comments, ok := p.comments[commentPath]
	if !ok {
		return
	}
	for _, comment := range comments {
		comment = yaml.HeadComment(comment.Texts...)
		p.comments[packagePath] = appendYAMLComment(p.comments[packagePath], comment)
	}
	delete(p.comments, commentPath)
}

// appendYAMLComment combines comments that must occupy the same YAML position.
func appendYAMLComment(comments []*yaml.Comment, addition *yaml.Comment) []*yaml.Comment {
	for _, comment := range comments {
		if comment.Position == addition.Position {
			comment.Texts = append(comment.Texts, addition.Texts...)
			return comments
		}
	}
	return append(comments, addition)
}

// MarshalYAML renders a parsed package with pkg.yaml-specific name formatting.
func (p pkgYAMLPackage) MarshalYAML() (any, error) {
	return pkgYAMLPackageOutput{
		Name:    pkgYAMLPackageName(p.Name),
		Package: p.Package,
	}, nil
}

// MarshalYAML uses plain style when a package name is safe without quotes.
func (n pkgYAMLPackageName) MarshalYAML() ([]byte, error) {
	name := string(n)
	// Ignore '#' when checking whether another character still requires quotes.
	withoutHashes := strings.ReplaceAll(name, "#", "x")
	if !strings.HasPrefix(name, "#") &&
		!strings.ContainsAny(name, " \t\r\n") &&
		!token.IsNeedQuoted(withoutHashes) {
		return []byte(name), nil
	}
	b, err := yaml.Marshal(name)
	if err != nil {
		return nil, fmt.Errorf("marshal package name: %w", err)
	}
	return b, nil
}

// marshal renders the normalized package structs with the captured comments.
func (p *pkgYAML) marshal() ([]byte, error) {
	formatted, err := yaml.MarshalWithOptions(p, yaml.IndentSequence(true))
	if err != nil {
		return nil, fmt.Errorf("marshal pkg.yaml: %w", err)
	}
	if err := p.moveOrphanedComments(formatted); err != nil {
		return nil, err
	}
	formatted, err = yaml.MarshalWithOptions(
		p,
		yaml.IndentSequence(true),
		yaml.WithComment(p.comments),
	)
	if err != nil {
		return nil, fmt.Errorf("marshal pkg.yaml: %w", err)
	}
	return formatted, nil
}
