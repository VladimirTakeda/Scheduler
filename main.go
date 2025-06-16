package main

import (
	"InterviewScheduler/internal"
	dynamodb2 "InterviewScheduler/storage/dynamodb"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	dynamoStorage := dynamodb2.NewDynamoStorage()

	var req events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &req); err == nil && req.Body != "" {
		log.Printf("Received from API Gateway: %s", req.Body)
		return internal.HandleTelegramRequest(req, dynamoStorage)
	}

	var payload struct {
		ReminderID string `json:"reminder_id"`
		UserID     string `json:"user_id"`
	}
	if err := json.Unmarshal(event, &payload); err == nil && payload.ReminderID != "" {
		log.Printf("Received Scheduler event: %+v", payload)
		return internal.HandleSchedulerEvent(dynamoStorage, payload.ReminderID, payload.UserID)
	}

	log.Printf("Unknown event format: %s", string(event))
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       "Bad Request",
	}, nil
}
