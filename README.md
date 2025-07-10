# Robogo

**Modern, extensible test automation for APIs, databases, message queues, and more—built in Go.**

---

## 🏁 Shift Left & Transparency

Robogo is designed to help teams **shift left** in their testing strategy by giving developers full transparency and control over test cases. All tests are written in clear, version-controlled YAML, making it easy for anyone on the team to:
- 🧐 Understand exactly what is being tested
- ✍️ Review, modify, and extend test coverage early in the development lifecycle
- 🔄 Integrate tests directly into CI/CD pipelines for rapid feedback

This approach reduces bugs, increases confidence, and ensures quality is built in from the start.

---

## ❓ What is Robogo?

Robogo is a powerful, developer-friendly test automation framework written in Go. It lets you define, run, and report on complex end-to-end tests for APIs, databases, message queues (Kafka, RabbitMQ), SWIFT/SEPA messages, and more—all using a simple YAML-based DSL.

Robogo is designed for:
- 👩‍💻 API and microservice teams
- 💸 Financial services and payment systems
- 🛠️ Data engineering and integration testing
- 🚦 CI/CD pipelines and regression testing

---

## ✨ Features

- 🌐 **API Testing:** HTTP(S) with mTLS, custom headers, and full response validation
- 🗄️ **Database Testing:** PostgreSQL and Google Cloud Spanner support
- 📬 **Message Queues:** Native Kafka and RabbitMQ publish/consume actions
- 🗃️ **Test Data Management:** Structured data sets, environments, and lifecycle
- 💶 **SWIFT/SEPA Message Generation:** Template-based financial message support
- ⚡ **Parallel Execution:** Run tests and steps concurrently with dependency analysis
- 🔁 **Retry & Control Flow:** If, for, while, and robust retry logic
- 🔐 **Secrets Management:** File-based secrets with output masking
- 🖨️ **Multiple Output Formats:** Console, JSON, Markdown
- 🧩 **VS Code Extension:** Syntax highlighting, autocomplete, validation, and one-click execution
- 🧱 **Extensible:** Add your own actions and templates
- 📊 **Comprehensive Reporting:** Always-on summary, step-level analytics, and error introspection

---

## 🚀 Quick Start

### Prerequisites

- 🦫 Go 1.22+ (required)
- 🐳 Docker (for Kafka, RabbitMQ, Postgres, Spanner emulators)
- 💻 (Optional) VS Code for enhanced editing

### Installation

> **Note:**
> - On **Windows**, the built binary will be `robogo.exe` and you should run it as `./robogo.exe`.
> - On **Linux/macOS**, the binary will be `robogo` and you should run it as `./robogo`.

```bash
# Option 1: Build the binary (recommended for repeated use)
git clone https://github.com/JianLoong/robogo.git
cd robogo
go mod download
go build -o robogo ./cmd/robogo
```

```bash
# Option 2: Run directly without building (good for quick tests)
go run cmd/robogo/main.go run hello.robogo
```

### First Test Example

Create a file `hello.robogo`:

```yaml
testcase: "Hello World"
description: "A minimal Robogo test"
steps:
  - name: "Log Hello"
    action: log
    args: ["Hello, Robogo!"]
```

Run it:

```bash
# If you built the binary:
#   On Windows:
./robogo.exe run hello.robogo
#   On Linux/macOS:
./robogo run hello.robogo

# Or run directly without building (any OS):
go run cmd/robogo/main.go run hello.robogo
```

---

## 🧩 Examples: API, Kafka, Database, SWIFT

### 🌐 HTTP API Test

```yaml
testcase: "API Test"
steps:
  - name: "GET request"
    action: http
    args: ["GET", "https://httpbin.org/get"]
    result: response
  - name: "Assert status"
    action: assert
    args: ["${response.status_code}", "==", "200"]
  - name: "Log body"
    action: log
    args: ["Body: ${response.body}"]
```

### 📬 Kafka Publish/Consume

```yaml
testcase: "Kafka Integration"
variables:
  vars:
    kafka_broker: "localhost:9092"
    topic: "robogo_test"
    message: "hello from robogo"
steps:
  - name: "Publish message"
    action: kafka
    args: ["publish", "${kafka_broker}", "${topic}", "${message}"]
  - name: "Consume message"
    action: kafka
    args: ["consume", "${kafka_broker}", "${topic}", {fromOffset: "first", count: 1, timeout: 5}]
    result: consumed
  - name: "Assert consumed"
    action: assert
    args: ["${consumed}", "contains", "${message}"]
```

### 🗄️ PostgreSQL Query

```yaml
testcase: "Postgres Query"
variables:
  secrets:
    db_password:
      file: "db_secret.txt"
      mask_output: true
steps:
  - name: "Query DB"
    action: postgres
    args: ["query", "postgres://user:${db_password}@localhost/db", "SELECT 1"]
    result: query_result
  - name: "Assert rows"
    action: assert
    args: ["${query_result.rows_affected}", "==", "1"]
```

### 💶 SWIFT Message Generation

```yaml
testcase: "SWIFT MT103 Generation"
variables:
  vars:
    sender_bic: "DEUTDEFF"
    sender_account: "1234567890"
    sender_name: "Sender Company"
    beneficiary_account: "0987654321"
    beneficiary_name: "Beneficiary Company"
    currency: "EUR"
    amount: "1000.00"
    reference: "INV-2024-001"
steps:
  - name: "Generate transaction timestamp"
    action: get_time
    args: ["unix_ms"]
    result: timestamp_ms

  - name: "Create unique transaction ID"
    action: concat
    args: ["TXN", "${timestamp_ms}"]
    result: transaction_id

  - name: "Get current date"
    action: get_time
    args: ["date"]
    result: current_date

  - name: "Generate MT103 SWIFT Message"
    action: template
    args:
      - "templates/mt103.tmpl"
      -
        Sender:
          BIC: "${sender_bic}"
          Account: "${sender_account}"
          Name: "${sender_name}"
        Beneficiary:
          Account: "${beneficiary_account}"
          Name: "${beneficiary_name}"
        TransactionID: "${transaction_id}"
        Timestamp: "${timestamp_ms}"
        Date: "${current_date}"
        Currency: "${currency}"
        Amount: "${amount}"
        Reference: "${reference}"
    result: mt103_message

  - name: "Log Swift Message"
    action: log
    args: ["Generated SWIFT MT103 message:\n${mt103_message}"]
```

---

## ⚡ Parallelism & Configuration

- **Suite-level:**
  ```yaml
  testsuite: "Parallel Suite"
  parallel: true
  max_concurrency: 4
  testcases:
    - file: ./test-api.robogo
    - file: ./test-kafka.robogo
  ```
- **Step-level:**
  ```yaml
  parallelism:
    enabled: true
    max_concurrency: 4
    steps: true
  ```

- **Output formats:**
  - 🖨️ Console (default), 🗂️ JSON, 📝 Markdown  
    `./robogo run test.robogo --output json`

- **Secrets:**
  ```yaml
  variables:
    secrets:
      api_key:
        file: "secret.txt"
        mask_output: true
  ```

---

## 🧠 Advanced Topics

- 🔁 **Retry Logic:**
  ```yaml
  retry:
    attempts: 3
    delay: "1s"
    backoff: "exponential"
  ```
- 🔄 **Loops & Data-driven:**
  ```yaml
  - for:
      condition: "1..5"
      steps:
        - name: "Random value"
          action: get_random
          args: [100]
          result: rand
        - name: "Log random"
          action: log
          args: ["Random: ${rand}"]
  ```
- 🏷️ **Templates:**
  Use Go templates for SWIFT/SEPA or custom payloads.

- 🪪 **Reserved Variables:**
  `__robogo_steps` gives you access to all previous step results and errors.

---

## 🗂️ Project Structure

```
robogo/
├── cmd/robogo/          # CLI entry point
├── internal/            # Core engine: actions, parser, runner, util
├── tests/               # Test cases (core, integration, templates, tdm, edge)
├── examples/            # Example .robogo files and suites
├── templates/           # SWIFT/SEPA and custom templates
├── .vscode/             # VS Code extension
└── docker-compose.yml   # Local dev services (Kafka, RabbitMQ, Postgres, Spanner)
```

---

## 🛠️ Development & Contribution

- 🏗️ Build: `go build -o robogo ./cmd/robogo`
- 🧪 Run all Go tests: `go test ./...`
- 📄 See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for code style and PR guidelines.

---

## 🆘 Troubleshooting & FAQ

- ❓ **Summary not printing?** Always prints in latest version, even on error/panic.
- 🌀 **Parallel deadlocks?** All goroutines send results, even on panic. Check for infinite loops.
- 🚨 **Error handling?** Steps fail on error unless `expect_error` is set and error matches.
- 📚 **Need more examples?** See the `examples/` and `tests/` directories.

---

## 📄 License

MIT License. See LICENSE file for details.

---

**Robogo**: Modern, robust, and scalable test automation for Go, with first-class support for APIs, databases, message queues, and financial messaging.
