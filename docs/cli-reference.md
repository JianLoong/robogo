# CLI Reference

This document provides a comprehensive reference for the Robogo command-line interface.

## Overview

Robogo provides a simple and intuitive CLI for running test cases and managing your test automation workflow.

## Command Structure

```bash
robogo [command] [options] [arguments]
```

## Global Options

All commands support these global options:

- `--version, -v`: Show version information
- `--help, -h`: Show help for the command

## Commands

### run

Executes a test case file.

**Syntax:**
```bash
robogo run [test-file] [options]
```

**Arguments:**
- `test-file`: Path to the test case file (supports .yaml, .yml, .robogo)

**Options:**
- `--output, -o`: Output format (console, json, markdown) [default: console]

**Behavior:**
- When using `json` or `markdown` output format, runner console output is suppressed (silent mode)
- Individual action outputs (like sleep messages, log messages) are still displayed
- Console format provides detailed step-by-step output from both runner and actions
- JSON and Markdown formats provide summary results only, with action outputs still visible

**Examples:**
```bash
# Run a test case with default console output
./robogo run tests/test-variables.robogo

# Run a test case with JSON output (runner output suppressed)
./robogo run tests/test-time-formats.robogo --output json

# Run a test case with markdown output (runner output suppressed)
./robogo run tests/test-time-formats.robogo --output markdown

# Run with different file extensions
./robogo run tests/test-syntax.yaml
./robogo run tests/example.yml
```

**Output Formats:**

1. **Console (default):**
   ```
   🚀 Running test case: Variable Support Test
   📋 Description: Test file to demonstrate variable assignment and usage
   📝 Steps: 10

   Step 1: get_time
   ✅ Step 1 completed in 123.45µs

   Step 2: log
   📝 Current timestamp: 2024-12-19 10:30:00
   ✅ Step 2 completed in 45.67µs

   Step 3: get_random
   ✅ Step 3 completed in 23.45µs

   Step 4: log
   📝 Random number: 42
   ✅ Step 4 completed in 34.56µs

   🏁 Test completed in 227.13µs

   📊 Test Results:
   ✅ Status: PASSED
   ⏱️  Duration: 227.13µs
   📝 Steps: 10 total, 10 passed, 0 failed
   ```

2. **JSON:**
   ```json
   {
     "testcase": "Variable Support Test",
     "status": "PASSED",
     "duration": "227.13µs",
     "total_steps": 10,
     "passed_steps": 10,
     "failed_steps": 0,
     "error_message": ""
   }
   ```

3. **Markdown:**
   ```markdown
   # Test Results: Variable Support Test

   ## Summary
   ✅ **Status:** PASSED  
   ⏱️ **Duration:** 227.13µs  
   📝 **Steps:** 10 total, 10 passed, 0 failed

   ## Test Case Details
   - **Name:** Variable Support Test
   - **Description:** Test file to demonstrate variable assignment and usage

   ## Step Results
   - ✅ Step 1: get_time (123.45µs)
   - ✅ Step 2: log (45.67µs)
   - ✅ Step 3: get_random (23.45µs)
   - ✅ Step 4: log (34.56µs)
   ```

### list

Lists all available actions with their descriptions and examples.

**Syntax:**
```bash
robogo list [options]
```

**Options:**
- `--output, -o`: Output format (console, json) [default: console]

**Examples:**
```bash
# List all actions in console format
./robogo list

# List all actions in JSON format
./robogo list --output json
```

**Console Output:**
```
📋 Available Actions (10 total):

- assert: Assert two values are equal
  Example: - action: assert
  args: ["expected", "actual", "message"]

- concat: Concatenate strings
  Example: - action: concat
  args: ["Hello", " ", "World"]
  result: message

- get_random: Get a random number
  Example: - action: get_random
  args: [100]
  result: random_number

- get_time: Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)
  Example: - action: get_time
  args: ["iso"]
  result: timestamp

- http: Perform HTTP request (GET, POST, PUT, DELETE, etc.). Supports client cert (cert/key) and custom CA (ca) options.
  Example: - action: http
  args: ["GET", "https://secure.example.com/api", {"cert": "client.crt", "key": "client.key", "ca": "ca.crt", "Authorization": "Bearer ..."}]
  result: response

- http_get: Perform HTTP GET request
  Example: - action: http_get
  args: ["https://api.example.com/users"]
  result: response

- http_post: Perform HTTP POST request
  Example: - action: http_post
  args: ["https://api.example.com/users", '{"name": "John"}']
  result: response

- length: Get length of string or array
  Example: - action: length
  args: ["Hello World"]
  result: str_length

- log: Log a message
  Example: - action: log
  args: ["message"]

- sleep: Sleep for a duration
  Example: - action: sleep
  args: [2]
```

**JSON Output:**
```json
[
  {
    "name": "assert",
    "description": "Assert two values are equal",
    "example": "- action: assert\n  args: [\"expected\", \"actual\", \"message\"]"
  },
  {
    "name": "concat",
    "description": "Concatenate strings",
    "example": "- action: concat\n  args: [\"Hello\", \" \", \"World\"]\n  result: message"
  }
]
```

### completions

Provides action name completions for autocomplete functionality.

**Syntax:**
```bash
robogo completions [prefix] [options]
```

**Arguments:**
- `prefix`: Optional prefix to filter completions

**Options:**
- `--output, -o`: Output format (console, json) [default: console]

**Examples:**
```bash
# Get all action completions
./robogo completions

# Get completions starting with "http"
./robogo completions http

# Get completions in JSON format
./robogo completions http --output json
```

**Console Output:**
```
🔍 Completions for 'http':
  http
  http_get
  http_post
```

**JSON Output:**
```json
["http", "http_get", "http_post"]
```

## Exit Codes

Robogo uses the following exit codes:

- `0`: Success - All tests passed
- `1`: Failure - One or more tests failed or command error
- `2`: Error - Invalid command or file not found

## Environment Variables

Robogo respects the following environment variables:

- `ROBOGO_CONFIG_PATH`: Path to configuration file
- `ROBOGO_LOG_LEVEL`: Logging level (debug, info, warn, error)
- `ROBOGO_CACHE_DIR`: Directory for caching

## File Extensions

Robogo supports the following file extensions for test cases:

- `.robogo` - Robogo-specific format
- `.yaml` - Standard YAML format
- `.yml` - Standard YAML format

## Examples

### Basic Usage

```bash
# Run a simple test
./robogo run tests/test-variables.robogo

# Run with different output format
./robogo run tests/test-http.robogo --output json

# List available actions
./robogo list

# Get completions for autocomplete
./robogo completions http
```

### Advanced Usage

```bash
# Run multiple test cases (one by one)
for test in tests/*.robogo; do
  echo "Running $test..."
  ./robogo run "$test"
done

# Run with error handling
./robogo run tests/test-variables.robogo || {
  echo "Test failed!"
  exit 1
}

# Generate JSON report for CI/CD
./robogo run tests/test-http.robogo --output json > test-results.json

# Run all tests and collect results
for test in tests/*.robogo; do
  echo "Running $test..."
  ./robogo run "$test" --output json > "${test%.robogo}.json"
done
```

### Integration Examples

**GitHub Actions:**
```yaml
name: Run Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go build -o robogo cmd/robogo/main.go
      - run: ./robogo run tests/ --output json > results.json
      - name: Upload test results
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: results.json
```

**Docker:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o robogo cmd/robogo/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/robogo /usr/local/bin/
ENTRYPOINT ["robogo"]
```

**Shell Script:**
```bash
#!/bin/bash
set -e

# Build Robogo
go build -o robogo cmd/robogo/main.go

# Run all test cases
for test_file in tests/*.robogo; do
  echo "Running $test_file..."
  ./robogo run "$test_file" --output json > "${test_file%.robogo}.json"
done

# Generate summary
echo "Test execution completed. Check individual JSON files for results."
```

**PowerShell (Windows):**
```powershell
# Build Robogo
go build -o robogo.exe cmd/robogo/main.go

# Run all test cases
Get-ChildItem tests/*.robogo | ForEach-Object {
    Write-Host "Running $($_.Name)..."
    ./robogo.exe run $_.FullName --output json > "$($_.BaseName).json"
}
```

## Troubleshooting

### Common Issues

**"command not found: robogo"**
- Ensure the binary is built: `go build -o robogo cmd/robogo/main.go`
- Check if the binary is in your PATH
- Try running with `./robogo` from the project directory

**"failed to parse YAML"**
- Check YAML syntax in your test file
- Ensure proper indentation
- Validate against the schema

**"unknown action"**
- Use `./robogo list` to see available actions
- Check action name spelling
- Ensure correct YAML structure

**"file not found"**
- Verify the test file path is correct
- Check file permissions
- Ensure the file has a supported extension (.yaml, .yml, .robogo)

**"unsupported output format"**
- Use only: console, json, or markdown
- Check spelling of output format

**"Still seeing action output in JSON/Markdown mode"**
- This is expected behavior - only runner output is suppressed
- Individual action outputs (sleep, log, etc.) are always displayed
- To capture only the structured output, redirect to file: `./robogo run test.robogo --output json > results.json`

### Debug Mode

For debugging, you can set the log level:

```bash
export ROBOGO_LOG_LEVEL=debug
./robogo run tests/debug-test.robogo
```

### Getting Help

```bash
# General help
./robogo --help

# Command-specific help
./robogo run --help
./robogo list --help
./robogo completions --help

# Version information
./robogo --version
```

## Best Practices

### 1. Use Descriptive Test Names

```yaml
# Good
testcase: "API User Creation Test"
testcase: "Database Connection Validation"

# Avoid
testcase: "Test 1"
testcase: "API Test"
```

### 2. Organize Test Files

```
tests/
├── smoke/
│   ├── basic-functionality.robogo
│   └── health-check.robogo
├── api/
│   ├── user-management.robogo
│   └── authentication.robogo
└── integration/
    ├── end-to-end.robogo
    └── performance.robogo
```

### 3. Use Output Formats Appropriately

```bash
# Interactive development
./robogo run test.robogo

# CI/CD integration
./robogo run test.robogo --output json

# Documentation
./robogo run test.robogo --output markdown > report.md
```

### 4. Handle Errors Gracefully

```bash
# Check exit code
./robogo run test.robogo
if [ $? -eq 0 ]; then
  echo "Test passed"
else
  echo "Test failed"
  exit 1
fi
```

### 5. Use Output Format Appropriately

```bash
# JSON output for CI/CD (runner output suppressed, actions still visible)
./robogo run test.robogo --output json > results.json

# Verbose mode for debugging (full output)
./robogo run test.robogo  # Default console output

# Markdown output for documentation (runner output suppressed)
./robogo run test.robogo --output markdown > report.md
```

## Performance Considerations

### File Size
- Large test files may take longer to parse
- Consider splitting large test suites into smaller files

### Network Requests
- HTTP actions have built-in timeouts
- Use appropriate sleep durations between requests

### Memory Usage
- JSON and Markdown output formats use less memory than console output
- Consider using JSON/Markdown output for large test suites to reduce runner output

For more information, see the [Troubleshooting Guide](troubleshooting.md). 