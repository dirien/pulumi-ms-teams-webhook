package aws

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	apigateway "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"math/rand"
)

type WebHookArgs struct {
	Name              string
	Region            string
	MSTeamsWebhookURL string
	SharedSecret      string
}

func WebHook(args WebHookArgs) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		//mstClient := goteamsnotify.NewTeamsClient()

		// Set webhook url.
		//webhookUrl := "https://outlook.office.com/webhook/YOUR_WEBHOOK_URL_OF_TEAMS_CHANNEL"

		role, err := iam.NewRole(ctx, "task-exec-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
							"Version": "2012-10-17",
							"Statement": [{
								"Action": "sts:AssumeRole",
								"Principal": {
									"Service": "lambda.amazonaws.com"
								},
								"Effect": "Allow"
							}]
						}`),
			ManagedPolicyArns: pulumi.StringArray{
				iam.ManagedPolicyAWSLambdaBasicExecutionRole,
			},
		})
		if err != nil {
			return err
		}

		logPolicy, err := iam.NewRolePolicy(ctx, "lambda-log-policy", &iam.RolePolicyArgs{
			Role: role.Name,
			Policy: pulumi.String(`{
                "Version": "2012-10-17",
                "Statement": [{
                    "Effect": "Allow",
                    "Action": [
                        "logs:CreateLogGroup",
                        "logs:CreateLogStream",
                        "logs:PutLogEvents"
                    ],
                    "Resource": "arn:aws:logs:*:*:*"
                }]
            }`),
		})
		if err != nil {
			return err
		}

		f, err := lambda.NewFunction(ctx, "lambda", &lambda.FunctionArgs{
			Runtime: lambda.RuntimeGo1dx,
			Code:    pulumi.NewFileArchive("handler.zip"),
			Timeout: pulumi.Int(300),
			Handler: pulumi.String("handler"),
			Role:    role.Arn,
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"MSTEAMS_WEBHOOK_URL": pulumi.String(args.MSTeamsWebhookURL),
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{logPolicy}))

		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating lambda function: %v", err), nil)
			return err
		}

		restAPI, err := apigateway.NewRestApi(ctx, "api", &apigateway.RestApiArgs{
			Description: pulumi.String(args.Name),
			BinaryMediaTypes: pulumi.StringArray{
				pulumi.String("application/json"),
			},
			EndpointConfiguration: &apigateway.RestApiEndpointConfigurationArgs{
				Types: pulumi.String("REGIONAL"),
			},
			Policy: pulumi.String(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    },
    {
      "Action": "execute-api:Invoke",
      "Resource": "*",
      "Principal": "*",
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}`)})
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API: %v", err), nil)
			return err
		}
		account, err := aws.GetCallerIdentity(ctx)

		if err != nil {
			return err
		}
		permission, err := lambda.NewPermission(ctx, "APIPermission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  f.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*/*", args.Region, account.AccountId, restAPI.ID()),
		}, pulumi.DependsOn([]pulumi.Resource{restAPI}))
		if err != nil {
			return err
		}

		// Create the GET method on the root resource
		method, err := apigateway.NewMethod(ctx, "method", &apigateway.MethodArgs{
			RestApi:       restAPI.ID(),
			ResourceId:    restAPI.RootResourceId,
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("NONE"),
		})
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API method: %v", err), nil)
			return err
		}
		integration, err := apigateway.NewIntegration(ctx, "integration", &apigateway.IntegrationArgs{
			RestApi:               restAPI.ID(),
			ResourceId:            method.ResourceId,
			HttpMethod:            method.HttpMethod,
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   f.InvokeArn,
		})

		// Create the POST method on the root resource
		methodPost, err := apigateway.NewMethod(ctx, "method-post", &apigateway.MethodArgs{
			RestApi:       restAPI.ID(),
			ResourceId:    restAPI.RootResourceId,
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("NONE"),
		})
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API method: %v", err), nil)
			return err
		}
		integrationPost, err := apigateway.NewIntegration(ctx, "integration-post", &apigateway.IntegrationArgs{
			RestApi:               restAPI.ID(),
			ResourceId:            methodPost.ResourceId,
			HttpMethod:            methodPost.HttpMethod,
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   f.InvokeArn,
		})
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API integration: %v", err), nil)
			return err
		}

		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API integration: %v", err), nil)
			return err
		}
		deployment, err := apigateway.NewDeployment(ctx, "response", &apigateway.DeploymentArgs{
			RestApi:     restAPI.ID(),
			Description: pulumi.String("Deployed by Pulumi"),
			Triggers: pulumi.StringMap{
				"redeploy": pulumi.Sprintf("%d", rand.Int()),
			},
		}, pulumi.DependsOn([]pulumi.Resource{method, methodPost, integration, integrationPost, permission}))
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API deployment: %v", err), nil)
			return err
		}
		stage, err := apigateway.NewStage(ctx, "response", &apigateway.StageArgs{
			RestApi:    restAPI.ID(),
			Deployment: deployment.ID(),
			StageName:  pulumi.String("prod"),
		})
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error creating REST API stage: %v", err), nil)
			return err
		}

		//https://${apiGateway}.execute-api.${AWS::Region}.amazonaws.com/${apiGatewayStageName}
		ctx.Export("url", pulumi.All(restAPI.ID(), args.Region, stage.StageName).ApplyT(func(args []interface{}) (string, error) {
			return fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", args[0], args[1], args[2]), nil
		}))
		/*
			var getMethod = apigateway.MethodGET

			webhookAPI, err := apigateway.NewRestAPI(ctx, "api", &apigateway.RestAPIArgs{
				StageName: pulumi.String("prod"),
				Routes: []apigateway.RouteArgs{
					{
						Path:         "/",
						Method:       &getMethod,
						EventHandler: f,
						Target: apigateway.TargetArgs{
							HttpMethod: apigateway.MethodGET,
							Type:       apigateway.IntegrationTypeMock,
						},
					},
				},
			})
			if err != nil {
				return err
			}
			ctx.Export("webhookAPI", webhookAPI.Url)
		*/
		return nil
	}
}
