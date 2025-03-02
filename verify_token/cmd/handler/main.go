package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"verify_token/internal/auth"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Token string `json:"token,omitempty"`
}

type Response struct {
	Verified bool   `json:"verified,omitempty"`
	Error    string `json:"error,omitempty"`
}

func extractTokenFromHeaders(headers map[string]string) string {
	authHeader, exists := headers["Authorization"]
	if !exists {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	return ""
}

func processRequest(token string, methodArn string) (events.APIGatewayCustomAuthorizerResponse, int) {
	log.Println("[INFO] Processing request...")

	if token == "" {
		log.Println("[ERROR] Missing token")
		return generatePolicy("unauthorized", "Deny", methodArn), http.StatusUnauthorized
	}

	if _, err := auth.Verify(token); err != nil {
		log.Println("[ERROR] Token verification failed:", err)
		return generatePolicy("unauthorized", "Deny", methodArn), http.StatusUnauthorized
	}

	return generatePolicy("authorized-user", "Allow", methodArn), http.StatusOK
}

func generatePolicy(principalID, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		},
	}
}

func extractToken(token string) string {
	token = strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	}
	return token
}

func lambdaHandler(ctx context.Context, request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	log.Println("[INFO] Lambda Authorizer Invoked")
	log.Printf("[DEBUG] Context: %+v", ctx)

	requestJSON, _ := json.MarshalIndent(request, "", "  ")
	log.Printf("[DEBUG] Incoming Request: %s", string(requestJSON))

	token := extractToken(request.AuthorizationToken)
	log.Printf("[DEBUG] Extracted Token: %s", token)

	resp, statusCode := processRequest(token, request.MethodArn)

	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("[INFO] Response Status: %d", statusCode)
	log.Printf("[INFO] Response Body: %s", string(respJSON))

	return resp, nil
}

func localHTTPHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Local HTTP request received")
	log.Printf("[INFO] HTTP Method: %s, URL: %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	requestJSON, _ := json.MarshalIndent(r, "", "  ")
	log.Printf("[DEBUG] Incoming HTTP Request: %s", string(requestJSON))

	token := extractToken(r.Header.Get("Authorization"))
	if token == "" {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Failed to parse request: %v", err)
			http.Error(w, `{"error": "Invalid request format"}`, http.StatusBadRequest)
			return
		}
		token = extractToken(req.Token)
	}

	log.Printf("[DEBUG] Extracted Token: %s", token)

	resp, statusCode := processRequest(token, "*")

	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("[INFO] Response Status: %d", statusCode)
	log.Printf("[INFO] Response Body: %s", string(respJSON))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(respJSON)
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		log.Println("[INFO] Starting AWS Lambda Authorizer function")
		lambda.Start(lambdaHandler)
	} else {
		port := "3000"
		log.Printf("[INFO] Starting local server on port %s...", port)

		http.HandleFunc("/auth", localHTTPHandler)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("[ERROR] Failed to start server: %v", err)
		}
	}
}
