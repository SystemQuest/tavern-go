package yaml

import (
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

// processCustomTags recursively processes custom YAML tags like !anything, !int, !anyint, !float, !anyfloat, !str, !anystr, !bool, !anybool, !include
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

	// Check for !bool or !anybool tag
	if node.Tag == "!bool" || node.Tag == "!anybool" {
		// Mark as type convert/matcher token for boolean
		// !bool "true" -> <<BOOL>>true (type converter)
		// !anybool -> <<BOOL>> (type matcher)
		node.Tag = "!!str"
		node.Value = "<<BOOL>>" + node.Value
		node.Kind = goyaml.ScalarNode
		return
	}

	// Check for !approx tag - for approximate floating point comparison
	// Aligned with tavern-py commit 53690cf: Feature/approx numbers (#101)
	if node.Tag == "!approx" {
		// Mark as approximate float matcher
		// !approx 3.14159 -> <<APPROX>>3.14159
		node.Tag = "!!str"
		node.Value = "<<APPROX>>" + node.Value
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

		// Decode the test spec directly
		// The processCustomTags call above has already modified the node tree
		// to replace custom tags with marker strings like <<ANYTHING>>, <<INT>>, etc.
		var test schema.TestSpec
		err = node.Decode(&test)
		if err != nil {
			return nil, fmt.Errorf("failed to decode test spec: %w", err)
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
