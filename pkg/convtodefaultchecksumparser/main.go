package convtodefaultchecksumparser

import (
	"errors"
	"fmt"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

func Convert(filePaths ...string) error {
	for _, filePath := range filePaths {
		if err := convert(filePath); err != nil {
			return logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"file_path": filePath,
			})
		}
	}
	return nil
}

func convert(filePath string) error { //nolint:cyclop
	cfg := &registry.Config{}
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("open a configuration file %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("parse a configuration file %s as YAML: %w", filePath, err)
	}

	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	visitor := &Visitor{}
	ast.Walk(visitor, file.Docs[0])

	// changed := false

	// // if f, err := removeWithPath(file); err != nil {
	// // 	return err
	// // } else if f {
	// // 	changed = true
	// // }

	// pkgsAST, err := patchchecksum.GetPackagesAST(file)
	// if err != nil {
	// 	return err //nolint:wrapcheck
	// }

	// for i, pkgInfo := range cfg.PackageInfos {
	// 	pkgInfoNode, j := patchchecksum.FindFirstMappingNode(pkgsAST, i)
	// 	if j == -1 {
	// 		return nil
	// 	}

	// 	if f, err := convChecksumParserOfVersionOverrides(pkgInfoNode, pkgInfo); err != nil {
	// 		return fmt.Errorf("convert to the default checksum parser: %w", err)
	// 	} else if f {
	// 		changed = true
	// 	}

	// 	if pkgInfo.Checksum != nil && pkgInfo.Checksum.FileFormat != "" {
	// 		if f, err := removeFromChecksumParent(pkgInfoNode); err != nil {
	// 			return fmt.Errorf("convert to the default checksum parser: %w", err)
	// 		} else if f {
	// 			changed = true
	// 		}
	// 	}
	// }

	// if !changed {
	// 	return nil
	// }

	if !visitor.changed {
		return nil
	}

	if err := os.WriteFile(filePath, []byte(file.String()), 0o644); err != nil { //nolint:gosec,gomnd
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}

// func removeWithPath(file *ast.File) (bool, error) {
// 	ypath, err := goccyYAML.PathString("$.packages[*].version_constraints[*].checksum")
// 	if err != nil {
// 		return false, fmt.Errorf("create a yaml path from string: %w", err)
// 	}
// 	node, err := ypath.FilterFile(file)
// 	if err != nil {
// 		return false, fmt.Errorf("filter a file with yaml path: %w", err)
// 	}
// 	return removeFromChecksum(node)
// }

func getMappingValueNode(mappingNode *ast.MappingNode, key string) *ast.MappingValueNode {
	for _, mvNode := range mappingNode.Values {
		if mvNode.Key.String() == key {
			return mvNode
		}
	}
	return nil
}

// func removeFromChecksum(node ast.Node) (bool, error) {
// 	checksumNode, ok := node.(*ast.MappingNode)
// 	if !ok {
// 		return false, logerr.WithFields(errors.New("checksum must be *ast.MappingNode"), logrus.Fields{
// 			"node_path": node.GetPath(),
// 		})
// 	}
// 	values := make([]*ast.MappingValueNode, 0, len(checksumNode.Values))
// 	for _, mvNode := range checksumNode.Values {
// 		switch mvNode.Key.String() {
// 		case "file_format", "pattern":
// 			continue
// 		}
// 		values = append(values, mvNode)
// 	}
// 	if len(checksumNode.Values) == len(values) {
// 		return false, nil
// 	}
// 	checksumNode.Values = values
// 	return true, nil
// }

func removeFromChecksumParent(parentNode *ast.MappingNode) (bool, error) {
	mvNode := getMappingValueNode(parentNode, "checksum")
	if mvNode == nil {
		return false, nil
	}

	switch t := mvNode.Value.(type) {
	case *ast.MappingNode:
		values := make([]*ast.MappingValueNode, 0, len(t.Values))
		for _, mvNode := range t.Values {
			switch mvNode.Key.String() {
			case "file_format", "pattern":
				continue
			}
			values = append(values, mvNode)
		}
		if len(t.Values) == len(values) {
			return false, nil
		}
		t.Values = values
		return true, nil
	case *ast.MappingValueNode:
		return false, nil
	}
	return false, nil
}

func convChecksumParserOfVersionOverrides(pkgInfoNode *ast.MappingNode, pkgInfo *registry.PackageInfo) (bool, error) {
	if pkgInfo.VersionOverrides == nil {
		return false, nil
	}

	mvNode := getMappingValueNode(pkgInfoNode, "version_overrides")
	if mvNode == nil {
		return false, nil
	}

	seq, ok := mvNode.Value.(*ast.SequenceNode)
	if !ok {
		return false, errors.New("version_overrides must be *ast.SequenceNode")
	}
	changed := false
	for _, val := range seq.Values {
		val := val
		m, ok := val.(*ast.MappingNode)
		if !ok {
			continue
		}
		if f, err := removeFromChecksumParent(m); err != nil {
			return false, err
		} else if f {
			changed = true
		}
	}
	return changed, nil
}

type Visitor struct {
	changed bool
}

func (visitor *Visitor) Visit(node ast.Node) ast.Visitor {
	a, ok := node.(*ast.MappingValueNode)
	if !ok {
		return visitor
	}
	t, ok := a.Value.(*ast.MappingNode)
	if !ok {
		return visitor
	}
	values := make([]*ast.MappingValueNode, 0, len(t.Values))
	for _, mvNode := range t.Values {
		switch mvNode.Key.String() {
		case "file_format", "pattern":
			continue
		}
		values = append(values, mvNode)
	}
	if len(t.Values) == len(values) {
		return visitor
	}
	t.Values = values
	visitor.changed = true
	return visitor
}
