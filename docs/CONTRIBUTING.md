# Contributing to Gobot

Thank you for your interest in contributing to Gobot! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Docker (optional, for development containers)
- VS Code with Dev Containers extension (recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/gobot.git
   cd gobot
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/your-org/gobot.git
   ```

## Development Setup

### Option 1: Dev Container (Recommended)

1. Install VS Code and the Dev Containers extension
2. Open the project in VS Code
3. When prompted, click "Reopen in Container"
4. The development environment will be automatically configured

### Option 2: Local Development

1. Install Go 1.21 or later
2. Clone the repository
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Install development tools:
   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

### Verify Setup

Run the test suite to verify your setup:

```bash
go test ./...
```

## Contributing Guidelines

### Types of Contributions

We welcome various types of contributions:

- **Bug fixes** - Fix issues and improve reliability
- **New features** - Add new functionality
- **Documentation** - Improve docs, examples, and guides
- **Tests** - Add or improve test coverage
- **Performance** - Optimize code and improve performance
- **Security** - Fix security vulnerabilities

### Issue Guidelines

Before creating an issue:

1. **Search existing issues** - Check if the issue has already been reported
2. **Use the issue template** - Fill out the appropriate template
3. **Provide details** - Include steps to reproduce, expected vs actual behavior
4. **Include environment** - OS, Go version, Gobot version

### Feature Requests

For feature requests:

1. **Check the roadmap** - See if it's already planned
2. **Explain the use case** - Why is this feature needed?
3. **Consider alternatives** - Is there already a way to achieve this?
4. **Think about implementation** - How would this be implemented?

## Code Style

### Go Code Style

Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines:

- Use `gofmt` for formatting
- Follow naming conventions
- Write clear, readable code
- Add comments for exported functions and types

### Project Structure

Follow the established project structure:

```
gobot/
├── cmd/gobot/          # CLI entry point
├── internal/           # Private application code
│   ├── parser/         # YAML parsing
│   ├── keywords/       # Keyword execution
│   ├── runner/         # Test orchestration
│   └── ...
├── pkg/                # Public libraries (future)
├── docs/               # Documentation
├── testcases/          # Example test cases
└── tests/              # Framework tests
```

### Naming Conventions

- **Files**: `snake_case.go`
- **Directories**: `snake_case/`
- **Functions**: `PascalCase` for exported, `camelCase` for private
- **Variables**: `camelCase`
- **Constants**: `UPPER_SNAKE_CASE`
- **Types**: `PascalCase`

### Code Formatting

Use `goimports` for automatic formatting:

```bash
goimports -w .
```

Or configure your editor to run `goimports` on save.

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test ./internal/parser -v

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

- **Unit tests** - Test individual functions and methods
- **Integration tests** - Test component interactions
- **End-to-end tests** - Test complete workflows
- **Benchmarks** - Test performance

### Test Guidelines

- Write tests for new functionality
- Maintain >80% code coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test error conditions and edge cases

### Example Test

```go
func TestParseTestCase(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected TestCase
        hasError bool
    }{
        {
            name: "valid test case",
            input: `
testcase: "Test"
steps:
  - keyword: log
    args: ["Hello"]
`,
            expected: TestCase{
                Name: "Test",
                Steps: []Step{
                    {Keyword: "log", Args: []interface{}{"Hello"}},
                },
            },
            hasError: false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParseTestCase([]byte(tt.input))
            if tt.hasError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

## Documentation

### Documentation Standards

- Write clear, concise documentation
- Include code examples
- Keep documentation up to date
- Use consistent formatting

### Documentation Types

- **User documentation** - How to use Gobot
- **API documentation** - Code reference
- **Contributing guides** - Development information
- **Examples** - Sample test cases and configurations

### Updating Documentation

When making code changes:

1. Update relevant documentation
2. Add examples for new features
3. Update API documentation
4. Review existing docs for accuracy

## Pull Request Process

### Before Submitting

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write code following the style guidelines
   - Add tests for new functionality
   - Update documentation

3. **Run checks**:
   ```bash
   # Format code
   goimports -w .
   
   # Run linter
   golangci-lint run
   
   # Run tests
   go test ./...
   
   # Build
   go build ./cmd/gobot
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

Examples:
```
feat(parser): add support for custom keywords
fix(runner): handle timeout errors gracefully
docs(api): update HTTP client documentation
```

### Submitting the PR

1. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a pull request**:
   - Use the PR template
   - Describe the changes clearly
   - Link related issues
   - Include screenshots for UI changes

3. **Wait for review**:
   - Address review comments
   - Make requested changes
   - Update the PR as needed

### PR Review Process

- **Automated checks** must pass
- **Code review** required from maintainers
- **Documentation** must be updated
- **Tests** must be included
- **Breaking changes** require special consideration

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features, backward compatible
- **Patch** (0.0.X): Bug fixes, backward compatible

### Release Steps

1. **Create release branch**:
   ```bash
   git checkout -b release/v1.0.0
   ```

2. **Update version**:
   - Update version in code
   - Update documentation
   - Update changelog

3. **Create release**:
   - Tag the release
   - Create GitHub release
   - Update documentation

4. **Announce**:
   - Update release notes
   - Announce on social media
   - Update package managers

## Getting Help

### Questions and Discussion

- **GitHub Discussions**: For questions and general discussion
- **GitHub Issues**: For bugs and feature requests
- **Documentation**: Check the docs first

### Mentorship

New contributors can:

- Ask for help in GitHub Discussions
- Request mentorship from maintainers
- Start with "good first issue" labels
- Join community calls (if available)

## Recognition

Contributors are recognized through:

- **Contributors list** in README
- **Release notes** for significant contributions
- **GitHub profile** contributions graph
- **Community acknowledgments**

## License

By contributing to Gobot, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Gobot! Your contributions help make the project better for everyone. 