package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/systemquest/tavern-go/pkg/core"
	_ "github.com/systemquest/tavern-go/pkg/testutils" // Register extension functions
	"github.com/systemquest/tavern-go/pkg/version"
)

var (
	// Flags
	globalCfgs []string // Support multiple global config files (aligned with tavern-py commit 76569fd)
	verbose    bool
	debug      bool
	validate   bool
	skipXfail  bool // Skip tests marked with _xfail (aligned with tavern-py commit 369a4bb)
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tavern [test-file]",
	Short: "Tavern - A high-performance RESTful API testing framework",
	Long: `Tavern is a command-line tool for testing RESTful APIs using YAML-based test specifications.
	
It provides a simple, concise syntax for defining API tests with support for:
- Multi-stage test workflows
- Variable substitution and data passing between stages
- Custom validation functions
- JSON Schema validation

Visit https://systemquest.dev for more information.`,
	Version: version.Version,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runTests,
}

func init() {
	rootCmd.Flags().StringSliceVarP(&globalCfgs, "global-cfg", "c", []string{}, "One or more global configuration files (aligned with tavern-py commit 76569fd)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	rootCmd.Flags().BoolVar(&validate, "validate", false, "Validate test files without running")
	rootCmd.Flags().BoolVar(&skipXfail, "skip-xfail", false, "Skip tests marked with _xfail (aligned with tavern-py commit 369a4bb)")
}

func runTests(cmd *cobra.Command, args []string) error {
	testFile := args[0]

	// Create runner config
	config := &core.Config{
		BaseDir:   ".",
		Verbose:   verbose,
		Debug:     debug,
		SkipXfail: skipXfail,
	}

	// Create runner
	runner, err := core.NewRunner(config)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Load global config if specified
	if len(globalCfgs) > 0 {
		if err := runner.LoadGlobalConfigs(globalCfgs); err != nil {
			return fmt.Errorf("failed to load global configs: %w", err)
		}
	}

	// Validate only mode
	if validate {
		if err := runner.ValidateFile(testFile); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		fmt.Println("✓ Validation passed")
		return nil
	}

	// Run tests
	if err := runner.RunFile(testFile); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Println("✓ All tests passed")
	return nil
}
