package schema

// TestSpec represents a complete test specification
type TestSpec struct {
	TestName string    `yaml:"test_name" json:"test_name"`
	Includes []Include `yaml:"includes,omitempty" json:"includes,omitempty"`
	Stages   []Stage   `yaml:"stages" json:"stages"`

	// Future: Protocol-specific configurations at test level
	// Following tavern-py's approach: if "mqtt" in test_spec, initialize MQTT client
	// MQTT map[string]interface{} `yaml:"mqtt,omitempty" json:"mqtt,omitempty"`
}

// Include represents an include block with variables
type Include struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// Stage represents a single test stage
// It uses a flexible structure to support multiple protocols
// Similar to tavern-py's approach of checking keys like "request", "mqtt_publish", etc.
type Stage struct {
	Name string `yaml:"name" json:"name"`

	// Delay controls (in seconds)
	DelayBefore *float64 `yaml:"delay_before,omitempty" json:"delay_before,omitempty"`
	DelayAfter  *float64 `yaml:"delay_after,omitempty" json:"delay_after,omitempty"`

	// REST/HTTP protocol fields
	Request  *RequestSpec  `yaml:"request,omitempty" json:"request,omitempty"`
	Response *ResponseSpec `yaml:"response,omitempty" json:"response,omitempty"`

	// Future: MQTT protocol fields
	// MQTTPublish  *MQTTPublishSpec  `yaml:"mqtt_publish,omitempty" json:"mqtt_publish,omitempty"`
	// MQTTResponse *MQTTResponseSpec `yaml:"mqtt_response,omitempty" json:"mqtt_response,omitempty"`

	// Future: Shell/CLI protocol fields
	// Command        *ShellCommandSpec  `yaml:"command,omitempty" json:"command,omitempty"`
	// CommandResponse *ShellResponseSpec `yaml:"command_response,omitempty" json:"command_response,omitempty"`
}

// RequestSpec represents an HTTP request specification
type RequestSpec struct {
	Method  string            `yaml:"method,omitempty" json:"method,omitempty"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	JSON    interface{}       `yaml:"json,omitempty" json:"json,omitempty"`
	Data    interface{}       `yaml:"data,omitempty" json:"data,omitempty"`
	Params  map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
	Auth    *AuthSpec         `yaml:"auth,omitempty" json:"auth,omitempty"`
	Files   map[string]string `yaml:"files,omitempty" json:"files,omitempty"`
	Cookies map[string]string `yaml:"cookies,omitempty" json:"cookies,omitempty"`
	Verify  *bool             `yaml:"verify,omitempty" json:"verify,omitempty"` // SSL certificate verification, defaults to true
}

// AuthSpec represents authentication configuration
type AuthSpec struct {
	Type     string `yaml:"type,omitempty" json:"type,omitempty"` // basic, bearer, etc.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Token    string `yaml:"token,omitempty" json:"token,omitempty"`
}

// ResponseSpec represents expected HTTP response
type ResponseSpec struct {
	StatusCode int                    `yaml:"status_code,omitempty" json:"status_code,omitempty"`
	Headers    map[string]interface{} `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body       interface{}            `yaml:"body,omitempty" json:"body,omitempty"`
	Cookies    []string               `yaml:"cookies,omitempty" json:"cookies,omitempty"` // Expected cookie names
	Save       interface{}            `yaml:"save,omitempty" json:"save,omitempty"`       // Can be SaveSpec or $ext map
}

// SaveSpec specifies what to save from the response
type SaveSpec struct {
	Body                map[string]string `yaml:"body,omitempty" json:"body,omitempty"`
	Headers             map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	RedirectQueryParams map[string]string `yaml:"redirect_query_params,omitempty" json:"redirect_query_params,omitempty"`
}

// ExtSpec represents an extension function specification
type ExtSpec struct {
	Function    string                 `yaml:"function" json:"function"`
	ExtraArgs   []interface{}          `yaml:"extra_args,omitempty" json:"extra_args,omitempty"`
	ExtraKwargs map[string]interface{} `yaml:"extra_kwargs,omitempty" json:"extra_kwargs,omitempty"`
}
