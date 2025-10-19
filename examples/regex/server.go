package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/verify", verifyHandler)

	fmt.Println("Server starting on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	// Return HTML with a link containing a token
	html := `<div><a href="http://127.0.0.1:5000/verify?token=c9bb34ba-131b-11e8-b642-0ed5f89f718b">Verify Link</a></div>`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "c9bb34ba-131b-11e8-b642-0ed5f89f718b" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Valid token"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Invalid token"))
	}
}
