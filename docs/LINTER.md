# Linter Configuration

tavern-go uses `golangci-lint` for code quality checks.

## Enabled Linters

We keep it simple with essential linters only:

- **errcheck**: Detects unchecked errors (critical for Go)
- **govet**: Standard Go vet analysis
- **staticcheck**: Advanced static analysis
- **unused**: Finds unused code
- **ineffassign**: Detects ineffective assignments
- **misspell**: Catches spelling mistakes

## Usage

```bash
# Run linter
make lint

# Auto-format code
make fmt
```

## Installation

If `golangci-lint` is not installed:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Philosophy

We focus on **real bugs** and avoid noise from style rules. The goal is:
- ✅ Catch actual errors
- ✅ Keep it fast
- ✅ Minimize false positives
- ❌ No nitpicky style rules

## Current Issues

Run `make lint` to see current issues. All must be fixed before merging PRs.
