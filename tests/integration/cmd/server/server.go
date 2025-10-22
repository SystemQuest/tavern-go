package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Token endpoint - returns HTML with a link containing a token
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := `<div><a src="http://127.0.0.1:5000/verify?token=c9bb34ba-131b-11e8-b642-0ed5f89f718b">Link</a></div>`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// Headers endpoint - returns response with custom header for regex testing
func headersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("X-Integration-Value", "_HelloWorld1")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Verify endpoint - validates token from query parameter
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "c9bb34ba-131b-11e8-b642-0ed5f89f718b" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// FakeDictionary endpoint - returns nested JSON structure
func fakeDictionaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"top": map[string]interface{}{
			"Thing": "value",
			"nested": map[string]interface{}{
				"doubly": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		"an_integer": 123,
		"a_string":   "abc",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// FakeList endpoint - returns JSON array with mixed types
func fakeListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mixed types: strings, integers, and floats (aligned with tavern-py commit 59e615d)
	response := []interface{}{
		"a", "b", "c",
		1, 2, 3,
		-1.0, -2.0, -3.0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// NestedList endpoint - returns JSON with nested array containing mixed types
func nestedListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"top": []interface{}{
			"a",
			"b",
			map[string]interface{}{
				"key": "val",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// File upload endpoint - receives files and saves them temporarily
func fakeUploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Check if files were provided
	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Process each uploaded file
	for key := range r.MultipartForm.File {
		fileHeaders := r.MultipartForm.File[key]
		for _, fileHeader := range fileHeaders {
			// Open the uploaded file
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Failed to open uploaded file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Save to /tmp
			path := fmt.Sprintf("/tmp/%s", fileHeader.Filename)
			dst, err := os.Create(path)
			if err != nil {
				http.Error(w, "Failed to create destination file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy file content
			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Nested/again endpoint - returns simple OK status
func nestedAgainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status": "OK",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	// Register handlers
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/headers", headersHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/fake_dictionary", fakeDictionaryHandler)
	http.HandleFunc("/fake_list", fakeListHandler)
	http.HandleFunc("/nested_list", nestedListHandler)
	http.HandleFunc("/fake_upload_file", fakeUploadFileHandler)
	http.HandleFunc("/nested/again", nestedAgainHandler)

	// Start server
	port := 5000
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting integration test server on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
