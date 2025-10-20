package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

var (
	// Session store with secret key
	store *sessions.CookieStore

	// User database
	users = map[string]User{
		"mark": {
			Password:  "password",
			Regular:   "foo",
			Protected: "bar",
		},
	}
)

func init() {
	store = sessions.NewCookieStore([]byte("secret"))
	// Set MaxAge to 0 to make it a session cookie (no Expires/Max-Age)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0, // Session cookie - cleared when browser closes
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteDefaultMode,
	}
}

// User represents a user with credentials and data
type User struct {
	Password  string
	Regular   string
	Protected string
}

// Remember token serializer (similar to Python's itsdangerous)
type TokenSerializer struct {
	secretKey string
	salt      string
}

func newTokenSerializer(secretKey, salt string) *TokenSerializer {
	return &TokenSerializer{
		secretKey: secretKey,
		salt:      salt,
	}
}

// dumps creates a signed token with timestamp
func (ts *TokenSerializer) dumps(data string) string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	payload := fmt.Sprintf("%s.%s", data, timestamp)
	signature := ts.sign(payload)
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s.%s", payload, signature)))
}

// loads verifies and extracts data from token
func (ts *TokenSerializer) loads(token string, maxAge int64) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid token encoding")
	}

	parts := strings.Split(string(decoded), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	data := parts[0]
	timestamp := parts[1]
	signature := parts[2]

	// Verify signature
	payload := fmt.Sprintf("%s.%s", data, timestamp)
	expectedSig := ts.sign(payload)
	if signature != expectedSig {
		return "", fmt.Errorf("invalid signature")
	}

	// Check timestamp
	var ts64 int64
	fmt.Sscanf(timestamp, "%d", &ts64)
	if time.Now().Unix()-ts64 > maxAge {
		return "", fmt.Errorf("token expired")
	}

	return data, nil
}

// sign creates HMAC-SHA512 signature
func (ts *TokenSerializer) sign(payload string) string {
	key := []byte(ts.secretKey + ts.salt)
	h := hmac.New(sha512.New, key)
	h.Write([]byte(payload))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

var serializer = newTokenSerializer("secret", "cookie")

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/regular", regularHandler)
	http.HandleFunc("/protected", protectedHandler)

	log.Println("Server starting on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

// loginHandler handles POST /login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, exists := users[credentials.Username]
	if !exists || user.Password != credentials.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set session cookie (session cookie - no Expires)
	session, _ := store.Get(r, "session")
	session.Values["user"] = credentials.Username
	session.Save(r, w)

	// Set remember cookie (persistent cookie - with Expires)
	rememberToken := serializer.dumps(credentials.Username)
	http.SetCookie(w, &http.Cookie{
		Name:     "remember",
		Value:    rememberToken,
		Expires:  time.Now().Add(30 * 24 * time.Hour), // 30 days
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

// regularHandler handles GET /regular
// Accepts both session cookie and remember cookie
func regularHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var username string

	// Try session cookie first
	session, _ := store.Get(r, "session")
	if user, ok := session.Values["user"].(string); ok {
		username = user
	}

	// If no session, try remember cookie
	if username == "" {
		if cookie, err := r.Cookie("remember"); err == nil {
			if user, err := serializer.loads(cookie.Value, 3600); err == nil {
				username = user
			}
		}
	}

	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, exists := users[username]
	if !exists {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"regular": user.Regular,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// protectedHandler handles GET /protected
// Requires session cookie only (not remember cookie)
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["user"].(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, exists := users[username]
	if !exists {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"protected": user.Protected,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
