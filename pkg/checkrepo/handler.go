package checkrepo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
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

func CheckRepo(ctx context.Context, afs afero.Fs, httpClient *http.Client, pkgName string) error {
	repoOwner, repoName, err := CheckRedirect(ctx, afs, httpClient, pkgName)
	if err != nil {
		return err
	}

	fmt.Printf("%s/%s\n", repoOwner, repoName)                                          //nolint:forbidigo
	return logerr.WithFields(errors.New("a repository was transferred"), logrus.Fields{ //nolint:wrapcheck
		"package_name": pkgName,
		"repo_owner":   repoOwner,
		"repo_name":    repoName,
	})
}

func CheckRedirect(ctx context.Context, afs afero.Fs, httpClient *http.Client, pkgName string) (string, string, error) { //nolint:cyclop
	registryPath := filepath.Join("pkgs", filepath.FromSlash(pkgName), "registry.yaml")
	f, err := afero.ReadFile(afs, registryPath)
	if err != nil {
		return "", "", fmt.Errorf("open a registry file: %w", err)
	}
	rgst := &Registry{}
	if err := yaml.Unmarshal(f, rgst); err != nil {
		return "", "", fmt.Errorf("decode a registry file as YAML: %w", err)
	}
	if len(rgst.Packages) != 1 {
		return "", "", errors.New("the number of packages must be one")
	}
	pkg := rgst.Packages[0]
	if pkg.RepoOwner == "" || pkg.RepoName == "" {
		return "", "", nil
	}

	resp, err := request(ctx, httpClient, pkg.RepoOwner, pkg.RepoName)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 { //nolint:gomnd,usestdlibvars
		return "", "", nil
	}
	if resp.StatusCode >= 500 { //nolint:gomnd,usestdlibvars
		return "", "", logerr.WithFields(errors.New("http status code >= 500"), logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}
	if resp.StatusCode >= 400 { //nolint:gomnd,usestdlibvars
		return "", "", logerr.WithFields(errors.New("http status code >= 400"), logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}

	location, err := resp.Location()
	if err != nil {
		return "", "", fmt.Errorf("get location header: %w", err)
	}
	return getRepoFromLocation(location.Path)
}

func request(ctx context.Context, httpClient *http.Client, repoOwner, repoName string) (*http.Response, error) {
	u := fmt.Sprintf("https://github.com/%s/%s", repoOwner, repoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send a http request: %w", err)
	}
	return resp, nil
}

func getRepoFromLocation(location string) (string, string, error) {
	if !strings.HasPrefix(location, "/") {
		return "", "", errors.New("location must start with /")
	}

	b := strings.Split(location, "/")
	if len(b) != 2 { //nolint:gomnd
		return "", "", errors.New("location path format must be <repo_owner>/<repo_name>")
	}
	return b[0], b[1], nil
}
