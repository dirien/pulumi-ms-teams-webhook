package cli

import (
	createInfra "github.com/dirien/pulumi-ms-teams-webhook/cmd/cli/create"
	deleteInfra "github.com/dirien/pulumi-ms-teams-webhook/cmd/cli/delete"
	"github.com/spf13/cobra"
)

func InitializeCLI() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:  "pulumi-ms-teams-webhook",
		Long: "A CLI for creating Microsoft Teams webhooks for Pulumi",
	}

	rootCommand.AddCommand(createInfra.Command())
	rootCommand.AddCommand(deleteInfra.Command())
	return rootCommand
}
