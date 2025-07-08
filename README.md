# Robogo - Modern Test Automation Framework

A modern, git-driven test automation framework written in Go, designed for comprehensive API testing, SWIFT message generation, database operations, and Test Data Management (TDM).

---

**Recent Improvements:**
- Robust, consistent error handling across all actions and outputs
- Output summary always prints, even on error or parallel execution
- Safe, deadlock-free parallel execution at suite and step level
- Modernized command usage: unified `run` command for both test files and suites
- Improved expect_error logic and step status handling
- Clean, up-to-date documentation and examples

---

## ✨ Key Features

- **🔧 Template-based SWIFT Message Generation** - Create and test SWIFT messages with dynamic variable substitution
- **🌐 HTTP API Testing** - Full HTTP support with mTLS, custom headers, and comprehensive response validation
- **💾 Database Integration** - PostgreSQL operations with connection pooling and secure credential management
- **📊 Test Data Management (TDM)** - Structured data sets, environment management, and data lifecycle
- **🎲 Enhanced Random Generation** - Support for both integer and decimal random values with precision control
- **🔄 Advanced Control Flow** - If statements, for loops, while loops with conditional logic and retry mechanisms
- **🔐 Secret Management** - Secure handling of API keys, certificates, and sensitive data with masking
- **📊 Multiple Output Formats** - Console, JSON, and Markdown reporting with detailed step-level analytics
- **⚡ Performance Testing** - Built-in timing, load testing, and retry capabilities
- **🔍 Comprehensive Validation** - Data validation, format checking, and assertion framework
- **🚀 Parallel Execution** - Concurrent test file execution and parallel step execution with dependency analysis
- **🔄 Batch Operations** - Parallel HTTP requests and database operations with concurrency control
- **🛠️ VS Code Integration** - Complete extension with syntax highlighting, autocomplete, and code snippets
- **✅ Consistent Error Handling** - Unified error formatting and reporting, including expect_error logic
- **📈 Reliable Output Summary** - Always-on summary table, even on error or panic

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or later (required)
- VS Code (optional, for enhanced development experience)

### Installation

```bash
# Clone the repository
git clone https://github.com/JianLoong/robogo.git
cd robogo

# Install dependencies
go mod download

# Build the binary
go build -o robogo.exe ./cmd/robogo
```

### Upgrading from Older Versions
- Replace any usage of `run-suite` with the new unified `run` command.
- Review your test suite YAML files for deprecated fields; see updated examples below.
- Ensure you are using Go 1.22+ for best compatibility and performance.

### VS Code Extension Setup

The project includes a complete VS Code extension for enhanced development:

```bash
# Quick setup - run the extension launcher
./run-extension.ps1
```

---

## 🏃 Usage

### Run a Test File or Suite

```bash
# Run a single test file
./robogo.exe run examples/sample.robogo

# Run a test suite (YAML file with 'testsuite:' at the top)
./robogo.exe run examples/test-suite.robogo

# Run multiple test files or suites in parallel
./robogo.exe run tests/core/test-assert.robogo tests/core/test-control-flow.robogo --parallel

# Limit concurrency
./robogo.exe run tests/*.robogo --parallel --max-concurrency 2
```

---

## ⚙️ Features & Best Practices

- **Consistent Error Handling:** All errors are formatted and reported uniformly, both in console and output files. Steps with errors are marked as FAILED unless `expect_error` is set and the error matches expectations.
- **Expect Error Logic:** If a step is expected to error (via `expect_error`), and the error occurs, the step is marked as PASSED.
- **Output Summary:** The summary table always prints, even if errors or panics occur during execution or in parallel mode.
- **Parallel Execution:**
  - Enable at suite or step level via YAML (`parallel: true` or `parallelism:` block).
  - Safe, deadlock-free result collection; always sends results even on panic.
  - See examples below for configuration.
- **Modern Command Usage:** Use `run` for both test files and suites; `run-suite` is deprecated.

---

## 🧩 Example Test Suite (YAML)

```yaml
testsuite: "Parallel Suite Example"
description: "Demonstrate parallel test case execution"
parallel: true
max_concurrency: 4

testcases:
  - testcase: "Test A"
    steps:
      - action: log
        args: ["Hello from A"]
  - testcase: "Test B"
    steps:
      - action: log
        args: ["Hello from B"]
```

---

## 🛠️ Troubleshooting

- **Missing Summary Table:**
  - Ensure you are using the latest version. The summary now always prints, even on error or panic.
- **Parallel Deadlocks:**
  - All goroutines now send results even on panic. If you see a hang, check for infinite loops or blocking steps.
- **Error Handling:**
  - Errors in step output or Go errors both mark steps as FAILED unless `expect_error` is set.
- **Upgrading:**
  - Replace `run-suite` with `run` in all scripts and documentation.

---

## 🔬 Development & Contribution

- Build: `go build -o robogo.exe ./cmd/robogo`
- Run all Go tests: `go test ./...`
- Run a Robogo test: `./robogo.exe run tests/core/test-assert.robogo`
- Run in parallel: `./robogo.exe run tests/*.robogo --parallel --max-concurrency 4`
- See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for code style, pull request, and issue guidelines.

---

## 📚 Documentation

- [Quick Start Guide](docs/quickstart.md)
- [Actions Reference](docs/actions.md)
- [Test Cases Guide](docs/test-cases.md)
- [CLI Reference](docs/cli-reference.md)
- [TDM Implementation Guide](docs/tdm-implementation.md)
- [Framework Comparison](docs/framework-comparison.md)

---

## 🏦 Use Cases

- Financial services: SWIFT message testing, payment API validation, TDM for banking
- API testing: REST, mTLS, performance, data-driven
- Database: PostgreSQL, Spanner, validation, lifecycle
- Automation: CI/CD, regression, load, cross-platform

---

## 📄 License

MIT License - see LICENSE file for details.

---

Robogo: Modern, robust, and scalable test automation for Go, with powerful error handling, parallel execution, and comprehensive reporting.

## Reserved Variable: __robogo_steps

robogo automatically populates a reserved variable named `__robogo_steps` after each step execution. This variable is a slice of maps, where each map contains the following fields for each step:
- `name`: The step's name
- `status`: The step's status (e.g., PASSED, FAILED)
- `output`: The step's output
- `error`: The step's error message (if any)
- `timestamp`: The time the step was executed

### Usage Example
You can reference the result or error of any previous step using indexed access:

```yaml
- name: Assert Timeout Error
  action: assert
  args:
    - "${__robogo_steps[0].error}"
    - ==
    - timeout
    - "Expected a timeout error when no message is available"
```

### Warning
`__robogo_steps` is reserved for internal use by robogo. If you manually set this variable in your test case, it will be overwritten and a warning will be printed.

```
gcloud config configurations create emulator
gcloud config set project test-project

```