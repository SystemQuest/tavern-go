package yaml

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

// processIncludes processes !include directives in YAML
func (l *Loader) processIncludes(data string) (string, error) {
	// Pattern to match !include filename (with optional leading dash for array items)
	re := regexp.MustCompile(`(-\s+)?!include\s+(\S+)`)

	result := data
	matches := re.FindAllStringSubmatch(data, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		fullMatch := match[0]
		arrayPrefix := match[1] // Will be "- " if it's an array item, empty otherwise
		filename := match[2]

		// Read the included file
		includePath := filepath.Join(l.baseDir, filename)
		content, err := os.ReadFile(includePath)
		if err != nil {
			return "", fmt.Errorf("failed to read include file %s: %w", filename, err)
		}

		// Recursively process includes in the included file
		processedContent, err := l.processIncludes(string(content))
		if err != nil {
			return "", err
		}

		// Remove the leading "---" if present in included file
		processedContent = strings.TrimPrefix(processedContent, "---\n")
		processedContent = strings.TrimPrefix(processedContent, "---\r\n")
		processedContent = strings.TrimSpace(processedContent)

		// If this is an array item (has "- " prefix), remove the dash since
		// the included content will become the array item
		var replacement string
		if arrayPrefix != "" {
			// For array items, indent the included content to align with array items
			indent := l.getIndent(data, fullMatch)
			indentedContent := l.indentYAML(processedContent, indent)
			replacement = indentedContent
		} else {
			// For non-array contexts, just indent normally
			indentedContent := l.indentYAML(processedContent, l.getIndent(data, fullMatch))
			replacement = indentedContent
		}

		result = strings.Replace(result, fullMatch, replacement, 1)
	}

	return result, nil
}

// getIndent determines the indentation level of a match
func (l *Loader) getIndent(data, match string) int {
	idx := strings.Index(data, match)
	if idx == -1 {
		return 0
	}

	// Find the start of the line
	lineStart := idx
	for lineStart > 0 && data[lineStart-1] != '\n' {
		lineStart--
	}

	// Count spaces
	indent := 0
	for i := lineStart; i < idx; i++ {
		switch data[i] {
		case ' ':
			indent++
		case '\t':
			indent += 4
		}
	}

	return indent
}

// indentYAML indents YAML content
func (l *Loader) indentYAML(content string, spaces int) string {
	if spaces == 0 {
		return content
	}

	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}

	return strings.Join(lines, "\n")
}

// processCustomTags recursively processes custom YAML tags like !anything, !int, !float, !include
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

	// Check for !int tag
	if node.Tag == "!int" {
		// Mark as type convert token
		node.Tag = "!!str"
		node.Value = "<<INT>>" + node.Value
		node.Kind = goyaml.ScalarNode
		return
	}

	// Check for !float tag
	if node.Tag == "!float" {
		// Mark as type convert token
		node.Tag = "!!str"
		node.Value = "<<FLOAT>>" + node.Value
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

		var test schema.TestSpec
		err = node.Decode(&test)
		if err != nil {
			return nil, fmt.Errorf("failed to decode test spec: %w", err)
		}

		// Skip empty documents
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
