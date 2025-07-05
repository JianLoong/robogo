# Troubleshooting Guide

This guide helps you resolve common issues when using Robogo.

## Quick Diagnosis

If you're experiencing issues, start with these diagnostic steps:

1. **Check Robogo version:**
   ```bash
   ./robogo --version
   ```

2. **Verify test file syntax:**
   ```bash
   ./robogo run testcases/hello-world.yaml
   ```

3. **List available actions:**
   ```bash
   ./robogo list
   ```

4. **Check for syntax errors:**
   ```bash
   # Validate YAML syntax
   python3 -c "import yaml; yaml.safe_load(open('your-test.yaml'))"
   ```

## Common Issues and Solutions

### Build and Installation Issues

#### "command not found: robogo"

**Problem:** The Robogo binary is not found in your PATH.

**Solutions:**
```bash
# Build the binary
go build -o robogo cmd/robogo/main.go

# Run from current directory
./robogo --version

# Install to system PATH (optional)
sudo mv robogo /usr/local/bin/
```

#### "go: module github.com/your-org/robogo: not found"

**Problem:** The module name is a placeholder and not published.

**Solutions:**
```bash
# This is expected for local development
# The module name should be updated for your organization
# For now, build locally:
go build -o robogo cmd/robogo/main.go
```

#### Build fails with dependency errors

**Problem:** Missing Go dependencies.

**Solutions:**
```bash
# Download dependencies
go mod download

# Tidy module
go mod tidy

# Verify dependencies
go mod verify
```

### Test Case Issues

#### "failed to parse YAML"

**Problem:** Invalid YAML syntax in test file.

**Common causes and solutions:**

1. **Incorrect indentation:**
   ```yaml
   # ❌ Wrong
   testcase: "Test"
   steps:
   - action: log
   args: ["message"]
   
   # ✅ Correct
   testcase: "Test"
   steps:
     - action: log
       args: ["message"]
   ```

2. **Missing required fields:**
   ```yaml
   # ❌ Wrong - missing testcase name
   description: "Test description"
   steps:
     - action: log
       args: ["message"]
   
   # ✅ Correct
   testcase: "Test Name"
   description: "Test description"
   steps:
     - action: log
       args: ["message"]
   ```

3. **Invalid file extension:**
   ```bash
   # ❌ Wrong - unsupported extension
   ./robogo run test.txt
   
   # ✅ Correct - supported extensions
   ./robogo run test.yaml
   ./robogo run test.yml
   ./robogo run test.robogo
   ```

#### "unknown action: [action_name]"

**Problem:** The action name is not recognized.

**Solutions:**
```bash
# Check available actions
./robogo list

# Verify action name spelling
# Common mistakes:
# - "keyword" instead of "action"
# - "http_request" instead of "http_get"
# - "print" instead of "log"
```

**Correct syntax:**
```yaml
# ✅ Correct
steps:
  - action: log
    args: ["message"]
  - action: http_get
    args: ["https://api.example.com"]
  - action: assert
    args: [true, true, "message"]

# ❌ Wrong
steps:
  - keyword: log
    args: ["message"]
  - action: http_request
    args: ["https://api.example.com"]
```

#### "unsupported file extension"

**Problem:** Test file has an unsupported extension.

**Solutions:**
```bash
# Rename file to supported extension
mv test.txt test.yaml

# Or use supported extensions:
# - .yaml
# - .yml
# - .robogo
```

### HTTP Request Issues

#### "request failed: dial tcp: lookup [hostname]: no such host"

**Problem:** DNS resolution failure or invalid URL.

**Solutions:**
```bash
# Test connectivity
curl -I https://api.example.com

# Check DNS resolution
nslookup api.example.com

# Verify URL format
# ✅ Correct
- action: http_get
  args: ["https://api.example.com/health"]

# ❌ Wrong
- action: http_get
  args: ["api.example.com/health"]  # Missing protocol
```

#### "certificate signed by unknown authority"

**Problem:** SSL certificate validation failure.

**Solutions:**
```yaml
# For development/testing, use mTLS with custom CA
- action: http
  args:
    - "GET"
    - "https://secure.example.com/api"
    - 
      ca: "${CA_CERT_PATH}"
      Authorization: "Bearer ${API_TOKEN}"
```

#### "failed to load client certificate/key"

**Problem:** Invalid or missing mTLS certificates.

**Solutions:**
```bash
# Check certificate files exist
ls -la ${CLIENT_CERT_PATH}
ls -la ${CLIENT_KEY_PATH}

# Verify certificate format
openssl x509 -in ${CLIENT_CERT_PATH} -text -noout

# Check file permissions
chmod 600 ${CLIENT_KEY_PATH}
```

### Variable Substitution Issues

#### "variable not found: ${variable_name}"

**Problem:** Variable is not defined or has incorrect syntax.

**Solutions:**
```yaml
# ✅ Correct - variable is set before use
steps:
  - action: get_time
    args: ["iso"]
    result: timestamp
  - action: log
    args: ["Time: ${timestamp}"]

# ❌ Wrong - variable used before being set
steps:
  - action: log
    args: ["Time: ${timestamp}"]  # timestamp not set yet
  - action: get_time
    args: ["iso"]
    result: timestamp
```

#### Variable substitution not working

**Problem:** Incorrect variable syntax or scope.

**Solutions:**
```yaml
# ✅ Correct syntax
- action: log
  args: ["User: ${user_name}"]
- action: http_get
  args: ["${api_url}/users"]

# ❌ Wrong syntax
- action: log
  args: ["User: $user_name"]  # Missing braces
- action: http_get
  args: ["$api_url/users"]    # Missing braces
```

### Assertion Issues

#### "assertion failed: expected [value1], got [value2]"

**Problem:** Assertion condition is not met.

**Solutions:**
```yaml
# ✅ Correct assertion syntax
- action: assert
  args: [expected_value, actual_value, "descriptive message"]

# Examples:
- action: assert
  args: [200, "${response.status_code}", "API should return 200"]
- action: assert
  args: ["hello", "hello", "String comparison"]
- action: assert
  args: [42, 42, "Number comparison"]
```

#### Type mismatch in assertions

**Problem:** Comparing different data types.

**Solutions:**
```yaml
# ✅ Correct - same types
- action: assert
  args: ["200", "${response.status_code}"]  # Both strings
- action: assert
  args: [200, 200]  # Both numbers

# ❌ Wrong - different types
- action: assert
  args: [200, "${response.status_code}"]  # Number vs string
```

### Performance Issues

#### Tests running slowly

**Problem:** Network timeouts or inefficient test design.

**Solutions:**
```yaml
# Add timeouts to HTTP requests
- action: http
  args:
    - "GET"
    - "https://api.example.com/data"
    - 
      timeout: 30  # 30 second timeout
      Authorization: "Bearer ${API_TOKEN}"

# Use appropriate sleep durations
- action: sleep
  args: [0.5]  # 500ms instead of 5s
```

#### Memory usage issues

**Problem:** Large response bodies or excessive logging.

**Solutions:**
```yaml
# Limit response processing
- action: http_get
  args: ["https://api.example.com/large-data"]
  result: response

# Only log essential information
- action: log
  args: ["Status: ${response.status_code}"]  # Don't log entire response body
```

### Output Format Issues

#### JSON output parsing errors

**Problem:** Invalid JSON output format.

**Solutions:**
```bash
# Validate JSON output
./robogo run test.yaml --output json | jq .

# Check for malformed JSON
./robogo run test.yaml --output json > output.json
python3 -m json.tool output.json
```

#### Markdown output formatting issues

**Problem:** Markdown output is not properly formatted.

**Solutions:**
```bash
# Check markdown syntax
./robogo run test.yaml --output markdown > report.md
markdownlint report.md  # If you have markdownlint installed
```

## Debugging Techniques

### Enable Debug Logging

```bash
# Set debug level
export ROBOGO_LOG_LEVEL=debug

# Run test with verbose output
./robogo run test.yaml
```

### Step-by-Step Debugging

1. **Test individual actions:**
   ```yaml
   testcase: "Debug Test"
   steps:
     - action: log
       args: ["Step 1: Basic logging"]
     - action: get_time
       args: ["iso"]
       result: debug_time
     - action: log
       args: ["Step 2: Time is ${debug_time}"]
   ```

2. **Isolate problematic steps:**
   ```yaml
   # Comment out problematic steps
   testcase: "Isolation Test"
   steps:
     - action: log
       args: ["This works"]
     # - action: http_get
     #   args: ["https://problematic-url.com"]
     - action: log
       args: ["This also works"]
   ```

### Network Debugging

```bash
# Test network connectivity
curl -v https://api.example.com/health

# Check DNS resolution
dig api.example.com

# Test with different HTTP clients
wget -O- https://api.example.com/health
```

## Environment-Specific Issues

### Docker Issues

**Problem:** Tests fail in Docker containers.

**Solutions:**
```dockerfile
# Use multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o robogo cmd/robogo/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/robogo /usr/local/bin/
ENTRYPOINT ["robogo"]
```

### CI/CD Issues

**Problem:** Tests pass locally but fail in CI/CD.

**Solutions:**
```yaml
# GitHub Actions example
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go build -o robogo cmd/robogo/main.go
      - run: ./robogo run testcases/ --output json > results.json
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: results.json
```

### Windows Issues

**Problem:** Path separators or line endings cause issues.

**Solutions:**
```bash
# Use forward slashes in paths
- action: http
  args:
    - "GET"
    - "https://api.example.com"
    - 
      cert: "certs/client.crt"  # Use forward slashes
      key: "certs/client.key"

# Convert line endings
dos2unix test.yaml
```

## Getting Help

### Before Asking for Help

1. **Check this troubleshooting guide**
2. **Search existing issues** on GitHub
3. **Try the diagnostic steps** above
4. **Gather relevant information:**
   - Robogo version: `./robogo --version`
   - Go version: `go version`
   - Operating system: `uname -a` (Linux/macOS) or `systeminfo` (Windows)
   - Test case content (sanitized)
   - Error messages and stack traces

### Reporting Issues

When reporting an issue, include:

1. **Environment details:**
   - Operating system and version
   - Go version
   - Robogo version

2. **Reproduction steps:**
   - Exact commands run
   - Test case content
   - Expected vs actual behavior

3. **Error information:**
   - Complete error messages
   - Stack traces (if available)
   - Logs with debug level enabled

4. **Additional context:**
   - Is this a regression?
   - Does it work in other environments?
   - Any recent changes that might have caused this?

### Community Resources

- **GitHub Issues:** Report bugs and request features
- **GitHub Discussions:** Ask questions and share experiences
- **Documentation:** Check the [main documentation](README.md)
- **Examples:** Review the [examples directory](../examples/)

## Prevention Best Practices

### 1. Use Version Control

```bash
# Track test cases in git
git add testcases/
git commit -m "Add new test cases"

# Use branches for experimental changes
git checkout -b feature/new-tests
```

### 2. Implement Continuous Testing

```bash
# Run tests regularly
./robogo run testcases/ --output json > results.json

# Monitor for regressions
diff previous-results.json results.json
```

### 3. Use Descriptive Test Names

```yaml
# ✅ Good
testcase: "API User Creation with Valid Data"
testcase: "Database Connection with SSL"

# ❌ Avoid
testcase: "Test 1"
testcase: "API Test"
```

### 4. Implement Proper Error Handling

```yaml
# Use assertions with clear messages
- action: assert
  args: ["${response.status_code}", "==", 200, "API should return 200 OK"]

# Log important information
- action: log
  args: ["API response status: ${response.status_code}"]
```

### 5. Regular Maintenance

```bash
# Update dependencies
go mod tidy
go mod download

# Run tests regularly
./robogo run testcases/

# Review and update documentation
# Keep test cases up to date with API changes
```

For additional help, see the [main documentation](README.md) or [GitHub issues](https://github.com/your-org/robogo/issues). 