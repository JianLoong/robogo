# Robogo Test Examples

This directory contains example `.robogo` test files that demonstrate the features and syntax of the Robogo test automation framework.

## Files

- **test-variables.robogo**
  - Demonstrates variable assignment, variable substitution, and using variables in assertions and actions.
  - Shows how to use actions like `get_time`, `get_random`, `concat`, and `length`.

- **test-time-formats.robogo**
  - Demonstrates the `get_time` action with all available time format options, including predefined and custom Go time formats.

- **test-syntax.robogo**
  - Provides a comprehensive syntax showcase for Robogo, including all supported step fields, actions, and YAML features.

- **test-http.robogo**
  - Demonstrates HTTP actions with various methods, headers, and certificate options.
  - Shows both file path and PEM content for certificates.

## Usage

You can run any of these test files with the Robogo CLI:

```sh
robogo run examples/tests/test-variables.robogo
robogo run examples/tests/test-time-formats.robogo
robogo run examples/tests/test-syntax.robogo
robogo run examples/tests/test-http.robogo
```

You can also run all tests in this directory with a simple script or loop.

## Creating Your Own Test Files

1. **Copy an existing example** as a starting point.
2. Use the `name` field to give each step a descriptive name (strongly recommended).
3. Use the `action` field to specify what to do in each step.
4. Use the `result` field to store the output of an action in a variable.
5. Reference variables in later steps using `${variable}` syntax.
6. Use built-in actions like `log`, `sleep`, `assert`, `get_time`, `get_random`, `concat`, `length`, and HTTP actions.
7. Save your file with a `.robogo` extension.

### Example
```yaml
steps:
  - name: "Get current timestamp"
    action: get_time
    args: ["iso"]
    result: now
  - name: "Log the current time"
    action: log
    args: ["The current time is ${now}"]
```

## Step Fields

Each step can have the following fields:

- **`name`** (optional but strongly recommended): A descriptive name for the step that appears in logs and reports.
- **`action`** (required): The action to execute (e.g., `log`, `http_get`, `assert`).
- **`args`** (required): Arguments to pass to the action.
- **`result`** (optional): Variable name to store the action's output.

### Step Naming Best Practices

- **Use descriptive names** that explain what the step does (e.g., "Login to API", "Verify response status").
- **Be consistent** with naming conventions across your test cases.
- **Keep names concise** but informative.
- **Use action verbs** to describe the step's purpose.

## Best Practices

- **Use descriptive step names** for clarity and better reporting.
- **Use descriptive variable names** for clarity.
- **Group related steps** with comments for readability.
- **Use assertions** to validate expected outcomes.
- **Leverage variables** to avoid repeating logic or values.
- **Use the VS Code extension** for syntax highlighting, autocompletion, and documentation.
- **Keep test cases focused**: one test case per file is recommended for clarity and maintainability.

## More Information

- See the main project README for a full list of available actions and advanced features.
- For custom actions or advanced usage, refer to the documentation or source code.

Feel free to use these files as templates for your own tests or to explore the capabilities of Robogo! 