# Quick Start Guide

Get up and running with Gobot in 5 minutes! This guide will walk you through creating and running your first test case.

## Prerequisites

- Gobot installed (see [Installation Guide](installation.md))
- Basic familiarity with YAML syntax

## Step 1: Create Your First Test Case

Create a file called `hello-world.yaml` with the following content:

```yaml
testcase: "Hello World Test"
description: "A simple test to verify Gobot is working"
steps:
  - keyword: log
    args: ["Hello from Gobot!"]
  - keyword: log
    args: ["Current time is: ${datetime.now()}"]
  - keyword: assert
    args: [true, "This test should always pass"]
```

## Step 2: Run Your Test

Execute the test case using the Gobot CLI:

```bash
gobot run hello-world.yaml
```

You should see output similar to:

```
🚀 Gobot v1.0.0
📋 Running test case: Hello World Test
   📝 Step 1: log - Hello from Gobot!
   📝 Step 2: log - Current time is: 2024-01-15T10:30:00Z
   ✅ Step 3: assert - This test should always pass
✅ Test completed successfully in 0.05s
```

## Step 3: Create a More Complex Test

Let's create a test that demonstrates more features. Create `api-test.yaml`:

```yaml
testcase: "API Health Check"
description: "Test API endpoint health"
steps:
  - keyword: log
    args: ["Starting API health check"]
  
  - keyword: http_request
    args:
      url: "https://httpbin.org/status/200"
      method: "GET"
      timeout: 10
  
  - keyword: assert
    args: ["${response.status_code}", "==", 200]
  
  - keyword: log
    args: ["API health check completed successfully"]
```

Run this test:

```bash
gobot run api-test.yaml
```

## Step 4: Using Environment Variables

Create a test that uses environment variables for configuration:

```yaml
testcase: "Environment Variable Test"
description: "Demonstrate environment variable usage"
steps:
  - keyword: log
    args: ["User: ${USER}"]
  - keyword: log
    args: ["Home directory: ${HOME}"]
  - keyword: assert
    args: ["${USER}", "!=", ""]
```

Set an environment variable and run:

```bash
export CUSTOM_VAR="Hello from environment"
gobot run env-test.yaml
```

## Step 5: Git Integration

Test Gobot's git integration by running a test from a repository:

```bash
# Run a test from a public repository
gobot run --repo https://github.com/your-org/test-examples.git --branch main --path examples/basic
```

## Step 6: Parallel Execution

Create multiple test cases and run them in parallel:

```yaml
# parallel-tests.yaml
testcases:
  - name: "Test A"
    steps:
      - keyword: log
        args: ["Running test A"]
      - keyword: sleep
        args: [1]
      - keyword: assert
        args: [true]
  
  - name: "Test B"
    steps:
      - keyword: log
        args: ["Running test B"]
      - keyword: sleep
        args: [1]
      - keyword: assert
        args: [true]

parallel: true
max_workers: 2
```

Run with parallel execution:

```bash
gobot run parallel-tests.yaml
```

## Step 7: View Test Results

Gobot provides detailed output by default. For different formats:

```bash
# JSON output
gobot run hello-world.yaml --output json

# JUnit XML (for CI/CD integration)
gobot run hello-world.yaml --output junit

# Verbose output with timing
gobot run hello-world.yaml --verbose
```

## Step 8: Configuration

Create a configuration file for your project:

```yaml
# .gobot.yaml
defaults:
  timeout: 30s
  retry_attempts: 3

environments:
  development:
    base_url: "http://localhost:8080"
    log_level: "debug"
  
  production:
    base_url: "https://api.example.com"
    log_level: "info"
```

Run with specific environment:

```bash
gobot run api-test.yaml --env production
```

## What You've Learned

In this quick start, you've:

✅ Created and ran your first test case  
✅ Used built-in keywords (log, assert, http_request)  
✅ Worked with environment variables  
✅ Explored git integration  
✅ Run tests in parallel  
✅ Generated different output formats  
✅ Used configuration files  

## Next Steps

Now that you're familiar with the basics:

1. **Read the [Test Case Writing Guide](test-cases.md)** for best practices
2. **Explore [Built-in Keywords](keywords.md)** for more functionality
3. **Learn about [Git Integration](git-integration.md)** for team workflows
4. **Check out [Secret Management](secrets.md)** for secure credential handling
5. **Review [Parallel Execution](parallel.md)** for performance optimization

## Troubleshooting

### Common Issues

**Test fails with "keyword not found"**
- Check the [Built-in Keywords](keywords.md) reference
- Ensure correct YAML syntax

**HTTP request fails**
- Verify the URL is accessible
- Check network connectivity
- Review timeout settings

**Environment variable not found**
- Ensure the variable is set: `echo $VARIABLE_NAME`
- Use `${VARIABLE_NAME:-default}` for defaults

**Git integration issues**
- Verify repository URL is correct
- Check authentication if using private repos
- Ensure the specified path exists

### Getting Help

- Check the [Troubleshooting Guide](troubleshooting.md)
- Search [GitHub Issues](https://github.com/your-org/gobot/issues)
- Ask in [GitHub Discussions](https://github.com/your-org/gobot/discussions)

## Examples Repository

For more examples and advanced use cases, check out our [examples repository](https://github.com/your-org/gobot-examples) which includes:

- API testing patterns
- Database testing
- UI automation examples
- CI/CD integration samples
- Performance testing scenarios 