package main

import (
	"log"

	genrg "github.com/aquaproj/registry-tool/internal/generate-registry"
)

func main() {
	if err := genrg.GenerateRegistry(); err != nil {
		log.Fatal(err)
	}
}
