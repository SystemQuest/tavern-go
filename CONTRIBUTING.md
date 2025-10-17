# Contributing to Tavern-Go

Thank you for your interest in contributing to Tavern-Go! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/tavern-go.git
   cd tavern-go
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build the project:
   ```bash
   make build
   ```

5. Run tests:
   ```bash
   make test
   ```

## Development Workflow

### Branch Naming

- Feature: `feature/your-feature-name`
- Bug fix: `fix/bug-description`
- Documentation: `docs/what-you-document`

### Code Style

We follow standard Go conventions:

- Run `gofmt` before committing
- Use `golangci-lint` for linting
- Write meaningful commit messages

Format your code:
```bash
make fmt
```

Lint your code:
```bash
make lint
```

### Testing

- Write tests for new features
- Ensure all tests pass before submitting PR
- Aim for high test coverage

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make coverage
```

### Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

Example:
```
feat(extension): add support for custom savers

Add ability to register custom saver functions that extract
data from HTTP responses.

Closes #123
```

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass
4. Update CHANGELOG.md
5. Submit PR with clear description

### PR Title Format

```
<type>(<scope>): <description>
```

Example: `feat(core): add parallel test execution`

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How has this been tested?

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Code formatted
```

## Project Structure

```
tavern-go/
├── cmd/tavern/         # CLI application
├── pkg/                # Public packages
│   ├── core/          # Test execution engine
│   ├── request/       # HTTP request handling
│   ├── response/      # Response validation
│   ├── schema/        # Schema validation
│   ├── extension/     # Extension system
│   ├── yaml/          # YAML loading
│   └── util/          # Utilities
├── examples/          # Example tests
├── docs/              # Documentation
└── internal/          # Internal packages
```

## Adding New Features

### Adding a New Extension Function

1. Register in `pkg/extension/registry.go`:
   ```go
   func init() {
       RegisterValidator("myapp:validator", myValidator)
   }
   ```

2. Add tests in `pkg/extension/registry_test.go`

3. Document in README.md

### Adding a New CLI Flag

1. Add flag in `cmd/tavern/main.go`
2. Update help text
3. Add to documentation

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for public APIs
- Include examples where appropriate

## Reporting Issues

### Bug Reports

Include:
- Go version
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Test files (if applicable)

### Feature Requests

Include:
- Use case description
- Proposed solution
- Alternative solutions considered

## Code Review

All contributions require code review:

- Be respectful and constructive
- Respond to feedback promptly
- Update PR based on review comments

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

- Open an issue for questions
- Join discussions on GitHub
- Email: dev@systemquest.dev

## Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md
- Release notes
- Project README

Thank you for contributing to Tavern-Go! 🚀
