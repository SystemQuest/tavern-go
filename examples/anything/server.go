package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// User represents a user with dynamic fields
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserResponse represents the response for /api/user endpoint
type UserResponse struct {
	User User `json:"user"`
}

// ItemsResponse represents the response for /api/items endpoint
type ItemsResponse struct {
	Items []interface{} `json:"items"`
	Count int           `json:"count"`
}

// userHandler returns a user with dynamic UUID and timestamp
func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := User{
		ID:        uuid.New().String(), // Dynamic UUID
		Name:      "John Doe",          // Fixed name
		Email:     "john@example.com",  // Fixed email
		CreatedAt: time.Now(),          // Dynamic timestamp
	}

	response := UserResponse{User: user}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// itemsHandler returns a list with dynamic values
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// First item is dynamic (current timestamp), second is fixed, third is dynamic (random number)
	items := []interface{}{
		time.Now().Unix(),   // Dynamic timestamp
		"fixed-value",       // Fixed string
		uuid.New().String(), // Dynamic UUID
		map[string]interface{}{ // Fixed nested object
			"type": "nested",
			"data": "constant",
		},
	}

	response := ItemsResponse{
		Items: items,
		Count: len(items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// nestedHandler returns nested data with some dynamic fields
func nestedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   uuid.New().String(), // Dynamic
			"name": "Alice",             // Fixed
			"profile": map[string]interface{}{
				"age":       time.Now().Unix() % 100, // Dynamic
				"city":      "New York",              // Fixed
				"lastLogin": time.Now(),              // Dynamic
			},
		},
		"metadata": map[string]interface{}{
			"version":   "1.0.0",             // Fixed
			"requestId": uuid.New().String(), // Dynamic
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Register handlers
	http.HandleFunc("/api/user", userHandler)
	http.HandleFunc("/api/items", itemsHandler)
	http.HandleFunc("/api/nested", nestedHandler)

	// Start server
	port := ":5001"
	log.Printf("🚀 Starting !anything example server on http://localhost%s\n", port)
	log.Println("Available endpoints:")
	log.Println("  - GET /api/user    - Returns user with dynamic ID and timestamp")
	log.Println("  - GET /api/items   - Returns array with mixed dynamic/fixed values")
	log.Println("  - GET /api/nested  - Returns nested data with dynamic fields")
	log.Println("\nPress Ctrl+C to stop")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
