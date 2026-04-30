// Package libc detects whether a registry.yaml contains packages whose
// variants distinguish musl and gnu libc (i.e. variants with `key: libc`).
package libc

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

const variantKeyLibc = "libc"

type variant struct {
	Key string `yaml:"key"`
}

type override struct {
	Variants []variant `yaml:"variants"`
}

type versionOverride struct {
	Overrides []override `yaml:"overrides"`
}

type pkg struct {
	Overrides        []override        `yaml:"overrides"`
	VersionOverrides []versionOverride `yaml:"version_overrides"`
}

type root struct {
	Packages []pkg `yaml:"packages"`
}

// HasVariant reports whether the registry.yaml at path contains any package
// with a variant whose key is "libc". It returns (false, nil) when the file
// does not exist.
func HasVariant(path string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	var r root
	if err := yaml.Unmarshal(b, &r); err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}
	for _, p := range r.Packages {
		if anyOverrideHasLibc(p.Overrides) {
			return true, nil
		}
		for _, vo := range p.VersionOverrides {
			if anyOverrideHasLibc(vo.Overrides) {
				return true, nil
			}
		}
	}
	return false, nil
}

func anyOverrideHasLibc(ovs []override) bool {
	for _, ov := range ovs {
		for _, v := range ov.Variants {
			if v.Key == variantKeyLibc {
				return true
			}
		}
	}
	return false
}
