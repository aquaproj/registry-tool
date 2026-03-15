package main

import (
	"github.com/aquaproj/registry-tool/pkg/cli"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var version = ""

func main() {
	urfave.Main("argd", version, cli.Run)
}
