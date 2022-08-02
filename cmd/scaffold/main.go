package main

import (
	"context"
	"log"
	"os"

	"github.com/aquaproj/registry-tool/internal/scaffold"
)

func main() {
	ctx := context.Background()
	if err := scaffold.Scaffold(ctx, os.Args...); err != nil {
		log.Fatal(err)
	}
}
