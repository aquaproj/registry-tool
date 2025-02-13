package mv

import (
	"errors"
	"fmt"
	"strings"

	wast "github.com/aquaproj/aqua/v2/pkg/ast"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	goccyYAML "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func editRegistry(afs afero.Fs, newRegistryYAMLPath, oldPackageName, newPackageName string) error { //nolint:cyclop
	b, err := afero.ReadFile(afs, newRegistryYAMLPath)
	if err != nil {
		return fmt.Errorf("open a registry.yaml: %w", err)
	}

	cfg := &registry.Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("unmarshal registry.yaml as YAML: %w", err)
	}
	pkg, err := getPackageFromConfig(cfg, oldPackageName)
	if err != nil {
		return err
	}

	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}
	body := file.Docs[0].Body
	mv, err := wast.FindMappingValueFromNode(body, "packages")
	if err != nil {
		return fmt.Errorf("find a mapping node `packages`: %w", err)
	}
	seq, ok := mv.Value.(*ast.SequenceNode)
	if !ok {
		return errors.New("the value must be a sequence node")
	}
	for _, value := range seq.Values {
		idx, up, err := parseRegistryNode(value, oldPackageName, newPackageName, pkg)
		if err != nil {
			return err
		}
		if idx != 0 {
			// Insert aliases
			if err := insertAliases(value, idx, oldPackageName); err != nil {
				return err
			}
			up = true
		}
		if up {
			if err := afero.WriteFile(afs, newRegistryYAMLPath, []byte(file.String()), 0o644); err != nil { //nolint:mnd
				return fmt.Errorf("write registry.yaml: %w", err)
			}
			return nil
		}
	}
	return nil
}

func insertAliases(value ast.Node, idx int, oldPackageName string) error {
	// Insert aliases
	mv, ok := value.(*ast.MappingNode)
	if !ok {
		return errors.New("value must be a mapping node")
	}

	f, err := parser.ParseBytes([]byte("aliases:\n  - name: "+oldPackageName), parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse text as YAML: %w", err)
	}
	mn, ok := f.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		return errors.New("body must be a mapping node")
	}

	latterValues := make([]*ast.MappingValueNode, len(mv.Values[idx:]))
	copy(latterValues, mv.Values[idx:])
	mv.Values = mv.Values[:idx]
	mv.Merge(mn)
	mv.Merge(&ast.MappingNode{
		Values: latterValues,
	})
	return nil
}

func getPackageFromConfig(cfg *registry.Config, name string) (*registry.PackageInfo, error) {
	for _, pkgInfo := range cfg.PackageInfos {
		if name == pkgInfo.GetName() {
			return pkgInfo, nil
		}
	}
	return nil, errors.New("package isn't found in registry.yaml")
}

func appendNode(mapValue *ast.MappingValueNode, node ast.Node) error {
	switch mapValue.Value.Type() {
	case ast.NullType:
		mapValue.Value = node
		return nil
	case ast.SequenceType:
		if err := ast.Merge(mapValue.Value, node); err != nil {
			return fmt.Errorf("merge nodes: %w", err)
		}
		return nil
	default:
		return errors.New("node must be null or array")
	}
}

const wordRepoName = "repo_name"

func parseRegistryNode(node ast.Node, oldPackageName, newPackageName string, pkg *registry.PackageInfo) (int, bool, error) { //nolint:funlen,cyclop,gocognit
	mvs, err := wast.NormalizeMappingValueNodes(node)
	if err != nil {
		return 0, false, fmt.Errorf("normalize mapping value nodes: %w", err)
	}
	var newRepoOwner string
	var newRepoName string
	arr := strings.Split(newPackageName, "/")
	if len(arr) > 1 {
		newRepoOwner = arr[0]
		newRepoName = arr[1]
	}
	updated := false
	prevField := ""
	if pkg.Aliases == nil {
		if pkg.Name != "" {
			prevField = wordName
		} else if pkg.RepoName != "" {
			prevField = wordRepoName
		}
	}
	insertIdx := 0
	for i, mvn := range mvs {
		key, ok := mvn.Key.(*ast.StringNode)
		if !ok {
			continue
		}
		switch key.Value {
		case wordName:
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return 0, false, errors.New("name must be a string")
			}
			if sn.Value != oldPackageName {
				return 0, false, nil
			}
			sn.Value = newPackageName
			updated = true
			if prevField == wordName {
				insertIdx = i + 1
			}
		case "repo_owner":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return 0, false, errors.New("name must be a string")
			}
			if sn.Value != newRepoOwner {
				sn.Value = newRepoOwner
				updated = true
			}
		case wordRepoName:
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return 0, false, errors.New("name must be a string")
			}
			if sn.Value != newRepoName {
				sn.Value = newRepoName
				updated = true
			}
			if prevField == wordRepoName {
				insertIdx = i + 1
			}
		case "aliases":
			// Add oldPackageName
			aliases := []*registry.Alias{
				{
					Name: oldPackageName,
				},
			}
			node, err := goccyYAML.ValueToNode(aliases)
			if err != nil {
				return 0, false, fmt.Errorf("convert an alias to node: %w", err)
			}
			if err := appendNode(mvn, node); err != nil {
				return 0, false, fmt.Errorf("append the old package to aliases: %w", err)
			}
		default:
			continue // Ignore unknown fields
		}
	}
	return insertIdx, updated, nil
}
