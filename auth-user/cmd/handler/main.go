package main

import (
	"context"
	"internal/models"
	"internal/services"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Provider services.ProviderEnum  `json:"provider" binding:"required,oneof=telegram"`
	Data     map[string]interface{} `json:"data"`
}

type Response struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
	Error error       `json:"error"`
}

func processRequest(ctx context.Context, req Request) (Response, int) {
	log.Println("[INFO] Processing request...")

	auth := services.Auth{
		Data:     req.Data,
		Provider: req.Provider,
	}

	userData, err := auth.Authorize(ctx)
	if err != nil {
		return Response{Error: err}, http.StatusBadRequest
	}

	token, err := auth.EncodeUser()
	if err != nil {
		return Response{Error: err}, http.StatusBadRequest
	}

	return Response{
		User:  *userData,
		Token: token,
		Error: nil,
	}, http.StatusOK
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		log.Println("[INFO] Starting AWS Lambda function")
		lambda.Start(lambdaHandler)
	} else {
		port := "3000"
		log.Printf("[INFO] Starting local server on port %s...", port)

		http.HandleFunc("/auth/user", localHTTPHandler)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("[ERROR] Failed to start server: %v", err)
		}
	}
}
