# Test Directory Structure

This directory contains all test cases for the Robogo project, organized by domain and feature for clarity and maintainability.

## Structure

- **templates/**
  - **swift/**: SWIFT message template tests (MT103, MT202, MT900, MT910, etc.)
  - **sepa/**: SEPA (pain.001, etc.) template tests
- **integration/**: Integration tests for external systems (Kafka, RabbitMQ, HTTP, Postgres, parallel DB, etc.)
- **core/**: Core engine and language feature tests (assert, control flow, variables, random, secrets, time, etc.)
- **tdm/**: Test Data Management (TDM) tests
- **edge/**: Edge cases, error handling, minimal, silent, syntax, and verbosity tests

## How to Run

To run a specific test:

```
go run cmd/robogo/main.go run <path-to-test-file>
```

For example:

```
go run cmd/robogo/main.go run tests/templates/swift/test-template-mt103.robogo
```

## Test Categories

### templates/
- Contains tests for file-based message templates (SWIFT, SEPA, etc.)

### integration/
- Tests for integration with external systems and services (Kafka, RabbitMQ, HTTP, Postgres, parallel DB, etc.)

### core/
- Tests for core language features, built-in actions, and utilities

### tdm/
- Tests for Test Data Management features

### edge/
- Tests for error handling, edge cases, minimal and silent runs, syntax, and verbosity

---

Feel free to add new tests in the appropriate subdirectory! 