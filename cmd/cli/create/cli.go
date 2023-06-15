package create

import (
	"github.com/dirien/pulumi-ms-teams-webhook/cmd/cli/create/aws"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create commands",
		Long:  "Commands that creates the infrastructure",
	}

	command.AddCommand(aws.Command())

	return command
}
