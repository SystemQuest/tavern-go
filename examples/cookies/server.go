package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// Session storage
type Session struct {
	Username string
}

var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

type ProtectedResponse struct {
	Message string `json:"message"`
	User    string `json:"user"`
	Data    string `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	if req.Username == "testuser" && req.Password == "testpass" {
		// Create session
		sessionID := uuid.New().String()
		mu.Lock()
		sessions[sessionID] = &Session{Username: req.Username}
		mu.Unlock()

		// Set cookies
		http.SetCookie(w, &http.Cookie{
			Name:  "session_id",
			Value: sessionID,
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "user_pref",
			Value: "theme_dark",
			Path:  "/",
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{
			Message:  "Login successful",
			Username: req.Username,
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid credentials"})
	}
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Check session
	mu.RLock()
	session, exists := sessions[cookie.Value]
	mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProtectedResponse{
		Message: "Access granted",
		User:    session.Username,
		Data:    "secret information",
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err == nil {
		// Delete session
		mu.Lock()
		delete(sessions, cookie.Value)
		mu.Unlock()
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LogoutResponse{Message: "Logged out"})
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/api/protected", protectedHandler)
	http.HandleFunc("/logout", logoutHandler)

	addr := ":5555"
	fmt.Printf("Starting cookie test server on http://localhost%s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /login - Login with username/password")
	fmt.Println("  GET  /api/protected - Access protected resource")
	fmt.Println("  POST /logout - Logout")

	log.Fatal(http.ListenAndServe(addr, nil))
}
