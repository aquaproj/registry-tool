package checkrepo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v3"
)

type Registry struct {
	Packages []*Package
}

type Package struct {
	RepoOwner string `yaml:"repo_owner"`
	RepoName  string `yaml:"repo_name"`
	Aliases   []*Alias
}

type Alias struct {
	Name string
}

func CheckRepo(ctx context.Context, fixOpt bool, pkgName string) error { //nolint:cyclop,funlen
	a := strings.Split(pkgName, "/")
	registryPath := filepath.Join(append(append([]string{"pkgs"}, a...), "registry.yaml")...)
	f, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("open a registry file: %w", err)
	}
	rgst := &Registry{}
	if err := yaml.Unmarshal(f, rgst); err != nil {
		return fmt.Errorf("decode a registry file as YAML: %w", err)
	}
	if len(rgst.Packages) != 1 {
		return errors.New("the number of packages must be one")
	}
	pkg := rgst.Packages[0]
	if pkg.RepoOwner == "" || pkg.RepoName == "" {
		return nil
	}
	u := fmt.Sprintf("https://github.com/%s/%s", pkg.RepoOwner, pkg.RepoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return fmt.Errorf("create a http request: %w", err)
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send a http request: %w", err)
	}
	defer resp.Body.Close()

	logrus.Info(resp.StatusCode)

	if resp.StatusCode < 300 { //nolint:gomnd,usestdlibvars
		return nil
	}
	if resp.StatusCode >= 500 { //nolint:gomnd,usestdlibvars
		return logerr.WithFields(errors.New("http status code >= 500"), logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}
	if resp.StatusCode >= 400 { //nolint:gomnd,usestdlibvars
		return logerr.WithFields(errors.New("http status code >= 400"), logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}

	location, err := resp.Location()
	if err != nil {
		return fmt.Errorf("get location header: %w", err)
	}
	if !strings.HasPrefix(location.Path, "/") {
		return errors.New("location must start with /")
	}

	b := strings.Split(location.Path[1:], "/")
	if len(b) != 2 { //nolint:gomnd
		return errors.New("location path format must be <repo_owner>/<repo_name>")
	}
	repoOwner := b[0]
	repoName := b[1]

	fmt.Printf("%s/%s\n", repoOwner, repoName)                                          //nolint:forbidigo
	return logerr.WithFields(errors.New("a repository was transferred"), logrus.Fields{ //nolint:wrapcheck
		"package_name": pkgName,
		"repo_owner":   repoOwner,
		"repo_name":    repoName,
	})
}
