package core

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/systemquest/tavern-go/pkg/request"
	"github.com/systemquest/tavern-go/pkg/response"
	"github.com/systemquest/tavern-go/pkg/schema"
	"github.com/systemquest/tavern-go/pkg/util"
	yamlpkg "github.com/systemquest/tavern-go/pkg/yaml"
)

// Runner executes Tavern tests
type Runner struct {
	config    *Config
	loader    *yamlpkg.Loader
	validator *schema.Validator
	logger    *logrus.Logger
}

// Config holds runner configuration
type Config struct {
	BaseDir      string
	GlobalConfig map[string]interface{}
	Variables    map[string]interface{}
	Verbose      bool
	Debug        bool
}

// NewRunner creates a new test runner
func NewRunner(config *Config) (*Runner, error) {
	if config == nil {
		config = &Config{
			BaseDir:   ".",
			Variables: make(map[string]interface{}),
		}
	}

	if config.GlobalConfig == nil {
		config.GlobalConfig = make(map[string]interface{})
	}

	if config.Variables == nil {
		config.Variables = make(map[string]interface{})
	}

	// Create logger
	logger := logrus.New()
	if config.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else if config.Verbose {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.WarnLevel)
	}

	// Create schema validator
	validator, err := schema.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &Runner{
		config:    config,
		loader:    yamlpkg.NewLoader(config.BaseDir),
		validator: validator,
		logger:    logger,
	}, nil
}

// RunFile runs all tests in a file
func (r *Runner) RunFile(filename string) error {
	r.logger.Infof("Loading tests from %s", filename)

	// Load tests
	tests, err := r.loader.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load tests: %w", err)
	}

	r.logger.Infof("Found %d test(s)", len(tests))

	// Run each test
	var firstError error
	for i, test := range tests {
		r.logger.Infof("Running test %d/%d: %s", i+1, len(tests), test.TestName)

		// Validate test schema
		if err := r.validator.Validate(test); err != nil {
			r.logger.Errorf("Schema validation failed for test '%s': %v", test.TestName, err)
			if firstError == nil {
				firstError = err
			}
			continue
		}

		// Run test
		if err := r.RunTest(test); err != nil {
			r.logger.Errorf("Test failed: %s: %v", test.TestName, err)
			if firstError == nil {
				firstError = err
			}
			continue
		}

		r.logger.Infof("Test passed: %s", test.TestName)
	}

	return firstError
}

// RunTest runs a single test
func (r *Runner) RunTest(test *schema.TestSpec) error {
	r.logger.Infof("Running test: %s", test.TestName)

	// Create shared HTTP client for session persistence (aligned with tavern-py's requests.Session)
	// This enables:
	// - Cookie persistence across stages
	// - Connection reuse (HTTP keep-alive)
	// - Session-based authentication
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}

	sharedHTTPClient := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically (aligned with tavern-py)
			return http.ErrUseLastResponse
		},
		Jar: jar, // Enable automatic cookie management
	}

	// Create shared persistent cookies map for clear_session_cookies support
	// This map is shared across all stages to track persistent cookies
	sharedPersistentCookies := make(map[string][]*http.Cookie)

	// Initialize test configuration
	testConfig := &request.Config{
		Variables:         make(map[string]interface{}),
		HTTPClient:        sharedHTTPClient,        // Share HTTP client across all stages
		PersistentCookies: sharedPersistentCookies, // Share persistent cookies tracking
	}

	// Inject tavern magic variables (aligned with tavern-py commit 1b55d6e)
	// Provides access to environment variables via {tavern.env_vars.VAR_NAME}
	testConfig.Variables["tavern"] = map[string]interface{}{
		"env_vars": getEnvVarsMap(),
	}

	// Merge global variables
	for k, v := range r.config.Variables {
		testConfig.Variables[k] = v
	}

	// Merge global config variables
	if globalVars, ok := r.config.GlobalConfig["variables"].(map[string]interface{}); ok {
		for k, v := range globalVars {
			testConfig.Variables[k] = v
		}
	}

	// Process includes
	for _, include := range test.Includes {
		r.logger.Debugf("Processing include: %s", include.Name)
		for k, v := range include.Variables {
			testConfig.Variables[k] = v
		}
	}

	// Run each stage
	for i, stage := range test.Stages {
		r.logger.Infof("Running stage %d/%d: %s", i+1, len(test.Stages), stage.Name)

		// Delay before stage execution
		delay(&stage, "before")

		// Protocol detection - check stage-level keys (aligned with tavern-py)
		// tavern-py checks: if "request" in stage / elif "mqtt_publish" in stage
		if stage.Request != nil {
			// REST/HTTP protocol
			if stage.Response == nil {
				return fmt.Errorf("stage '%s': REST request requires response specification", stage.Name)
			}

			executor := request.NewRestClient(testConfig)
			resp, err := executor.Execute(*stage.Request)
			if err != nil {
				return fmt.Errorf("stage '%s' request failed: %w", stage.Name, err)
			}

			// Inject request_vars into tavern namespace (aligned with tavern-py commit 35e52d9)
			// Enables access to request parameters in response validation: {tavern.request_vars.json.field}
			if tavernVars, ok := testConfig.Variables["tavern"].(map[string]interface{}); ok {
				tavernVars["request_vars"] = executor.RequestVars
			}

			validatorConfig := &response.Config{
				Variables: testConfig.Variables,
			}
			validator := response.NewRestValidator(stage.Name, *stage.Response, validatorConfig)
			saved, err := validator.Verify(resp)
			if err != nil {
				return fmt.Errorf("stage '%s' validation failed: %w", stage.Name, err)
			}

			// Clean up request_vars after validation (aligned with tavern-py commit 35e52d9)
			if tavernVars, ok := testConfig.Variables["tavern"].(map[string]interface{}); ok {
				delete(tavernVars, "request_vars")
			}

			// Save variables for next stages
			for k, v := range saved {
				r.logger.Debugf("Saved variable: %s = %v", k, v)
				testConfig.Variables[k] = v
			}
		} else {
			// Future protocols can be added here with elif-style checks:
			// } else if stage.MQTTPublish != nil {
			//     // MQTT protocol
			// } else if stage.Command != nil {
			//     // Shell/CLI protocol
			// } else {
			return fmt.Errorf("stage '%s': unable to detect protocol (no request field found)", stage.Name)
		}

		r.logger.Infof("Stage passed: %s", stage.Name)

		// Delay after stage execution
		delay(&stage, "after")
	}

	return nil
}

// LoadGlobalConfig loads a global configuration file
func (r *Runner) LoadGlobalConfig(filename string) error {
	r.logger.Infof("Loading global config from %s", filename)

	config, err := r.loader.LoadGlobalConfig(filename)
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	r.config.GlobalConfig = config

	// Merge variables
	if vars, ok := config["variables"].(map[string]interface{}); ok {
		for k, v := range vars {
			r.config.Variables[k] = v
		}
	}

	return nil
}

// SetVariable sets a variable in the runner config
func (r *Runner) SetVariable(key string, value interface{}) {
	r.config.Variables[key] = value
}

// GetVariable gets a variable from the runner config
func (r *Runner) GetVariable(key string) (interface{}, bool) {
	val, ok := r.config.Variables[key]
	return val, ok
}

// GetLogger returns the logger
func (r *Runner) GetLogger() *logrus.Logger {
	return r.logger
}

// ValidateFile validates a test file without running it
func (r *Runner) ValidateFile(filename string) error {
	tests, err := r.loader.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load tests: %w", err)
	}

	for _, test := range tests {
		if err := r.validator.Validate(test); err != nil {
			return util.NewBadSchemaError(
				fmt.Sprintf("validation failed for test '%s'", test.TestName),
				err,
			)
		}
	}

	return nil
}

// getEnvVarsMap returns all environment variables as a map
// Aligned with tavern-py commit 1b55d6e: provides access to os.environ
func getEnvVarsMap() map[string]interface{} {
	envMap := make(map[string]interface{})
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	return envMap
}
