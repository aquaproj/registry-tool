package mv

import (
	"errors"
	"fmt"
	"strings"

	wast "github.com/aquaproj/aqua/v2/pkg/ast"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/spf13/afero"
)

func editPackageYAML(afs afero.Fs, pkgYAMLPath string, oldPackageName, newPackageName string) error {
	b, err := afero.ReadFile(afs, pkgYAMLPath)
	if err != nil {
		return fmt.Errorf("read pkg.yaml: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	body := file.Docs[0].Body // DocumentNode

	mv, err := wast.FindMappingValueFromNode(body, "packages")
	if err != nil {
		return fmt.Errorf("find a mapping node `packages`: %w", err)
	}

	seq, ok := mv.Value.(*ast.SequenceNode)
	if !ok {
		return errors.New("the value must be a sequence node")
	}
	updated, err := editPackageAST(seq.Values, oldPackageName, newPackageName)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}
	if err := afero.WriteFile(afs, pkgYAMLPath, []byte(file.String()), 0o644); err != nil { //nolint:gomnd
		return fmt.Errorf("write registry.yaml: %w", err)
	}
	return nil
}

func editPackageAST(values []ast.Node, oldPackageName, newPackageName string) (bool, error) {
	updated := false
	for _, value := range values {
		up, err := parsePackageNode(value, oldPackageName, newPackageName)
		if err != nil {
			return false, err
		}
		if up {
			updated = up
		}
	}
	return updated, nil
}

func parsePackageNode(node ast.Node, oldPackageName, newPackageName string) (bool, error) {
	mvs, err := wast.NormalizeMappingValueNodes(node)
	if err != nil {
		return false, fmt.Errorf("normalize mapping value nodes: %w", err)
	}
	for _, mvn := range mvs {
		key, ok := mvn.Key.(*ast.StringNode)
		if !ok {
			continue
		}
		switch key.Value {
		case "name":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return false, errors.New("name must be a string")
			}
			name, version, ok := strings.Cut(sn.Value, "@")
			if name != oldPackageName {
				return false, nil
			}
			if ok {
				sn.Value = fmt.Sprintf("%s@%s", newPackageName, version)
				return true, nil
			}
			sn.Value = newPackageName
			return true, nil
		default:
			continue // Ignore unknown fields
		}
	}
	return false, nil
}
