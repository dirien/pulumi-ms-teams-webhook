package aws

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Program(name string, ctx context.Context, args WebHookArgs) (auto.Stack, error) {
	projectName := "ms-teams-webhook"
	stackName := name

	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, WebHook(args))
	if err != nil {
		return s, err
	}

	w := s.Workspace()

	err = w.InstallPlugin(ctx, "aws", "v5.29.1")
	if err != nil {
		return s, fmt.Errorf("error installing AWS resource plugin: %v", err)
	}

	s.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: args.Region})

	return s, nil

}
