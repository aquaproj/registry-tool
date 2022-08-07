package patchchecksum

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github"
	goccyYAML "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func PatchChecksum(ctx context.Context, logE *logrus.Entry, configFilePath string) error {
	cfg := &registry.Config{}
	b, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("open a configuration file %s: %w", configFilePath, err)
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("parse a configuration file %s as YAML: %w", configFilePath, err)
	}

	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	ghClient := github.New(ctx)
	size := len(cfg.PackageInfos)
	pkgsAST, err := getPackagesAST(file)
	if err != nil {
		return err
	}

	idx := 0
	for i := 0; i < size; i++ {
		pkgInfo := cfg.PackageInfos[i]
		node, j := findFirstMappingNode(pkgsAST, idx)
		if j == -1 {
			return nil
		}
		if err := patchChecksumOfPkg(ctx, logE, ghClient, node, pkgInfo); err != nil {
			logE.WithFields(logrus.Fields{
				"pkg_name": pkgInfo.GetName(),
			}).WithError(err).Error("patch a checksum config")
		}
	}
	if err := os.WriteFile(configFilePath, []byte(file.String()+"\n"), 0o644); err != nil { //nolint:gosec,gomnd
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}

func findFirstMappingNode(seq *ast.SequenceNode, idx int) (*ast.MappingNode, int) {
	s := len(seq.Values)
	for i := idx; i < s; i++ {
		value := seq.Values[i]
		m, ok := value.(*ast.MappingNode)
		if !ok {
			continue
		}
		return m, i
	}
	return nil, -1
}

func getPackagesAST(file *ast.File) (*ast.SequenceNode, error) { //nolint:cyclop
	for _, doc := range file.Docs {
		var values []*ast.MappingValueNode
		switch body := doc.Body.(type) {
		case *ast.MappingNode:
			values = body.Values
		case *ast.MappingValueNode:
			values = append(values, body)
		default:
			continue
		}
		for _, mapValue := range values {
			if mapValue.Key.String() != "packages" {
				continue
			}
			switch mapValue.Value.Type() {
			case ast.NullType:
				return nil, nil //nolint:nilnil
			case ast.SequenceType:
				seq, ok := mapValue.Value.(*ast.SequenceNode)
				if ok {
					return seq, nil
				}
				return nil, errors.New("packages must be *ast.SequenceNode")
			default:
				return nil, errors.New("packages must be null or array")
			}
		}
	}
	return nil, nil //nolint:nilnil
}

func patchChecksumOfPkg(ctx context.Context, logE *logrus.Entry, ghClient *github.RepositoriesService, node *ast.MappingNode, pkgInfo *registry.PackageInfo) error {
	if pkgInfo.Type != "github_release" {
		return nil
	}
	if pkgInfo.Checksum != nil {
		return nil
	}
	release, _, err := ghClient.GetLatestRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName)
	if err != nil {
		return fmt.Errorf("get the latest release: %w", err)
	}
	assets := listReleaseAssets(ctx, logE, ghClient, pkgInfo, release.GetID())
	if strings.Contains(strings.ToLower(pkgInfo.GetName()), "checksum") {
		return nil
	}
	for _, asset := range assets {
		assetName := asset.GetName()
		if !strings.Contains(strings.ToLower(assetName), "checksum") {
			continue
		}
		fileName := strings.ReplaceAll(assetName, release.GetTagName(), "{{.Version}}")
		fileName = strings.ReplaceAll(fileName, strings.TrimPrefix(release.GetTagName(), "v"), "{{trimV .Version}}")
		n, err := goccyYAML.ValueToNode(&registry.PackageInfo{
			Type: "github_release", // I don't know the reason, but without this attribute type makes empty. `type: ""`
			Checksum: &registry.Checksum{
				Type:       "github_release",
				Path:       fileName,
				FileFormat: "regexp",
				Pattern: &registry.ChecksumPattern{
					Checksum: "^(.{64})",
					File:     "^.{64}\\s+(\\S*)$",
				},
			},
		})
		if err != nil {
			return fmt.Errorf("create a YAML AST Node: %w", err)
		}
		if err := ast.Merge(node, n); err != nil {
			return fmt.Errorf("patch checksum: %w", err)
		}
		return nil
	}
	return nil
}

func listReleaseAssets(ctx context.Context, logE *logrus.Entry, ghClient *github.RepositoriesService, pkgInfo *registry.PackageInfo, releaseID int64) []*github.ReleaseAsset {
	opts := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var arr []*github.ReleaseAsset
	for i := 0; i < 10; i++ {
		assets, _, err := ghClient.ListReleaseAssets(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, releaseID, opts)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"repo_owner": pkgInfo.RepoOwner,
				"repo_name":  pkgInfo.RepoName,
			}).WithError(err).Warn("list release assets")
			return arr
		}
		arr = append(arr, assets...)
		if len(assets) < opts.PerPage {
			return arr
		}
		opts.Page++
	}
	return arr
}
