package main

import (
	"encoding/json"
	"internal/db"
	"log"
	"net/http"
)

func localHTTPHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Local HTTP request received")
	log.Printf("[INFO] HTTP Method: %s, URL: %s", r.Method, r.URL.Path)

	addCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

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

	ctx := db.InjectDBToContext(r.Context())

	resp, statusCode := processRequest(ctx, req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
