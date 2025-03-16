package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"internal/db"
	"log"
	"net/http"
)

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("[INFO] Lambda function invoked")

	var req Request
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("[ERROR] Failed to parse request: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: `{"error":"Invalid request format"}`}, nil
	}

	ctx = db.InjectDBToContext(ctx)

	resp, statusCode := processRequest(ctx, req)

	body, _ := json.Marshal(resp)
	log.Printf("[INFO] Response Status: %d", statusCode)
	log.Printf("[INFO] Response Body: %s", string(body))

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}
