package main

import (
	"github.com/dirien/pulumi-ms-teams-webhook/cmd/cli"
	"os"
)

func main() {
	rootCommand := cli.InitializeCLI()

	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
