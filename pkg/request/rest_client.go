package request

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/systemquest/tavern-go/pkg/extension"
	"github.com/systemquest/tavern-go/pkg/schema"
	"github.com/systemquest/tavern-go/pkg/util"
)

// RestClient handles REST API requests
type RestClient struct {
	httpClient  *http.Client
	config      *Config
	RequestVars map[string]interface{} // Stores request arguments for access in response validation
	// persistentCookies stores cookies that have Expires or Max-Age set (persist across browser restarts)
	persistentCookies map[string][]*http.Cookie
}

// Config holds request configuration
type Config struct {
	Variables         map[string]interface{}
	Timeout           time.Duration
	HTTPClient        *http.Client              // Optional: shared HTTP client for session persistence
	PersistentCookies map[string][]*http.Cookie // Optional: shared map for tracking persistent cookies across stages
}

// NewRestClient creates a new REST API client
func NewRestClient(config *Config) *RestClient {
	if config == nil {
		config = &Config{
			Variables: make(map[string]interface{}),
			Timeout:   30 * time.Second,
		}
	}

	var client *http.Client
	if config.HTTPClient != nil {
		// Use provided client (for session persistence)
		client = config.HTTPClient
	} else {
		// Create new client with cookie jar
		jar, _ := cookiejar.New(nil)
		client = &http.Client{
			Timeout: config.Timeout,
			Jar:     jar, // Enable cookie jar for session persistence
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects automatically
				return http.ErrUseLastResponse
			},
		}
	}

	// Use shared persistent cookies map if provided, otherwise create new one
	persistentCookies := config.PersistentCookies
	if persistentCookies == nil {
		persistentCookies = make(map[string][]*http.Cookie)
	}

	return &RestClient{
		httpClient:        client,
		config:            config,
		persistentCookies: persistentCookies,
	}
}

// Execute executes an HTTP request
func (c *RestClient) Execute(spec schema.RequestSpec) (*http.Response, error) {
	// Process meta operations before request execution
	// Aligned with tavern-py commit 1dcffc6: support for meta operations like clear_session_cookies
	if err := c.processMeta(spec.Meta); err != nil {
		return nil, fmt.Errorf("failed to process meta operations: %w", err)
	}

	// Format the request spec with variables
	formattedSpec, err := c.formatRequestSpec(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to format request: %w", err)
	}

	// Build the request
	req, err := c.buildRequest(formattedSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Store request variables for access in response validation
	// Aligned with tavern-py commit 35e52d9: enables {tavern.request_vars.*}
	c.RequestVars = c.buildRequestVars(formattedSpec, req)

	// Configure HTTP client based on verify setting
	client := c.httpClient
	if formattedSpec.Verify != nil && !*formattedSpec.Verify {
		// Create a client with TLS verification disabled
		client = &http.Client{
			Timeout:       c.httpClient.Timeout,
			CheckRedirect: c.httpClient.CheckRedirect,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Track persistent cookies for clear_session_cookies support
	// Aligned with tavern-py commit 1dcffc6: preserve persistent cookies when clearing session cookies
	if c.httpClient.Jar != nil && resp != nil {
		c.trackPersistentCookies(req.URL, resp)
	}

	return resp, nil
}

// formatRequestSpec formats the request spec with variables
func (c *RestClient) formatRequestSpec(spec schema.RequestSpec) (schema.RequestSpec, error) {
	formatted := spec

	// Format URL
	if spec.URL != "" {
		formattedURL, err := util.FormatKeys(spec.URL, c.config.Variables)
		if err != nil {
			return formatted, err
		}
		formatted.URL = formattedURL.(string)
	}

	// Format headers
	if spec.Headers != nil {
		formattedHeaders := make(map[string]string)
		for k, v := range spec.Headers {
			formattedVal, err := util.FormatKeys(v, c.config.Variables)
			if err != nil {
				return formatted, err
			}
			formattedHeaders[k] = formattedVal.(string)
		}
		formatted.Headers = formattedHeaders
	}

	// Format params
	if spec.Params != nil {
		formattedParams := make(map[string]string)
		for k, v := range spec.Params {
			formattedVal, err := util.FormatKeys(v, c.config.Variables)
			if err != nil {
				return formatted, err
			}
			formattedParams[k] = formattedVal.(string)
		}
		formatted.Params = formattedParams
	}

	// Format JSON body
	if spec.JSON != nil {
		formattedJSON, err := util.FormatKeys(spec.JSON, c.config.Variables)
		if err != nil {
			return formatted, err
		}
		formatted.JSON = formattedJSON
	}

	// Check for $ext in JSON
	if formatted.JSON != nil {
		if jsonMap, ok := formatted.JSON.(map[string]interface{}); ok {
			if extSpec, ok := jsonMap["$ext"]; ok {
				generated, err := c.generateFromExt(extSpec)
				if err != nil {
					return formatted, err
				}
				formatted.JSON = generated
			}
		}
	}

	return formatted, nil
}

// generateFromExt generates data using an extension function
func (c *RestClient) generateFromExt(extSpec interface{}) (interface{}, error) {
	extMap, ok := extSpec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("$ext must be a map")
	}

	functionName, ok := extMap["function"].(string)
	if !ok {
		return nil, fmt.Errorf("$ext.function must be a string")
	}

	generator, err := extension.GetGenerator(functionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator: %w", err)
	}

	return generator(), nil
}

// buildRequest builds an HTTP request
func (c *RestClient) buildRequest(spec schema.RequestSpec) (*http.Request, error) {
	// Default method to GET
	method := spec.Method
	if method == "" {
		method = "GET"
	}

	// Build URL with query parameters
	requestURL := spec.URL
	if len(spec.Params) > 0 {
		parsedURL, err := url.Parse(spec.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		query := parsedURL.Query()
		for k, v := range spec.Params {
			query.Add(k, v)
		}
		parsedURL.RawQuery = query.Encode()
		requestURL = parsedURL.String()
	}

	// Build request body
	var body io.Reader
	var contentType string

	// Handle file uploads with multipart/form-data
	if len(spec.Files) > 0 {
		// Use multipart writer for file uploads
		bodyBuf := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuf)

		// Add each file to the multipart form
		for fieldName, filePath := range spec.Files {
			// Open the file
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			defer file.Close()

			// Get file name from path
			fileName := filepath.Base(filePath)

			// Create form file
			part, err := writer.CreateFormFile(fieldName, fileName)
			if err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to create form file: %w", err)
			}

			// Copy file content
			if _, err := io.Copy(part, file); err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to copy file content: %w", err)
			}
		}

		// Close multipart writer
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		body = bodyBuf
		contentType = writer.FormDataContentType()
	} else if spec.JSON != nil {
		jsonData, err := json.Marshal(spec.JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		body = bytes.NewReader(jsonData)
		contentType = "application/json"
	} else if spec.Data != nil {
		// Handle form data
		switch v := spec.Data.(type) {
		case string:
			body = strings.NewReader(v)
			contentType = "application/x-www-form-urlencoded"
		case map[string]interface{}:
			values := url.Values{}
			for key, val := range v {
				values.Add(key, fmt.Sprintf("%v", val))
			}
			body = strings.NewReader(values.Encode())
			contentType = "application/x-www-form-urlencoded"
		default:
			return nil, fmt.Errorf("unsupported data type: %T", spec.Data)
		}
	}

	// These verbs CAN send a body but the body SHOULD be ignored according to the HTTP specs.
	// While technically allowed, it's semantically incorrect and many servers/proxies may reject or ignore the body.
	// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
	// Aligned with tavern-py commits: 8d4db83 (warning logic), da8ed22 (documentation)
	if (method == "GET" || method == "HEAD" || method == "OPTIONS") && body != nil {
		logrus.Warnf("You are trying to send a body with HTTP %s which has no semantic use for it", method)
	}

	// Create request
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type if we have a body
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Set headers
	if spec.Headers != nil {
		for k, v := range spec.Headers {
			req.Header.Set(k, v)
		}
	}

	// Override content-type if explicitly set
	if spec.Headers != nil {
		for k, v := range spec.Headers {
			if strings.ToLower(k) == "content-type" {
				req.Header.Set("Content-Type", v)
				break
			}
		}
	}

	// Set authentication
	if spec.Auth != nil {
		if err := c.setAuth(req, spec.Auth); err != nil {
			return nil, err
		}
	}

	// Set cookies
	if spec.Cookies != nil {
		for k, v := range spec.Cookies {
			req.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	}

	return req, nil
}

// setAuth sets authentication on the request
func (c *RestClient) setAuth(req *http.Request, auth *schema.AuthSpec) error {
	switch strings.ToLower(auth.Type) {
	case "basic":
		if auth.Username == "" || auth.Password == "" {
			return fmt.Errorf("basic auth requires username and password")
		}
		req.SetBasicAuth(auth.Username, auth.Password)

	case "bearer":
		if auth.Token == "" {
			return fmt.Errorf("bearer auth requires token")
		}
		req.Header.Set("Authorization", "Bearer "+auth.Token)

	default:
		if auth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		} else if auth.Username != "" && auth.Password != "" {
			req.SetBasicAuth(auth.Username, auth.Password)
		}
	}

	return nil
}

// buildRequestVars builds a map of request variables for access in response validation
// Aligned with tavern-py commit 35e52d9: provides tavern.request_vars
func (c *RestClient) buildRequestVars(spec schema.RequestSpec, req *http.Request) map[string]interface{} {
	requestVars := make(map[string]interface{})

	// Store method from actual request
	requestVars["method"] = req.Method

	// Store URL
	requestVars["url"] = spec.URL

	// Store headers from actual request
	if len(req.Header) > 0 {
		headers := make(map[string]interface{})
		for key, values := range req.Header {
			if len(values) == 1 {
				headers[key] = values[0]
			} else {
				headers[key] = values
			}
		}
		requestVars["headers"] = headers
	}

	// Store params from URL query string
	if len(req.URL.Query()) > 0 {
		params := make(map[string]interface{})
		for key, values := range req.URL.Query() {
			if len(values) == 1 {
				params[key] = values[0]
			} else {
				params[key] = values
			}
		}
		requestVars["params"] = params
	}

	// Store JSON body
	if spec.JSON != nil {
		requestVars["json"] = spec.JSON
	}

	// Store data
	if spec.Data != nil {
		requestVars["data"] = spec.Data
	}

	return requestVars
}

// processMeta handles meta operations from the request spec
// Aligned with tavern-py commit 1dcffc6: support for clear_session_cookies
func (c *RestClient) processMeta(meta []string) error {
	if len(meta) == 0 {
		return nil
	}

	for _, operation := range meta {
		switch operation {
		case "clear_session_cookies":
			// Clear session cookies to simulate browser close
			// This is useful for testing session vs persistent cookies
			if err := c.clearSessionCookies(); err != nil {
				return fmt.Errorf("failed to clear session cookies: %w", err)
			}
			logrus.Debug("Cleared session cookies")
		default:
			logrus.Warnf("Unknown meta operation: %s", operation)
		}
	}

	return nil
}

// clearSessionCookies removes all session cookies from the HTTP client's cookie jar
// Session cookies are cookies without an explicit Expires or Max-Age attribute
// Persistent cookies (with Expires or Max-Age) are preserved
// Aligned with tavern-py commit 1dcffc6: session.cookies.clear_session_cookies()
func (c *RestClient) clearSessionCookies() error {
	if c.httpClient == nil || c.httpClient.Jar == nil {
		return nil // No cookie jar, nothing to clear
	}

	// Create a new jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	// Restore persistent cookies to the new jar
	// Persistent cookies are stored by base URL (scheme://host)
	for baseURLStr, cookies := range c.persistentCookies {
		// For each cookie, set it with the proper URL based on its Path attribute
		for _, cookie := range cookies {
			// Build URL: baseURL + cookie.Path
			cookieURL := baseURLStr
			if cookie.Path != "" && cookie.Path != "/" {
				cookieURL += cookie.Path
			} else {
				cookieURL += "/"
			}

			u, err := url.Parse(cookieURL)
			if err != nil {
				continue // Skip invalid URLs
			}

			// SetCookies expects a slice
			jar.SetCookies(u, []*http.Cookie{cookie})
		}
	}

	c.httpClient.Jar = jar
	return nil
}

// trackPersistentCookies tracks cookies with Expires or Max-Age attributes
// These cookies persist across browser restarts and should be preserved when clearing session cookies
func (c *RestClient) trackPersistentCookies(reqURL *url.URL, resp *http.Response) {
	if reqURL == nil || resp == nil {
		return
	}

	// Parse cookies from the response headers (not from jar)
	// The jar doesn't preserve Expires information when returning cookies
	cookies := resp.Cookies()

	// Filter persistent cookies (those with Expires or Max-Age set)
	var persistentCookies []*http.Cookie
	for _, cookie := range cookies {
		// A cookie is persistent if it has an Expires time set and it's not zero
		// or if MaxAge > 0
		if !cookie.Expires.IsZero() || cookie.MaxAge > 0 {
			persistentCookies = append(persistentCookies, cookie)
		}
	}

	// Store persistent cookies using base URL (scheme://host) as key
	// This allows cookies to be restored for any path on the same host
	if len(persistentCookies) > 0 {
		baseURL := fmt.Sprintf("%s://%s", reqURL.Scheme, reqURL.Host)
		// Merge with existing persistent cookies for this base URL
		existing := c.persistentCookies[baseURL]
		// Create a map to deduplicate by cookie name
		cookieMap := make(map[string]*http.Cookie)
		for _, c := range existing {
			cookieMap[c.Name] = c
		}
		for _, c := range persistentCookies {
			cookieMap[c.Name] = c
		}
		// Convert back to slice
		merged := make([]*http.Cookie, 0, len(cookieMap))
		for _, c := range cookieMap {
			merged = append(merged, c)
		}
		c.persistentCookies[baseURL] = merged
	}
}
