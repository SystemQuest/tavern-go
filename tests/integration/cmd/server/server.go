package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
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
	w.Header().Set("ATestHEader", "orange")
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
		"a_bool":     true, // Aligned with tavern-py commit 09a7376
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
			defer func() { _ = file.Close() }()

			// Save to /tmp
			path := fmt.Sprintf("/tmp/%s", fileHeader.Filename)
			dst, err := os.Create(path)
			if err != nil {
				http.Error(w, "Failed to create destination file", http.StatusInternalServerError)
				return
			}
			defer func() { _ = dst.Close() }()

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

// BoolTest endpoint - returns JSON with boolean fields for testing !anybool
// Aligned with tavern-py commit 3ff6b3c
func boolTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"active":    true,
		"enabled":   false,
		"verified":  true,
		"completed": false,
		"flags": map[string]interface{}{
			"feature_a": true,
			"feature_b": false,
		},
		"status_list": []interface{}{
			map[string]interface{}{"ok": true},
			map[string]interface{}{"ok": false},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Echo endpoint - returns the posted JSON body for testing type conversions
// StatusCodeReturn endpoint - returns empty JSON with status code from request body
// Aligned with tavern-py commit c21d737: Add integration tests
func statusCodeReturnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// Get status code from request body
	statusCode := 200 // default
	if code, ok := body["status_code"]; ok {
		switch v := code.(type) {
		case float64:
			statusCode = int(v)
		case int:
			statusCode = v
		}
	}

	// Return empty JSON with the requested status code
	response := map[string]interface{}{}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

// Echo endpoint - echoes back the JSON body received in the request
func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}

// ExpectDtype endpoint - validates that the received value has the expected data type
// Aligned with tavern-py commit 963bdf6
func expectDtypeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	value, hasValue := body["value"]
	dtype, hasDtype := body["dtype"]

	if !hasValue || !hasDtype {
		http.Error(w, "Missing value or dtype", http.StatusBadRequest)
		return
	}

	dtypeStr, ok := dtype.(string)
	if !ok {
		http.Error(w, "dtype must be a string", http.StatusBadRequest)
		return
	}

	// Check if the value type matches the expected dtype
	var actualType string
	switch v := value.(type) {
	case bool:
		actualType = "bool"
	case float64:
		// JSON numbers are always float64 in Go
		// Check if it's actually an int
		if v == float64(int64(v)) {
			actualType = "int"
		} else {
			actualType = "float"
		}
	case string:
		actualType = "str"
	default:
		actualType = "unknown"
	}

	if actualType != dtypeStr {
		http.Error(w, fmt.Sprintf("Type mismatch: expected %s, got %s", dtypeStr, actualType), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Pi endpoint - returns the mathematical constant pi
// Aligned with tavern-py commit 53690cf: Feature/approx numbers (#101)
func piHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"pi": math.Pi,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// FormData endpoint - handles form-encoded data
// Aligned with tavern-py commit 689fa39
func formDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read raw body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse form-encoded data (key=value format)
	bodyStr := string(bodyBytes)
	parts := strings.SplitN(bodyStr, "=", 2)

	response := make(map[string]interface{})
	if len(parts) == 2 {
		key := parts[0]
		value := parts[1]
		response[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// expectRawDataHandler - validates raw data sent in request body
// Aligned with tavern-py commit 4f2f276: Allow sending of raw data in the 'data' key
func expectRawDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	rawData := string(bodyBytes)
	var response map[string]interface{}
	var code int

	switch rawData {
	case "OK":
		response = map[string]interface{}{"status": "ok"}
		code = http.StatusOK
	case "DENIED":
		response = map[string]interface{}{"status": "denied"}
		code = http.StatusUnauthorized
	default:
		response = map[string]interface{}{"status": fmt.Sprintf("err: '%s'", rawData)}
		code = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
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
	http.HandleFunc("/bool_test", boolTestHandler)
	http.HandleFunc("/status_code_return", statusCodeReturnHandler) // Aligned with tavern-py commit c21d737
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/expect_dtype", expectDtypeHandler)
	http.HandleFunc("/pi", piHandler)
	http.HandleFunc("/form_data", formDataHandler)
	http.HandleFunc("/expect_raw_data", expectRawDataHandler) // Aligned with tavern-py commit 4f2f276

	// Start server
	port := 5000
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting integration test server on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
