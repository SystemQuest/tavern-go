package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// JWT signing key - in production, use environment variable
	jwtSecret = "CGQgaG7GYvTcpaQZqosLy4"
	// JWT audience
	jwtAudience = "testserver"
	// Database file
	dbFile = "numbers.db"
)

var db *sql.DB

// LoginRequest represents login credentials
type LoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// LoginResponse contains the JWT token
type LoginResponse struct {
	Token string `json:"token"`
}

// NumberRequest for storing a number
type NumberRequest struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
}

// NumberResponse for retrieving a number
type NumberResponse struct {
	Number int `json:"number"`
}

// DoubleRequest for doubling a number
type DoubleRequest struct {
	Name string `json:"name"`
}

// ErrorResponse for error messages
type ErrorResponse struct {
	Error string `json:"error"`
}

// JWT Claims structure
type Claims struct {
	User string `json:"user"`
	jwt.RegisteredClaims
}

// Initialize database
func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}

	// Create numbers table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS numbers (
			name TEXT PRIMARY KEY,
			number INTEGER NOT NULL
		)
	`)
	return err
}

// Reset database - delete all entries
func resetDB() error {
	_, err := db.Exec("DELETE FROM numbers")
	return err
}

// Generate JWT token
func generateToken(username string) (string, error) {
	claims := Claims{
		User: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  jwt.ClaimStrings{jwtAudience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// Validate JWT token from Authorization header
func validateToken(authHeader string) (*Claims, error) {
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("invalid authorization format")
	}

	tokenString := parts[1]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// Middleware to check authentication
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		_, err := validateToken(authHeader)
		if err != nil {
			sendError(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Simple authentication - in production, use proper password hashing
	if req.User != "test-user" || req.Password != "correct-password" {
		sendError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateToken(req.User)
	if err != nil {
		sendError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

// Numbers handler - GET/POST
func numbersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getNumber(w, r)
	case http.MethodPost:
		storeNumber(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get a number by name
func getNumber(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		sendError(w, "name parameter required", http.StatusBadRequest)
		return
	}

	var number int
	err := db.QueryRow("SELECT number FROM numbers WHERE name = ?", name).Scan(&number)
	if err == sql.ErrNoRows {
		sendError(w, "number not found", http.StatusNotFound)
		return
	} else if err != nil {
		sendError(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NumberResponse{Number: number})
}

// Store a number
func storeNumber(w http.ResponseWriter, r *http.Request) {
	var req NumberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		sendError(w, "name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT OR REPLACE INTO numbers (name, number) VALUES (?, ?)", req.Name, req.Number)
	if err != nil {
		sendError(w, "failed to store number", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Double handler - doubles a stored number
func doubleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DoubleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		sendError(w, "name is required", http.StatusBadRequest)
		return
	}

	// Get current number
	var number int
	err := db.QueryRow("SELECT number FROM numbers WHERE name = ?", req.Name).Scan(&number)
	if err == sql.ErrNoRows {
		sendError(w, "number not found", http.StatusNotFound)
		return
	} else if err != nil {
		sendError(w, "database error", http.StatusInternalServerError)
		return
	}

	// Double it
	doubled := number * 2

	// Update in database
	_, err = db.Exec("UPDATE numbers SET number = ? WHERE name = ?", doubled, req.Name)
	if err != nil {
		sendError(w, "failed to update number", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NumberResponse{Number: doubled})
}

// Ping handler - health check endpoint
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// Reset handler - clear all numbers from database
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := resetDB(); err != nil {
		sendError(w, "failed to reset database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Send JSON error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	log.Println("Database initialized")

	// Register handlers
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/numbers", authMiddleware(numbersHandler))
	http.HandleFunc("/double", authMiddleware(doubleHandler))
	http.HandleFunc("/reset", resetHandler)

	// Start server
	port := ":5000"
	log.Printf("Server starting on http://localhost%s", port)
	log.Println("Endpoints:")
	log.Println("  GET  /ping               - Health check")
	log.Println("  POST /login              - Get JWT token")
	log.Println("  GET  /numbers?name=...   - Get number (requires auth)")
	log.Println("  POST /numbers            - Store number (requires auth)")
	log.Println("  POST /double             - Double number (requires auth)")
	log.Println("  POST /reset              - Reset database")
	log.Println("")
	log.Println("Try: curl -X POST http://localhost:5000/login -H 'Content-Type: application/json' -d '{\"user\":\"test-user\",\"password\":\"correct-password\"}'")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
