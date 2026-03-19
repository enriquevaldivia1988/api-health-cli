package main

import (
	"os"

	"github.com/enriquevaldivia1988/api-health-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
