# Quick Start Guide

Get up and running with Robogo in 5 minutes! This guide will walk you through creating and running your first test case.

## Prerequisites

- Robogo installed (see [Installation Guide](installation.md))
- Basic familiarity with YAML syntax

## Step 1: Create Your First Test Case

Create a file called `hello-world.yaml` with the following content:

```yaml
testcase: "Hello World Test"
description: "A simple test to verify Robogo is working"
steps:
  - action: log
    args: ["Hello from Robogo!"]
  - action: log
    args: ["This is a simple test case"]
  - action: sleep
    args: [1]
  - action: assert
    args: [true, true, "This should always pass"]
  - action: log
    args: ["Test completed successfully!"]
```

## Step 2: Run Your Test

Execute the test case using the Robogo CLI:

```bash
./robogo run hello-world.yaml
```

You should see output similar to:

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

## Step 3: Create a More Complex Test

Let's create a test that demonstrates more features. Create `api-test.yaml`:

```yaml
testcase: "API Health Check"
description: "Test API endpoint health"
steps:
  - action: log
    args: ["Starting API health check"]
  
  - action: http_get
    args: ["https://httpbin.org/status/200"]
    result: response
  
  - action: assert
    args: ["${response.status_code}", "==", 200]
  
  - action: log
    args: ["API health check completed successfully"]
```

Run this test:

```bash
./robogo run api-test.yaml
```

## Step 4: Using Variables

Create a test that uses variables for storing and referencing data:

```yaml
testcase: "Variable Test"
description: "Demonstrate variable usage"
steps:
  - action: get_time
    args: ["iso"]
    result: timestamp
  
  - action: log
    args: ["Current timestamp: ${timestamp}"]
  
  - action: get_random
    args: [100]
    result: random_number
  
  - action: log
    args: ["Random number: ${random_number}"]
  
  - action: assert
    args: ["${random_number}", ">=", 0]
```

Run this test:

```bash
./robogo run variable-test.yaml
```

## Step 5: String Operations

Create a test that demonstrates string manipulation:

```yaml
testcase: "String Operations Test"
description: "Demonstrate string operations"
steps:
  - action: concat
    args: ["Hello", " ", "World", "!"]
    result: greeting
  
  - action: log
    args: ["Greeting: ${greeting}"]
  
  - action: length
    args: ["${greeting}"]
    result: greeting_length
  
  - action: log
    args: ["Greeting length: ${greeting_length}"]
  
  - action: assert
    args: ["${greeting_length}", "==", 12]
```

## Step 6: HTTP Operations with mTLS

Create a test that demonstrates HTTP operations with mutual TLS:

```yaml
testcase: "HTTP mTLS Test"
description: "Test HTTP requests with mutual TLS"
steps:
  - action: log
    args: ["Starting HTTP mTLS test"]
  
  - action: http
    args: 
      - "GET"
      - "https://secure.example.com/api/health"
      - 
        cert: "${CLIENT_CERT_PATH}"
        key: "${CLIENT_KEY_PATH}"
        ca: "${CA_CERT_PATH}"
        Authorization: "Bearer ${API_TOKEN}"
    result: response
  
  - action: log
    args: ["Response status: ${response.status_code}"]
  
  - action: assert
    args: ["${response.status_code}", "==", 200]
```

## Step 7: View Test Results in Different Formats

Robogo provides multiple output formats:

```bash
# JSON output
./robogo run hello-world.yaml --output json

# Markdown output
./robogo run hello-world.yaml --output markdown

# Console output (default)
./robogo run hello-world.yaml --output console
```

## Step 8: List Available Actions

Explore what actions are available:

```bash
# List all actions
./robogo list

# Get completions for autocomplete
./robogo completions http
```

## Step 9: Run Multiple Test Cases

Create a test suite with multiple test cases:

```yaml
# test-suite.yaml
testcase: "Basic Functionality Test"
description: "Test basic Robogo functionality with various actions"
steps:
  - action: log
    args: ["Starting basic functionality test"]
  
  - action: sleep
    args: [0.5]
  
  - action: assert
    args: ["hello", "hello", "String comparison should pass"]
  
  - action: assert
    args: [42, 42, "Number comparison should pass"]
  
  - action: log
    args: ["All basic assertions passed!"]
  
  - action: sleep
    args: [0.5]
  
  - action: log
    args: ["Basic functionality test completed"]
```

## What You've Learned

In this quick start, you've:

✅ Created and ran your first test case  
✅ Used built-in actions (log, sleep, assert, http_get)  
✅ Worked with variables and result storage  
✅ Explored string operations  
✅ Made HTTP requests with mTLS support  
✅ Generated different output formats  
✅ Listed available actions  

## Next Steps

Now that you're familiar with the basics:

1. **Read the [Test Case Writing Guide](test-cases.md)** for best practices
2. **Explore [Built-in Actions](actions.md)** for more functionality
3. **Check out the [examples](../examples/)** directory for more sample test cases
4. **Read the [Contributing Guide](../CONTRIBUTING.md)** if you want to contribute

## Troubleshooting

### Common Issues

**"unknown action" error**
- Check the action name spelling
- Use `./robogo list` to see available actions
- Ensure you're using the correct syntax

**HTTP request failures**
- Check your network connection
- Verify the URL is accessible
- For mTLS, ensure certificate paths are correct

**Variable substitution issues**
- Use `${variable_name}` syntax
- Ensure variables are set before use
- Check for typos in variable names

For more help, see the [Troubleshooting Guide](troubleshooting.md). 