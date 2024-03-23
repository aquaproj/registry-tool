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
	redirect, err := CheckRedirect(ctx, afs, httpClient, pkgName)
	if err != nil {
		return err
	}
	if redirect == nil {
		return nil
	}

	fmt.Printf("%s/%s\n", redirect.NewRepoOwner, redirect.NewRepoName)                  //nolint:forbidigo
	return logerr.WithFields(errors.New("a repository was transferred"), logrus.Fields{ //nolint:wrapcheck
		"package_name": pkgName,
		"repo_owner":   redirect.NewRepoOwner,
		"repo_name":    redirect.NewRepoName,
	})
}

type Redirect struct {
	RepoOwner      string
	RepoName       string
	NewRepoOwner   string
	NewRepoName    string
	NewPackageName string
}

func CheckRedirect(ctx context.Context, afs afero.Fs, httpClient *http.Client, pkgName string) (*Redirect, error) { //nolint:cyclop
	registryPath := filepath.Join("pkgs", filepath.FromSlash(pkgName), "registry.yaml")
	f, err := afero.ReadFile(afs, registryPath)
	if err != nil {
		return nil, fmt.Errorf("open a registry file: %w", err)
	}
	rgst := &Registry{}
	if err := yaml.Unmarshal(f, rgst); err != nil {
		return nil, fmt.Errorf("decode a registry file as YAML: %w", err)
	}
	if len(rgst.Packages) != 1 {
		return nil, errors.New("the number of packages must be one")
	}
	pkg := rgst.Packages[0]
	if pkg.RepoOwner == "" || pkg.RepoName == "" {
		return nil, nil //nolint:nilnil
	}

	resp, err := request(ctx, httpClient, pkg.RepoOwner, pkg.RepoName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 { //nolint:gomnd
		return nil, nil //nolint:nilnil
	}
	if resp.StatusCode >= 500 { //nolint:gomnd
		return nil, logerr.WithFields(errors.New("http status code >= 500"), logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}
	if resp.StatusCode >= 400 { //nolint:gomnd
		return nil, logerr.WithFields(errors.New("http status code >= 400"), logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}

	location, err := resp.Location()
	if err != nil {
		return nil, fmt.Errorf("get location header: %w", err)
	}
	repoOwner, repoName, err := getRepoFromLocation(location.Path)
	if err != nil {
		return nil, err
	}
	redirect := &Redirect{
		RepoOwner:      pkg.RepoOwner,
		RepoName:       pkg.RepoName,
		NewRepoOwner:   repoOwner,
		NewRepoName:    repoName,
		NewPackageName: pkgName,
	}
	repoFullName := fmt.Sprintf("%s/%s", pkg.RepoOwner, pkg.RepoName)
	newRepoFullName := fmt.Sprintf("%s/%s", repoOwner, repoName)
	if pkgName == repoFullName {
		redirect.NewPackageName = newRepoFullName
	} else if strings.HasPrefix(pkgName, repoFullName+"/") {
		redirect.NewPackageName = strings.Replace(pkgName, repoFullName, newRepoFullName, 1)
	}
	return redirect, nil
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

	// /<repo_owner>/<repo_name>
	b := strings.Split(location, "/")
	if len(b) != 3 { //nolint:gomnd
		return "", "", errors.New("location path format must be /<repo_owner>/<repo_name>")
	}
	return b[1], b[2], nil
}
