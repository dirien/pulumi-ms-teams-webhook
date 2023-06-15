package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/atc0005/go-teams-notify/v2/messagecard"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"os"
)

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Print(request.HTTPMethod)
	switch request.HTTPMethod {
	case "GET":
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Hello, World!",
		}, nil
	case "POST":
		mstClient := goteamsnotify.NewTeamsClient()

		webhookUrl := os.Getenv("MSTEAMS_WEBHOOK_URL")

		decodedBytes, err := base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			fmt.Println("Error decoding base64 string:", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf("Error decoding base64 string: %v", err),
			}, nil
		}
		var data interface{}
		err = json.Unmarshal(decodedBytes, &data)
		if err != nil {
			fmt.Println("Error unmarshalling JSON string:", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf("Error unmarshalling JSON string: %v", err),
			}, nil
		}
		prettyJSON, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			fmt.Println("Error pretty printing JSON string:", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf("Error pretty printing JSON string: %v", err),
			}, nil
		}

		// Setup message card.
		msgCard := messagecard.NewMessageCard()
		msgCard.Title = "Pulumi Webhook"
		msgCard.Text = fmt.Sprintf("Payload: %s", string(prettyJSON))
		msgCard.ThemeColor = "#DF813D"

		// Send the message with default timeout/retry settings.
		if err := mstClient.Send(webhookUrl, msgCard); err != nil {
			log.Printf("failed to send message: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf("failed to send message: %v", err),
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Message sent",
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 405,
		Body:       fmt.Sprintf("Method '%s' not allowed", request.HTTPMethod),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
