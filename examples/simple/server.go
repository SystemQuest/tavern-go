package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// DoubleRequest represents the request body for the /double endpoint
type DoubleRequest struct {
	Number interface{} `json:"number"`
}

// DoubleResponse represents the successful response
type DoubleResponse struct {
	Double int `json:"double"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// doubleHandler handles POST requests to /double endpoint
// It doubles the provided number and returns the result
func doubleHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req DoubleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "no number passed", http.StatusBadRequest)
		return
	}

	// Check if number field is present
	if req.Number == nil {
		sendError(w, "no number passed", http.StatusBadRequest)
		return
	}

	// Convert to integer (handle different types)
	num, err := convertToInt(req.Number)
	if err != nil {
		sendError(w, "a number was not passed", http.StatusBadRequest)
		return
	}

	// Calculate and return the double
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DoubleResponse{Double: num * 2})
}

// convertToInt converts various types to integer
func convertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		num, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid number format")
		}
		return num, nil
	default:
		return 0, fmt.Errorf("unsupported type")
	}
}

// sendError sends a JSON error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func main() {
	// Register handler
	http.HandleFunc("/double", doubleHandler)

	// Start server
	port := ":5000"
	log.Printf("Server starting on %s", port)
	log.Printf("Try: curl -X POST http://localhost:5000/double -H 'Content-Type: application/json' -d '{\"number\": 5}'")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
