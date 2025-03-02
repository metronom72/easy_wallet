package main

import (
	"context"
	"encoding/json"
	"github.com/metronom72/crt_mmc/wallet_issue/internal/dynamo"
	"github.com/metronom72/crt_mmc/wallet_issue/internal/wallet"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type Response struct {
	PublicKey string `json:"public_key,omitempty"`
	Error     string `json:"error,omitempty"`
}

func processRequest(req Request) (Response, int) {
	log.Println("[INFO] Processing request...", "ID:", req.ID)

	if req.ID == "" || req.Password == "" {
		log.Println("[ERROR] Missing ID or Password")
		return Response{Error: "Missing ID or Password"}, http.StatusBadRequest
	}

	privateKey, publicKey, err := wallet.GenerateWallet()
	if err != nil {
		log.Printf("[ERROR] Wallet generation failed: %v", err)
		return Response{Error: "Failed to generate wallet"}, http.StatusInternalServerError
	}
	log.Println("[SUCCESS] Wallet generated")

	walletAddress, err := dynamo.StoreWallet(req.ID, req.Password, privateKey, publicKey)
	if err != nil {
		log.Printf("[ERROR] Failed to store wallet: %v", err)
		return Response{Error: "Failed to store wallet"}, http.StatusInternalServerError
	}
	log.Println("[SUCCESS] Wallet stored successfully")

	return Response{PublicKey: walletAddress}, http.StatusOK
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("[INFO] Lambda function invoked")

	log.Println("[INFO] Request Headers:")
	for key, value := range request.Headers {
		log.Printf("[INFO] %s: %s", key, value)
	}

	var req Request
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("[ERROR] Failed to parse request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error":"Invalid request format"}`,
		}, nil
	}

	resp, statusCode := processRequest(req)

	body, _ := json.Marshal(resp)
	log.Printf("[INFO] Response Status: %d", statusCode)
	log.Printf("[INFO] Response Body: %s", string(body))

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func localHTTPHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Local HTTP request received")
	log.Printf("[INFO] HTTP Method: %s, URL: %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request: %v", err)
		http.Error(w, `{"error": "Invalid request format"}`, http.StatusBadRequest)
		return
	}

	resp, statusCode := processRequest(req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
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

		http.HandleFunc("/wallet", localHTTPHandler)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("[ERROR] Failed to start server: %v", err)
		}
	}
}
