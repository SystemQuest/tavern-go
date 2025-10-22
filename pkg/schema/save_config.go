package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// SaveConfig represents save configuration that can contain both
// regular save fields (body, headers, redirect_query_params) and
// an optional $ext function. Unlike the previous union type design,
// this allows them to coexist, matching tavern-py's behavior.
type SaveConfig struct {
	spec      *SaveSpec // Regular save configuration (body, headers, redirect_query_params)
	extension *ExtSpec  // Extension-based save ($ext with function)
}

// NewRegularSave creates a SaveConfig with regular SaveSpec
func NewRegularSave(spec *SaveSpec) *SaveConfig {
	return &SaveConfig{spec: spec}
}

// NewExtensionSave creates a SaveConfig with ExtSpec for extension-based save
func NewExtensionSave(ext *ExtSpec) *SaveConfig {
	return &SaveConfig{extension: ext}
}

// IsExtension returns true if this SaveConfig uses an extension ($ext)
func (sc *SaveConfig) IsExtension() bool {
	return sc != nil && sc.extension != nil
}

// IsRegular returns true if this SaveConfig uses regular SaveSpec
func (sc *SaveConfig) IsRegular() bool {
	return sc != nil && sc.spec != nil
}

// GetSpec returns the regular SaveSpec, or nil if this is an extension save
func (sc *SaveConfig) GetSpec() *SaveSpec {
	if sc == nil {
		return nil
	}
	return sc.spec
}

// GetExtension returns the ExtSpec, or nil if this is a regular save
func (sc *SaveConfig) GetExtension() *ExtSpec {
	if sc == nil {
		return nil
	}
	return sc.extension
}

// UnmarshalYAML implements custom YAML unmarshaling to handle both save types
func (sc *SaveConfig) UnmarshalYAML(node *yaml.Node) error {
	// First, unmarshal as a generic map to inspect the structure
	var mapData map[string]interface{}
	if err := node.Decode(&mapData); err != nil {
		return fmt.Errorf("failed to decode save config: %w", err)
	}

	// Check if it contains $ext key (extension-based save)
	if extData, hasExt := mapData["$ext"]; hasExt {
		// Create a new node from the $ext data
		extNode := &yaml.Node{}
		if err := extNode.Encode(extData); err != nil {
			return fmt.Errorf("failed to encode $ext data: %w", err)
		}

		// Unmarshal as ExtSpec
		var ext ExtSpec
		if err := extNode.Decode(&ext); err != nil {
			return fmt.Errorf("failed to decode $ext: %w", err)
		}

		sc.extension = &ext

		// Remove $ext from mapData before processing as SaveSpec
		delete(mapData, "$ext")
	}

	// Check if there are any regular save fields (body, headers, redirect_query_params)
	hasRegularFields := false
	for key := range mapData {
		if key == "body" || key == "headers" || key == "redirect_query_params" {
			hasRegularFields = true
			break
		}
	}

	// If there are regular fields, decode them as SaveSpec
	if hasRegularFields {
		// Create a new node from the modified mapData (without $ext)
		specNode := &yaml.Node{}
		if err := specNode.Encode(mapData); err != nil {
			return fmt.Errorf("failed to encode save spec data: %w", err)
		}

		var spec SaveSpec
		if err := specNode.Decode(&spec); err != nil {
			return fmt.Errorf("failed to decode SaveSpec: %w", err)
		}

		sc.spec = &spec
	}

	// Both can be nil if save: {} is specified
	return nil
}

// MarshalYAML implements custom YAML marshaling
func (sc *SaveConfig) MarshalYAML() (interface{}, error) {
	if sc == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	// Add $ext if present
	if sc.IsExtension() {
		result["$ext"] = sc.extension
	}

	// Add regular spec fields if present
	if sc.IsRegular() {
		spec := sc.spec
		if spec.Body != nil {
			result["body"] = spec.Body
		}
		if spec.Headers != nil {
			result["headers"] = spec.Headers
		}
		if spec.RedirectQueryParams != nil {
			result["redirect_query_params"] = spec.RedirectQueryParams
		}
	}

	// Return nil if both are empty
	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

// NewSaveConfigFromInterface creates SaveConfig from interface{}.
// This is provided for backward compatibility with code that works with interface{}.
// It handles both map[string]string and map[string]interface{} (common with YAML anchors).
func NewSaveConfigFromInterface(data interface{}) (*SaveConfig, error) {
	if data == nil {
		return nil, nil
	}

	mapData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("save config must be a map, got %T", data)
	}

	config := &SaveConfig{}

	// Check for $ext key
	if extData, hasExt := mapData["$ext"]; hasExt {
		extMap, ok := extData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("$ext must be a map, got %T", extData)
		}

		function, ok := extMap["function"].(string)
		if !ok {
			return nil, fmt.Errorf("$ext.function must be a string")
		}

		ext := &ExtSpec{
			Function: function,
		}

		if extraKwargs, ok := extMap["extra_kwargs"].(map[string]interface{}); ok {
			ext.ExtraKwargs = extraKwargs
		}
		if extraArgs, ok := extMap["extra_args"].([]interface{}); ok {
			ext.ExtraArgs = extraArgs
		}

		config.extension = ext
	}

	// Check for regular save fields
	spec := &SaveSpec{}
	hasRegularFields := false

	if bodyData, ok := mapData["body"]; ok {
		// Body can be:
		// 1. map[string]interface{} - may contain strings (JSON paths) or $ext objects
		// 2. map[string]string - direct JSON paths
		if bodyMap, ok := bodyData.(map[string]interface{}); ok {
			spec.Body = bodyMap
			hasRegularFields = true
		} else if bodyStrMap, ok := bodyData.(map[string]string); ok {
			// Convert map[string]string to map[string]interface{}
			spec.Body = make(map[string]interface{})
			for k, v := range bodyStrMap {
				spec.Body[k] = v
			}
			hasRegularFields = true
		}
	}

	if headersData, ok := mapData["headers"]; ok {
		spec.Headers = convertToStringMap(headersData)
		hasRegularFields = true
	}

	if paramsData, ok := mapData["redirect_query_params"]; ok {
		spec.RedirectQueryParams = convertToStringMap(paramsData)
		hasRegularFields = true
	}

	if hasRegularFields {
		config.spec = spec
	}

	return config, nil
}

// convertToStringMap converts interface{} to map[string]string.
// It handles both map[string]string (direct) and map[string]interface{} (YAML anchors).
func convertToStringMap(data interface{}) map[string]string {
	result := make(map[string]string)

	// Try map[string]string first (most common case)
	if m, ok := data.(map[string]string); ok {
		return m
	}

	// Try map[string]interface{} (common with YAML anchors/aliases)
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result
	}

	return result
}
