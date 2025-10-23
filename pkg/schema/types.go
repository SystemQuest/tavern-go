package schema

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// TestSpec represents a complete test specification
type TestSpec struct {
	TestName string    `yaml:"test_name" json:"test_name"`
	Includes []Include `yaml:"includes,omitempty" json:"includes,omitempty"`
	Stages   []Stage   `yaml:"stages" json:"stages"`
	Strict   *Strict   `yaml:"strict,omitempty" json:"strict,omitempty"` // Response key matching strictness
	Xfail    string    `yaml:"_xfail,omitempty" json:"_xfail,omitempty"` // Expected failure mode: "verify" or "run"

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

	// Test control keywords (aligned with tavern-py commit cfdf901)
	Skip bool `yaml:"skip,omitempty" json:"skip,omitempty"` // Skip this stage
	Only bool `yaml:"only,omitempty" json:"only,omitempty"` // Run only this stage and stop

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
	Meta    []string          `yaml:"meta,omitempty" json:"meta,omitempty"`     // Meta operations like "clear_session_cookies"
}

// AuthSpec represents authentication configuration
type AuthSpec struct {
	Type     string `yaml:"type,omitempty" json:"type,omitempty"` // basic, bearer, etc.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Token    string `yaml:"token,omitempty" json:"token,omitempty"`
}

// StatusCode represents expected HTTP status code(s) - can be a single int or a list of ints
// Aligned with tavern-py commit af74465
type StatusCode struct {
	Single   int   // Single status code
	Multiple []int // Multiple acceptable status codes
}

// UnmarshalYAML implements custom unmarshaling for StatusCode
func (sc *StatusCode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as single int first
	var single int
	if err := unmarshal(&single); err == nil {
		sc.Single = single
		sc.Multiple = nil
		return nil
	}

	// Try to unmarshal as slice of ints
	var multiple []int
	if err := unmarshal(&multiple); err == nil {
		sc.Single = 0
		sc.Multiple = multiple
		return nil
	}

	return fmt.Errorf("status_code must be an integer or a list of integers")
}

// UnmarshalJSON implements custom unmarshaling for StatusCode
func (sc *StatusCode) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as single int first
	var single int
	if err := json.Unmarshal(data, &single); err == nil {
		sc.Single = single
		sc.Multiple = nil
		return nil
	}

	// Try to unmarshal as slice of ints
	var multiple []int
	if err := json.Unmarshal(data, &multiple); err == nil {
		sc.Single = 0
		sc.Multiple = multiple
		return nil
	}

	return fmt.Errorf("status_code must be an integer or a list of integers")
}

// MarshalYAML implements custom marshaling for StatusCode
func (sc *StatusCode) MarshalYAML() (interface{}, error) {
	if sc.Multiple != nil {
		return sc.Multiple, nil
	}
	return sc.Single, nil
}

// MarshalJSON implements custom marshaling for StatusCode
func (sc *StatusCode) MarshalJSON() ([]byte, error) {
	if sc.Multiple != nil {
		return json.Marshal(sc.Multiple)
	}
	return json.Marshal(sc.Single)
}

// Contains checks if the given status code matches the expected code(s)
func (sc *StatusCode) Contains(code int) bool {
	if sc.Multiple != nil {
		for _, c := range sc.Multiple {
			if c == code {
				return true
			}
		}
		return false
	}
	return sc.Single == code
}

// IsZero returns true if no status code is set
func (sc *StatusCode) IsZero() bool {
	return sc.Single == 0 && len(sc.Multiple) == 0
}

// String returns a string representation of the status code(s)
func (sc *StatusCode) String() string {
	if sc.Multiple != nil {
		codes := make([]string, len(sc.Multiple))
		for i, c := range sc.Multiple {
			codes[i] = strconv.Itoa(c)
		}
		return "[" + strings.Join(codes, ", ") + "]"
	}
	return strconv.Itoa(sc.Single)
}

// ResponseSpec represents expected HTTP response
type ResponseSpec struct {
	StatusCode *StatusCode            `yaml:"status_code,omitempty" json:"status_code,omitempty"`
	Headers    map[string]interface{} `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body       interface{}            `yaml:"body,omitempty" json:"body,omitempty"`
	Cookies    []string               `yaml:"cookies,omitempty" json:"cookies,omitempty"` // Expected cookie names
	Save       *SaveConfig            `yaml:"save,omitempty" json:"save,omitempty"`       // Union type: SaveSpec or ExtSpec
	Strict     *Strict                `yaml:"strict,omitempty" json:"strict,omitempty"`   // Response key matching strictness for this stage
}

// SaveSpec specifies what to save from the response
type SaveSpec struct {
	Body                map[string]interface{} `yaml:"body,omitempty" json:"body,omitempty"`
	Headers             map[string]string      `yaml:"headers,omitempty" json:"headers,omitempty"`
	RedirectQueryParams map[string]string      `yaml:"redirect_query_params,omitempty" json:"redirect_query_params,omitempty"`
}

// ExtSpec represents an extension function specification
type ExtSpec struct {
	Function    string                 `yaml:"function" json:"function"`
	ExtraArgs   []interface{}          `yaml:"extra_args,omitempty" json:"extra_args,omitempty"`
	ExtraKwargs map[string]interface{} `yaml:"extra_kwargs,omitempty" json:"extra_kwargs,omitempty"`
}

// Strict represents the strictness configuration for response key matching
// It can be:
// - nil: use default/legacy behavior (top-level keys ignored, nested keys strict)
// - bool: true = strict at all levels, false = lenient at all levels
// - []string: strict only for specified parts (e.g., ["body", "headers"])
type Strict struct {
	IsSet    bool        // Whether strict was explicitly set
	AsBool   bool        // Used when strict is a boolean
	AsList   []string    // Used when strict is a list of response parts
	IsList   bool        // Whether strict is a list
	IsLegacy bool        // Whether to use legacy behavior (nil/unset)
	RawValue interface{} // Raw value from YAML for validation error messages
}
