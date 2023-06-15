package aws

import (
	"context"
	"fmt"
	"github.com/dirien/pulumi-ms-teams-webhook/internal/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/spf13/cobra"
	"os"
)

var (
	region    string
	subnetIds []string
	name      string
	tailnet   string
	apiKey    string
	routes    []string
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "aws",
		Short: "Create the AWS infrastructure",
		Long:  `Create the AWS infrastructure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			program, err := aws.Program(name, ctx, aws.WebHookArgs{
				Region: region,
			})
			if err != nil {
				return err
			}

			stdoutStreamer := optdestroy.ProgressStreams(os.Stdout)
			_, err = program.Refresh(ctx)
			if err != nil {
				return fmt.Errorf("error refreshing stack: %v", err)
			}
			_, err = program.Destroy(ctx, stdoutStreamer)
			if err != nil {
				_, err = program.Destroy(ctx)
				if err != nil {
					return fmt.Errorf("failed to destroy stack: %v", err)
				}
				err = program.Workspace().RemoveStack(ctx, name)
				if err != nil {
					return fmt.Errorf("failed to remove stack: %v", err)
				}
				return fmt.Errorf("failed update: %v", err)
			}
			return nil
		},
	}

	command.Flags().StringVar(&region, "region", "", "The AWS Region to use.")
	command.Flags().StringVar(&name, "name", "", "Unique name to use for your bastion.")

	command.MarkFlagRequired("subnet-ids")

	return command
}
