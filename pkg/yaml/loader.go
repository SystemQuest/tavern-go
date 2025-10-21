package yaml

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/systemquest/tavern-go/pkg/schema"
	goyaml "gopkg.in/yaml.v3"
)

// Loader loads and parses YAML test files
type Loader struct {
	baseDir string
	cache   map[string]interface{}
	logger  *logrus.Logger
}

// NewLoader creates a new YAML loader
func NewLoader(baseDir string) *Loader {
	return &Loader{
		baseDir: baseDir,
		cache:   make(map[string]interface{}),
		logger:  logrus.New(),
	}
}

// Load loads test specifications from a YAML file
func (l *Loader) Load(filename string) ([]*schema.TestSpec, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	l.baseDir = filepath.Dir(absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML documents (custom tags will be processed during parsing)
	tests, err := l.parseYAML(string(data), absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return tests, nil
}

// processCustomTags recursively processes custom YAML tags like !anything, !int, !anyint, !float, !anyfloat, !str, !anystr, !include
func (l *Loader) processCustomTags(node *goyaml.Node) {
	if node == nil {
		return
	}

	// Check for !include tag
	if node.Tag == "!include" {
		// Load the included file
		filename := node.Value
		includePath := filepath.Join(l.baseDir, filename)

		// Read the included file
		data, err := os.ReadFile(includePath)
		if err != nil {
			// If we can't read the file, leave the node as-is
			// The error will be caught during validation
			return
		}

		// Parse the included file into a YAML node
		var includedNode goyaml.Node
		err = goyaml.Unmarshal(data, &includedNode)
		if err != nil {
			return
		}

		// Process custom tags in the included content
		l.processCustomTags(&includedNode)

		// Replace the !include node with the content of the included file
		// The included file should have a document node at the root
		if includedNode.Kind == goyaml.DocumentNode && len(includedNode.Content) > 0 {
			// Copy the first content node (the actual data)
			*node = *includedNode.Content[0]
		}
		return
	}

	// Check for !anything tag
	if node.Tag == "!anything" {
		// Replace with a special marker value that will be recognized during validation
		node.Tag = "!!str"
		node.Value = "<<ANYTHING>>"
		node.Kind = goyaml.ScalarNode
		return
	}

	// Check for !int or !anyint tag
	if node.Tag == "!int" || node.Tag == "!anyint" {
		// Mark as type convert token
		node.Tag = "!!str"
		node.Value = "<<INT>>" + node.Value
		node.Kind = goyaml.ScalarNode
		return
	}

	// Check for !float or !anyfloat tag
	if node.Tag == "!float" || node.Tag == "!anyfloat" {
		// Mark as type convert token
		node.Tag = "!!str"
		node.Value = "<<FLOAT>>" + node.Value
		node.Kind = goyaml.ScalarNode
		return
	}

	// Check for !str or !anystr tag
	if node.Tag == "!str" || node.Tag == "!anystr" {
		// Mark as type convert token
		node.Tag = "!!str"
		node.Value = "<<STR>>" + node.Value
		node.Kind = goyaml.ScalarNode
		return
	}

	// Recursively process child nodes
	for _, child := range node.Content {
		l.processCustomTags(child)
	}
}

// parseYAML parses YAML content into test specifications
func (l *Loader) parseYAML(data string, filename string) ([]*schema.TestSpec, error) {
	var tests []*schema.TestSpec

	decoder := goyaml.NewDecoder(strings.NewReader(data))

	// Register custom tag resolver for !anything
	decoder.KnownFields(false)

	for {
		var node goyaml.Node
		err := decoder.Decode(&node)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		// Process custom tags like !anything
		l.processCustomTags(&node)

		// Decode into a generic map first to preserve the processed values
		var rawTest map[string]interface{}
		err = node.Decode(&rawTest)
		if err != nil {
			return nil, fmt.Errorf("failed to decode test spec: %w", err)
		}

		// Convert to JSON and back to preserve the processed values
		// This works because JSON doesn't have custom tags
		jsonBytes, err := json.Marshal(rawTest)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
		}

		var test schema.TestSpec
		err = json.Unmarshal(jsonBytes, &test)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal from JSON: %w", err)
		} // Skip empty documents
		if test.TestName == "" {
			l.logger.Warnf("Empty document in input file '%s'", filename)
			continue
		}

		tests = append(tests, &test)
	}

	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests found in file")
	}

	return tests, nil
}

// LoadGlobalConfig loads a global configuration file
func (l *Loader) LoadGlobalConfig(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	err = goyaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}
