# Changelog

All notable changes to the Robogo VS Code Extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2024-01-XX

### Added
- **Enhanced Assertion Support**: Complete rewrite of assertion functionality with operator support
  - Numeric comparisons: `==`, `!=`, `>`, `<`, `>=`, `<=`
  - String comparisons: `==`, `!=`, `contains`, `not_contains`, `starts_with`, `ends_with`
  - Proper YAML block-style syntax support for complex arguments
  - Assertion operator autocompletion in VS Code
  - Syntax highlighting for comparison operators
- **New Code Snippets**:
  - `robogo-assert-gt`: Assert greater than
  - `robogo-assert-contains`: Assert string contains
  - `robogo-assert-retry`: Assert with retry support
- **Improved YAML Parsing**: Better support for block-style arrays vs inline arrays
- **Enhanced Documentation**: Updated examples and descriptions for assertion usage

### Fixed
- YAML parsing issues with complex argument arrays
- Assertion action not supporting comparison operators
- Inconsistent argument passing in assertion contexts

## [0.3.0] - 2024-01-XX

### Added
- **Verbosity Support**: Multiple verbosity levels (`basic`, `detailed`, `debug`) for better debugging and monitoring
- **Enhanced Autocompletion**: Context-aware suggestions for verbosity levels
- **New Code Snippets**: 
  - `robogo-http-verbose` - HTTP request with verbose output
  - `robogo-postgres-verbose` - Database query with detailed verbose output
  - `robogo-variable-verbose` - Variable operation with debug verbose output
  - `robogo-test-verbose` - Complete test with global verbose setting
- **Comprehensive Documentation**: 
  - Features Guide (`FEATURES.md`)
  - Quick Reference (`QUICK_REFERENCE.md`)
  - Verbosity documentation with examples
- **Enhanced Syntax Highlighting**: Support for `verbose` keyword
- **Better Variable Extraction**: Finds variables from `vars:` and `secrets:` sections

### Changed
- **Updated README**: Added verbosity documentation and examples
- **Enhanced Package Metadata**: Added verbosity-related keywords
- **Improved Autocompletion**: Better context detection for PostgreSQL and variable actions

### Fixed
- **Variable Suggestions**: Now properly extracts variables from all sections
- **Syntax Highlighting**: Added missing action keywords

## [0.2.0] - 2024-01-XX

### Added
- **PostgreSQL Support**: Autocompletion for database operations
- **Variable Management**: Enhanced variable operations and suggestions
- **Control Flow**: Support for if/else, for loops, and while loops
- **Enhanced Snippets**: Database and control flow snippets
- **Better Error Handling**: Improved error messages and debugging

### Changed
- **Updated Syntax**: Added PostgreSQL and control flow keywords
- **Enhanced Autocompletion**: Context-aware suggestions for database operations

## [0.1.0] - 2024-01-XX

### Added
- **Initial Release**: Basic extension functionality
- **Syntax Highlighting**: Support for Robogo keywords and actions
- **Autocompletion**: Action name suggestions
- **Hover Documentation**: Action documentation on hover
- **Code Snippets**: Basic test structure snippets
- **Test Execution**: Run tests directly from VS Code
- **Action Management**: List and explore available actions
- **HTTP Support**: Autocompletion for HTTP methods and headers
- **Time Operations**: Autocompletion for time formats
- **Secret Management**: Support for file-based and inline secrets

### Features
- Support for `.robogo`, `.yaml`, and `.yml` files
- Context-aware autocompletion
- Rich hover documentation
- Multiple output formats (console, JSON, markdown)
- Secret masking in output
- Variable substitution support

## [Unreleased]

### Planned
- **Web Interface**: Web-based test runner and dashboard
- **Plugin System**: Extensible framework for custom actions
- **Advanced Analytics**: Test performance metrics and trends
- **Scheduling**: Automated test execution schedules
- **Team Collaboration**: Multi-user support and sharing 