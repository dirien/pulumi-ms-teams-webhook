package create

import (
	"github.com/dirien/pulumi-ms-teams-webhook/cmd/cli/delete/aws"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "delete",
		Short: "delete commands",
		Long:  "Commands that delete the infrastructure",
	}

	command.AddCommand(aws.Command())

	return command
}
