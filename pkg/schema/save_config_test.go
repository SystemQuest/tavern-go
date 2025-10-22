package schema

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewRegularSave(t *testing.T) {
	spec := &SaveSpec{
		Body: map[string]interface{}{"token": "access_token"},
	}

	config := NewRegularSave(spec)

	if config == nil {
		t.Fatal("Expected non-nil SaveConfig")
	}
	if !config.IsRegular() {
		t.Error("Expected IsRegular() to return true")
	}
	if config.IsExtension() {
		t.Error("Expected IsExtension() to return false")
	}
	if config.GetSpec() != spec {
		t.Error("GetSpec() returned wrong spec")
	}
	if config.GetExtension() != nil {
		t.Error("GetExtension() should return nil for regular save")
	}
}

func TestNewExtensionSave(t *testing.T) {
	ext := &ExtSpec{
		Function: "my:validator",
		ExtraKwargs: map[string]interface{}{
			"pattern": "test",
		},
	}

	config := NewExtensionSave(ext)

	if config == nil {
		t.Fatal("Expected non-nil SaveConfig")
	}
	if !config.IsExtension() {
		t.Error("Expected IsExtension() to return true")
	}
	if config.IsRegular() {
		t.Error("Expected IsRegular() to return false")
	}
	if config.GetExtension() != ext {
		t.Error("GetExtension() returned wrong extension")
	}
	if config.GetSpec() != nil {
		t.Error("GetSpec() should return nil for extension save")
	}
}

func TestSaveConfig_NilHandling(t *testing.T) {
	var config *SaveConfig = nil

	// Nil SafeConfig should handle gracefully
	if config.IsRegular() {
		t.Error("Nil SaveConfig should return false for IsRegular()")
	}
	if config.IsExtension() {
		t.Error("Nil SaveConfig should return false for IsExtension()")
	}
	if config.GetSpec() != nil {
		t.Error("Nil SaveConfig should return nil for GetSpec()")
	}
	if config.GetExtension() != nil {
		t.Error("Nil SaveConfig should return nil for GetExtension()")
	}
}

func TestSaveConfig_UnmarshalYAML_Regular(t *testing.T) {
	yamlData := `
body:
  token: access_token
  user_id: id
headers:
  session: Set-Cookie
`

	var config SaveConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !config.IsRegular() {
		t.Error("Expected regular save config")
	}
	if config.IsExtension() {
		t.Error("Should not be extension save")
	}

	spec := config.GetSpec()
	if spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	if spec.Body["token"] != "access_token" {
		t.Errorf("Expected body.token='access_token', got '%s'", spec.Body["token"])
	}
	if spec.Body["user_id"] != "id" {
		t.Errorf("Expected body.user_id='id', got '%s'", spec.Body["user_id"])
	}
	if spec.Headers["session"] != "Set-Cookie" {
		t.Errorf("Expected headers.session='Set-Cookie', got '%s'", spec.Headers["session"])
	}
}

func TestSaveConfig_UnmarshalYAML_Extension(t *testing.T) {
	yamlData := `
$ext:
  function: tavern.testutils.helpers:validate_regex
  extra_kwargs:
    expression: "test-\\d+"
`

	var config SaveConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !config.IsExtension() {
		t.Error("Expected extension save config")
	}
	if config.IsRegular() {
		t.Error("Should not be regular save")
	}

	ext := config.GetExtension()
	if ext == nil {
		t.Fatal("Expected non-nil extension")
	}

	if ext.Function != "tavern.testutils.helpers:validate_regex" {
		t.Errorf("Expected function='tavern.testutils.helpers:validate_regex', got '%s'", ext.Function)
	}

	expression, ok := ext.ExtraKwargs["expression"].(string)
	if !ok || expression != "test-\\d+" {
		t.Errorf("Expected extra_kwargs.expression='test-\\\\d+', got '%v'", ext.ExtraKwargs["expression"])
	}
}

func TestSaveConfig_UnmarshalYAML_WithAnchor(t *testing.T) {
	// Simulates YAML anchor scenario where types become map[string]interface{}
	yamlData := `
test1: &save_anchor
  body:
    token: access_token
    
test2:
  save: *save_anchor
`

	type TestStruct struct {
		Test1 SaveConfig `yaml:"test1"`
		Test2 struct {
			Save SaveConfig `yaml:"save"`
		} `yaml:"test2"`
	}

	var tests TestStruct
	err := yaml.Unmarshal([]byte(yamlData), &tests)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check test1 (anchor definition)
	if !tests.Test1.IsRegular() {
		t.Error("test1 should be regular save")
	}
	if tests.Test1.GetSpec().Body["token"] != "access_token" {
		t.Error("test1 body.token should be 'access_token'")
	}

	// Check test2 (anchor reference)
	if !tests.Test2.Save.IsRegular() {
		t.Error("test2.save should be regular save")
	}
	if tests.Test2.Save.GetSpec().Body["token"] != "access_token" {
		t.Error("test2.save body.token should be 'access_token'")
	}
}

func TestSaveConfig_MarshalYAML_Regular(t *testing.T) {
	config := NewRegularSave(&SaveSpec{
		Body: map[string]interface{}{
			"token": "access_token",
		},
		Headers: map[string]string{
			"session": "Set-Cookie",
		},
	})

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back to verify
	var result SaveConfig
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if !result.IsRegular() {
		t.Error("Expected regular save after round-trip")
	}

	spec := result.GetSpec()
	if spec.Body["token"] != "access_token" {
		t.Error("Body.token lost in round-trip")
	}
	if spec.Headers["session"] != "Set-Cookie" {
		t.Error("Headers.session lost in round-trip")
	}
}

func TestSaveConfig_MarshalYAML_Extension(t *testing.T) {
	config := NewExtensionSave(&ExtSpec{
		Function: "my:func",
		ExtraKwargs: map[string]interface{}{
			"param": "value",
		},
	})

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back to verify
	var result SaveConfig
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if !result.IsExtension() {
		t.Error("Expected extension save after round-trip")
	}

	ext := result.GetExtension()
	if ext.Function != "my:func" {
		t.Error("Function lost in round-trip")
	}
	if ext.ExtraKwargs["param"] != "value" {
		t.Error("ExtraKwargs.param lost in round-trip")
	}
}

func TestSaveConfig_MarshalYAML_Nil(t *testing.T) {
	var config *SaveConfig = nil

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal nil: %v", err)
	}

	// Should produce "null" or empty
	if len(data) > 10 { // "null\n" is about 5 bytes
		t.Errorf("Expected short output for nil, got: %s", string(data))
	}
}

func TestNewSaveConfigFromInterface_Regular(t *testing.T) {
	// Test with map[string]string (direct type)
	data := map[string]interface{}{
		"body": map[string]string{
			"token": "access_token",
		},
		"headers": map[string]string{
			"session": "Set-Cookie",
		},
	}

	config, err := NewSaveConfigFromInterface(data)
	if err != nil {
		t.Fatalf("Failed to create from interface: %v", err)
	}

	if !config.IsRegular() {
		t.Error("Expected regular save")
	}

	spec := config.GetSpec()
	if spec.Body["token"] != "access_token" {
		t.Error("Body.token not preserved")
	}
	if spec.Headers["session"] != "Set-Cookie" {
		t.Error("Headers.session not preserved")
	}
}

func TestNewSaveConfigFromInterface_RegularWithInterfaceMap(t *testing.T) {
	// Test with map[string]interface{} (YAML anchor scenario)
	data := map[string]interface{}{
		"body": map[string]interface{}{
			"token":   "access_token",
			"user_id": "id",
		},
	}

	config, err := NewSaveConfigFromInterface(data)
	if err != nil {
		t.Fatalf("Failed to create from interface: %v", err)
	}

	if !config.IsRegular() {
		t.Error("Expected regular save")
	}

	spec := config.GetSpec()
	if spec.Body["token"] != "access_token" {
		t.Error("Body.token not preserved from interface{} map")
	}
	if spec.Body["user_id"] != "id" {
		t.Error("Body.user_id not preserved from interface{} map")
	}
}

func TestNewSaveConfigFromInterface_Extension(t *testing.T) {
	data := map[string]interface{}{
		"$ext": map[string]interface{}{
			"function": "my:validator",
			"extra_kwargs": map[string]interface{}{
				"pattern": "test",
			},
		},
	}

	config, err := NewSaveConfigFromInterface(data)
	if err != nil {
		t.Fatalf("Failed to create from interface: %v", err)
	}

	if !config.IsExtension() {
		t.Error("Expected extension save")
	}

	ext := config.GetExtension()
	if ext.Function != "my:validator" {
		t.Error("Function not preserved")
	}
	if ext.ExtraKwargs["pattern"] != "test" {
		t.Error("ExtraKwargs not preserved")
	}
}

func TestNewSaveConfigFromInterface_Nil(t *testing.T) {
	config, err := NewSaveConfigFromInterface(nil)
	if err != nil {
		t.Errorf("Expected no error for nil, got: %v", err)
	}
	if config != nil {
		t.Error("Expected nil config for nil input")
	}
}

func TestNewSaveConfigFromInterface_InvalidType(t *testing.T) {
	_, err := NewSaveConfigFromInterface("invalid")
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

func TestNewSaveConfigFromInterface_InvalidExtType(t *testing.T) {
	data := map[string]interface{}{
		"$ext": "invalid", // Should be a map
	}

	_, err := NewSaveConfigFromInterface(data)
	if err == nil {
		t.Error("Expected error for invalid $ext type")
	}
}

func TestNewSaveConfigFromInterface_MissingFunction(t *testing.T) {
	data := map[string]interface{}{
		"$ext": map[string]interface{}{
			"extra_kwargs": map[string]interface{}{},
		},
	}

	_, err := NewSaveConfigFromInterface(data)
	if err == nil {
		t.Error("Expected error for missing function")
	}
}

func TestConvertToStringMap_DirectStringMap(t *testing.T) {
	input := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result := convertToStringMap(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}
	if result["key1"] != "value1" {
		t.Error("key1 not preserved")
	}
	if result["key2"] != "value2" {
		t.Error("key2 not preserved")
	}
}

func TestConvertToStringMap_InterfaceMap(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 123, // Non-string value should be ignored
	}

	result := convertToStringMap(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries (ignoring non-string), got %d", len(result))
	}
	if result["key1"] != "value1" {
		t.Error("key1 not preserved")
	}
	if result["key2"] != "value2" {
		t.Error("key2 not preserved")
	}
	if _, exists := result["key3"]; exists {
		t.Error("Non-string key3 should be ignored")
	}
}

func TestConvertToStringMap_InvalidType(t *testing.T) {
	result := convertToStringMap("invalid")

	if len(result) != 0 {
		t.Error("Expected empty map for invalid type")
	}
}

func TestConvertToStringMap_Nil(t *testing.T) {
	result := convertToStringMap(nil)

	if len(result) != 0 {
		t.Error("Expected empty map for nil")
	}
}
