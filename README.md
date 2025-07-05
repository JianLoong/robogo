# Robogo - Test Automation Framework

A modern, git-driven test automation framework written in Go, inspired by Robot Framework.

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/robogo.git
cd robogo

# Install dependencies
go mod download

# Build the binary
go build -o robogo cmd/robogo/main.go
```

### Run Your First Test

```bash
# Run the hello world test
./robogo run testcases/hello-world.yaml
```

You should see output like:

```
🚀 Running test case: Hello World Test
📋 Description: A simple test to verify Robogo is working
📝 Steps: 5

Step 1: log
📝 Hello from Robogo!
✅ Step 1 completed in 123.45µs

Step 2: log
📝 This is a simple test case
✅ Step 2 completed in 45.67µs

Step 3: sleep
😴 Sleeping for 1s
✅ Step 3 completed in 1.000123s

Step 4: assert
✅ This should always pass
✅ Step 4 completed in 23.45µs

Step 5: log
📝 Test completed successfully!
✅ Step 5 completed in 34.56µs

🏁 Test completed in 1.000227s

📊 Test Results:
✅ Status: PASSED
⏱️  Duration: 1.000227s
📝 Steps: 5 total, 5 passed, 0 failed
```

## 📋 Available Keywords

### `log`
Outputs a message to the console.

```yaml
- keyword: log
  args: ["Hello, world!"]
```

### `sleep`
Pauses execution for a specified duration.

```yaml
- keyword: sleep
  args: [1]  # Sleep for 1 second
```

### `assert`
Verifies that two values are equal.

```yaml
- keyword: assert
  args: [actual_value, expected_value, "Optional message"]
```

## 📝 Test Case Format

Test cases are written in YAML format:

```yaml
testcase: "Test Case Name"
description: "Optional description"
steps:
  - keyword: log
    args: ["Message to log"]
  - keyword: sleep
    args: [1]
  - keyword: assert
    args: [true, true, "This should pass"]
```

## 🧪 Example Test Cases

- `testcases/hello-world.yaml` - Simple hello world test
- `testcases/basic-test.yaml` - Basic functionality test

## 🏗️ Project Structure

```
robogo/
├── cmd/robogo/          # CLI entry point
├── internal/           # Core framework code
│   ├── parser/         # YAML parsing
│   ├── keywords/       # Keyword execution
│   └── runner/         # Test orchestration
├── testcases/          # Example test cases
└── docs/              # Documentation
```

## 🔧 Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o robogo cmd/robogo/main.go
```

## 📚 Documentation

See the [docs/](docs/) directory for comprehensive documentation:

- [Installation Guide](docs/installation.md)
- [Quick Start Guide](docs/quickstart.md)
- [Test Case Writing Guide](docs/test-cases.md)
- [Contributing Guide](docs/CONTRIBUTING.md)

## 🎯 Roadmap

This is a minimal proof of concept. Future versions will include:

- [ ] Git integration for test case management
- [ ] HTTP request keywords with mTLS support
- [ ] Parallel test execution
- [ ] Plugin system for custom keywords
- [ ] Web interface and API
- [ ] Advanced reporting and analytics

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## 📄 License

[Add your license here]

---

**Robogo** - Modern test automation for the Go ecosystem. 