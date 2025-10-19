package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// SaveConfig represents a union type for save configuration.
// It can be either a regular SaveSpec or an extension ExtSpec.
// This design provides type safety and eliminates the need for interface{}.
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
		sc.spec = nil
		return nil
	}

	// Otherwise, it's a regular SaveSpec
	var spec SaveSpec
	if err := node.Decode(&spec); err != nil {
		return fmt.Errorf("failed to decode SaveSpec: %w", err)
	}

	sc.spec = &spec
	sc.extension = nil
	return nil
}

// MarshalYAML implements custom YAML marshaling
func (sc *SaveConfig) MarshalYAML() (interface{}, error) {
	if sc == nil {
		return nil, nil
	}

	if sc.IsExtension() {
		return map[string]interface{}{
			"$ext": sc.extension,
		}, nil
	}

	if sc.IsRegular() {
		return sc.spec, nil
	}

	return nil, fmt.Errorf("SaveConfig is empty (neither spec nor extension set)")
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

		return NewExtensionSave(ext), nil
	}

	// Regular SaveSpec - handle various field types
	spec := &SaveSpec{}

	if bodyData, ok := mapData["body"]; ok {
		spec.Body = convertToStringMap(bodyData)
	}

	if headersData, ok := mapData["headers"]; ok {
		spec.Headers = convertToStringMap(headersData)
	}

	if paramsData, ok := mapData["redirect_query_params"]; ok {
		spec.RedirectQueryParams = convertToStringMap(paramsData)
	}

	return NewRegularSave(spec), nil
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
