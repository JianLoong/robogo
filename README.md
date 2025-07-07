# Robogo - Modern Test Automation Framework

A modern, git-driven test automation framework written in Go, designed for comprehensive API testing, SWIFT message generation, database operations, and Test Data Management (TDM).

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

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or later
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

### VS Code Extension Setup

The project includes a complete VS Code extension for enhanced development:

```bash
# Quick setup - run the extension launcher
./run-extension.ps1

# Or manually build and install
cd .vscode/extensions/robogo
npm install
npm run compile
```

### Run Your First Test

```bash
# Run a basic test
./robogo.exe run examples/sample.robogo

# Run SWIFT message testing
./robogo.exe run tests/templates/swift/test-swift-working.robogo

# Run Test Data Management
./robogo.exe run tests/tdm/test-tdm-simple.robogo

# Run decimal random testing
./robogo.exe run tests/core/test-random-decimals.robogo

# Run a test suite
./robogo.exe run-suite examples/test-suite.robogo
```

## 📋 Available Actions

### Basic Operations
- **`log`** - Output messages to console with verbosity control
- **`sleep`** - Pause execution for specified duration
- **`assert`** - Verify conditions with comparison operators (==, !=, >, <, >=, <=, contains, starts_with, ends_with)

### Time and Random
- **`get_time`** - Get current timestamp (iso, datetime, date, time, unix, unix_ms, custom formats)
- **`get_random`** - Generate random numbers (integers and decimals with precision control)

### String Operations
- **`concat`** - Concatenate multiple strings
- **`length`** - Get length of strings or arrays

### HTTP Operations
- **`http`** - Generic HTTP requests with mTLS support and custom options
- **`http_get`** - Simplified GET requests
- **`http_post`** - Simplified POST requests

### Database Operations
- **`postgres`** - PostgreSQL operations (query, execute, connect, close)

### Control Flow
- **`control`** - Conditional execution and loop control
- **`if`** - Conditional execution with then/else blocks
- **`for`** - Loop execution (ranges, arrays, counts)
- **`while`** - Conditional loops with max iteration limits

### Test Data Management
- **`tdm`** - Test Data Management operations (generate, validate, load_dataset, set_environment)
- **`variable`** - Variable management operations (set_variable, get_variable, list_variables)

### Template Operations
- **`template`** - Generate content from templates with variable substitution

## 📊 Test Data Management (TDM)

Robogo includes a comprehensive Test Data Management system for structured data handling:

```yaml
testcase: "TDM Example"
description: "Demonstrate Test Data Management features"

# Environment configuration
environments:
  - name: "development"
    description: "Development environment"
    variables:
      api_base_url: "https://dev-api.example.com"
      timeout: 30
    overrides:
      debug_mode: true

# Test Data Management configuration
data_management:
  environment: "development"
  isolation: true
  cleanup: true
  
  # Structured data sets
  data_sets:
    - name: "test_users"
      description: "Test user data"
      version: "1.0"
      data:
        user1:
          name: "John Doe"
          email: "john@example.com"
          age: 30
        user2:
          name: "Jane Smith"
          email: "jane@example.com"
          age: 25
      schema:
        name: "string"
        email: "email"
        age: "number"
      required: ["name", "email"]
      unique: ["email"]

  # Data validation
  validation:
    - name: "email_validation"
      type: "format"
      field: "test_users.user1.email"
      rule: "email"
      message: "User email must be valid"
      severity: "error"

  # Setup and teardown
  setup:
    - name: "TDM Setup"
      action: log
      args: ["Setting up test environment"]

  teardown:
    - name: "TDM Cleanup"
      action: log
      args: ["Cleaning up test environment"]

steps:
  # Use TDM data
  - name: "Log user data"
    action: log
    args: ["User: ${test_users.user1.name} (${test_users.user1.email})"]
  
  # Database operations with TDM data
  - name: "Insert user"
    action: postgres
    args: ["execute", "postgres://user:pass@localhost/db", "INSERT INTO users (name, email) VALUES ($1, $2)", ["${test_users.user1.name}", "${test_users.user1.email}"]]
```

## 🏦 SWIFT Message Testing

Robogo excels at SWIFT message generation and testing:

```yaml
testcase: "SWIFT Message Test"
description: "Generate and test SWIFT messages"

variables:
  vars:
    bank_bic: "DEUTDEFF"
    currency: "EUR"
    test_amount: "1000.00"
  secrets:
    swift_api_key:
      file: "secret.txt"
      mask_output: true

steps:
  # Generate unique transaction ID
  - action: get_time
    args: ["unix_ms"]
    result: timestamp_ms
  
  - action: concat
    args: ["TXN", "${timestamp_ms}"]
    result: transaction_id

  # Generate SWIFT MT103 message
  - action: concat
    args: [
      "{1:F01", "${bank_bic}", "XXXX", "U", "3003", "1234567890", "}",
      "{2:I103", "${bank_bic}", "XXXX", "U}",
      "{3:{113:SEPA}",
      "{108:${transaction_id}}",
      "{111:001}",
      "{121:${timestamp_ms}}}",
      "{4:",
      ":20:${transaction_id}",
      ":23B:CRED",
      ":32A:${current_date}${currency}${test_amount}",
      ":33B:${currency}${test_amount}",
      ":50K:/1234567890",
      "1/Account Name",
      ":59:/0987654321",
      "1/Beneficiary Name",
      ":70:INV-2024-001",
      ":71A:SHA",
      "-}",
      "{5:{CHK:1234567890ABCD}{TNG:}}{S:{COP:S}}"
    ]
    result: swift_message

  # Test via HTTP API
  - action: http_post
    args: 
      - "https://api.swift.com/v1/messages"
      - '{"message": "${swift_message}", "type": "MT103"}'
    result: api_response

  # Validate response
  - action: assert
    args: ["${api_response.status_code}", "==", "200", "API should return 200"]
```

## 💳 Payment Payload Templating

Robogo provides powerful templating capabilities for generating payment payloads in various formats including SWIFT messages, SEPA XML, and custom payment formats.

### SWIFT Message Templates

Generate SWIFT messages using file-based templates:

```yaml
testcase: "SWIFT Template Test"
description: "Generate SWIFT messages using templates"

variables:
  vars:
    bank_bic: "DEUTDEFF"
    currency: "EUR"
    test_amount: "5000.00"
    beneficiary_name: "Acme Corporation"
    beneficiary_bic: "COBADEFF"
    beneficiary_account: "0987654321"
    reference: "INV-2024-001"
    sender_name: "Sender Company Ltd"

steps:
  # Generate transaction data
  - action: get_time
    args: ["unix_ms"]
    result: timestamp_ms
  
  - action: concat
    args: ["TXN", "${timestamp_ms}"]
    result: transaction_id
  
  - action: get_time
    args: ["date"]
    result: current_date

  # Generate MT103 message from template
  - action: template
    args:
      - "templates/mt103.tmpl"
      -
        TransactionID: "${transaction_id}"
        Amount: "${test_amount}"
        Currency: "${currency}"
        Sender:
          BIC: "${bank_bic}"
          Account: "1234567890"
          Name: "${sender_name}"
        Beneficiary:
          BIC: "${beneficiary_bic}"
          Account: "${beneficiary_account}"
          Name: "${beneficiary_name}"
        Reference: "${reference}"
        Timestamp: "${timestamp_ms}"
        Date: "${current_date}"
    result: swift_message

  # Validate generated message
  - action: assert
    args: ["${swift_message}", "contains", "{1:F01", "Message must contain SWIFT header"]
  
  - action: assert
    args: ["${swift_message}", "contains", ":20:${transaction_id}", "Message must contain transaction reference"]
```

### SEPA XML Templates

Generate SEPA Credit Transfer XML messages:

```yaml
testcase: "SEPA Template Test"
description: "Generate SEPA Credit Transfer XML using templates"

steps:
  # Generate SEPA Credit Transfer XML
  - action: template
    args:
      - "templates/sepa-credit-transfer.xml.tmpl"
      -
        MessageID: "MSG123456"
        CreationDateTime: "2025-01-15T12:00:00"
        NumberOfTransactions: "1"
        ControlSum: "1000.00"
        InitiatingPartyName: "My Company"
        PaymentInfoID: "PMTINF123"
        RequestedExecutionDate: "2025-01-16"
        DebtorName: "John Doe"
        DebtorIBAN: "DE89370400440532013000"
        DebtorBIC: "DEUTDEFF"
        EndToEndID: "E2E123456"
        Amount: "1000.00"
        Currency: "EUR"
        CreditorBIC: "COBADEFF"
        CreditorName: "Jane Smith"
        CreditorIBAN: "DE75512108001245126199"
        RemittanceInfo: "Invoice 2024-001"
    result: sepa_xml

  # Validate SEPA XML structure
  - action: assert
    args: ["${sepa_xml}", "contains", "<Document", "SEPA XML must start with <Document>"]
  
  - action: assert
    args: ["${sepa_xml}", "contains", "<MsgId>MSG123456</MsgId>", "SEPA XML must contain correct Message ID"]
  
  - action: assert
    args: ["${sepa_xml}", "contains", "<InstdAmt Ccy=\"EUR\">1000.00</InstdAmt>", "SEPA XML must contain correct amount"]
```

### Custom Payment Templates

Create custom payment payloads for various payment systems:

```yaml
testcase: "Custom Payment Template Test"
description: "Generate custom payment payloads using templates"

templates:
  payment_request: |
    {
      "payment_id": "{{.PaymentID}}",
      "amount": {{.Amount}},
      "currency": "{{.Currency}}",
      "sender": {
        "account": "{{.SenderAccount}}",
        "name": "{{.SenderName}}",
        "bic": "{{.SenderBIC}}"
      },
      "recipient": {
        "account": "{{.RecipientAccount}}",
        "name": "{{.RecipientName}}",
        "bic": "{{.RecipientBIC}}"
      },
      "reference": "{{.Reference}}",
      "execution_date": "{{.ExecutionDate}}",
      "priority": "{{.Priority}}"
    }

steps:
  # Generate payment data
  - action: get_time
    args: ["unix_ms"]
    result: payment_id
  
  - action: get_time
    args: ["date"]
    result: execution_date

  # Generate custom payment payload
  - action: template
    args:
      - "payment_request"
      -
        PaymentID: "${payment_id}"
        Amount: "2500.75"
        Currency: "EUR"
        SenderAccount: "DE89370400440532013000"
        SenderName: "Sender Company"
        SenderBIC: "DEUTDEFF"
        RecipientAccount: "DE75512108001245126199"
        RecipientName: "Recipient Company"
        RecipientBIC: "COBADEFF"
        Reference: "INV-2024-002"
        ExecutionDate: "${execution_date}"
        Priority: "normal"
    result: payment_payload

  # Send payment request
  - action: http_post
    args:
      - "https://api.payment.com/v1/transfers"
      - "${payment_payload}"
      -
        Content-Type: "application/json"
        Authorization: "Bearer ${API_TOKEN}"
    result: payment_response

  # Validate response
  - action: assert
    args: ["${payment_response.status_code}", "==", "200", "Payment should be processed successfully"]
```

### Template Features

Robogo's templating system supports:

- **File-based Templates**: Load templates from external files
- **Inline Templates**: Define templates within test cases
- **Variable Substitution**: Use dynamic data in templates
- **Nested Objects**: Support for complex data structures
- **Conditional Logic**: Use Go template conditionals
- **Loops**: Iterate over arrays and collections
- **Functions**: Access to Go template functions

### Available Payment Templates

The project includes pre-built templates for common payment formats:

- **`templates/mt103.tmpl`** - SWIFT MT103 (Customer Transfer)
- **`templates/mt202.tmpl`** - SWIFT MT202 (General Financial Institution Transfer)
- **`templates/mt900.tmpl`** - SWIFT MT900 (Confirmation of Debit)
- **`templates/mt910.tmpl`** - SWIFT MT910 (Confirmation of Credit)
- **`templates/sepa-credit-transfer.xml.tmpl`** - SEPA Credit Transfer XML (pain.001)

### Template Usage Examples

```yaml
# Basic template usage
- action: template
  args:
    - "templates/mt103.tmpl"
    -
      TransactionID: "TXN123"
      Amount: "1000.00"
      Currency: "EUR"
      Sender:
        BIC: "DEUTDEFF"
        Account: "1234567890"
        Name: "Sender Bank"
      Beneficiary:
        BIC: "COBADEFF"
        Account: "0987654321"
        Name: "Beneficiary Bank"
      Reference: "INV-001"
      Timestamp: "1705312800000"
      Date: "20250115"
  result: swift_message

# Template with dynamic data
- action: template
  args:
    - "templates/sepa-credit-transfer.xml.tmpl"
    -
      MessageID: "${message_id}"
      CreationDateTime: "${creation_time}"
      NumberOfTransactions: "${tx_count}"
      ControlSum: "${total_amount}"
      InitiatingPartyName: "${company_name}"
      PaymentInfoID: "${payment_info_id}"
      RequestedExecutionDate: "${execution_date}"
      DebtorName: "${debtor_name}"
      DebtorIBAN: "${debtor_iban}"
      DebtorBIC: "${debtor_bic}"
      EndToEndID: "${end_to_end_id}"
      Amount: "${amount}"
      Currency: "${currency}"
      CreditorBIC: "${creditor_bic}"
      CreditorName: "${creditor_name}"
      CreditorIBAN: "${creditor_iban}"
      RemittanceInfo: "${reference}"
  result: sepa_xml
```

## 🎲 Enhanced Random Generation

Support for both integer and decimal random values with precision control:

```yaml
# Integer random (backward compatible)
- action: get_random
  args: [100]
  result: int_random

# Decimal random (new feature)
- action: get_random
  args: [100.5]
  result: decimal_random

# SWIFT amount generation
- action: get_random
  args: [50000.00]
  result: swift_amount

# Multiple random values in loop
- for:
    condition: "1..5"
    steps:
      - action: get_random
        args: [1000.25]
        result: iteration_amount
      
      - action: log
        args: ["Amount ${iteration}: ${iteration_amount}"]
```

## 🌐 HTTP API Testing

Comprehensive HTTP testing with mTLS support:

```yaml
# Simple GET request
- action: http_get
  args: ["https://api.example.com/users"]
  result: response

# POST with JSON body
- action: http_post
  args: 
    - "https://api.example.com/users"
    - '{"name": "John", "email": "john@example.com"}'
  result: create_response

# mTLS request with certificates
- action: http
  args: 
    - "POST"
    - "https://secure.example.com/api"
    - '{"secure": true}'
    - 
      Content-Type: "application/json"
      Authorization: "Bearer ${API_TOKEN}"
    - 
      cert: "${CLIENT_CERT_PATH}"
      key: "${CLIENT_KEY_PATH}"
      ca: "${CA_CERT_PATH}"
  result: secure_response
```

## 🚀 Parallel Execution

Robogo supports parallel execution at multiple levels for improved performance:

### Test File Parallelism

Run multiple test files concurrently with configurable concurrency limits:

```bash
# Run multiple test files in parallel
./robogo.exe run tests/test-http.robogo tests/test-postgres.robogo --parallel

# Limit concurrency to 2 test files at a time
./robogo.exe run tests/*.robogo --parallel --max-concurrency 2
```

### Test Suite Execution

Run test suites with shared setup and teardown:

```bash
# Run a test suite
./robogo.exe run-suite examples/test-suite.robogo

# Run multiple suites
./robogo.exe run-suite examples/test-suite.robogo examples/multi-outcome-suite.robogo
```

### Step-Level Parallelism

Execute independent steps within a test case in parallel:

```yaml
testcase: "Parallel Step Execution"
description: "Demonstrate parallel step execution with dependency analysis"

parallelism:
  enabled: true
  max_concurrency: 4
  steps: true  # Enable parallel step execution

steps:
  # Independent steps that can run in parallel
  - name: "Get current time"
    action: get_time
    args: ["unix"]
    result: timestamp1
  
  - name: "Generate random number"
    action: get_random
    args: [1000]
    result: random1
  
  - name: "Get another timestamp"
    action: get_time
    args: ["unix"]
    result: timestamp2
  
  # Dependent step (waits for timestamp1 and random1)
  - name: "Use previous results"
    action: concat
    args: ["${timestamp1}", "-", "${random1}"]
    result: combined_result
    depends_on: ["timestamp1", "random1"]
```

## 💾 Database Operations

Robogo provides comprehensive database testing capabilities with support for PostgreSQL and Google Cloud Spanner:

### PostgreSQL Operations

PostgreSQL integration with secure credential management:

```yaml
variables:
  secrets:
    db_password:
      file: "db-secret.txt"
      mask_output: true

steps:
  # Build connection string
  - action: concat
    args: ["postgres://user:", "${db_password}", "@localhost/db"]
    result: db_connection

  # Execute query
  - action: postgres
    args: ["query", "${db_connection}", "SELECT * FROM users"]
    result: query_result

  # Validate results
  - action: assert
    args: ["${query_result.rows_affected}", ">", "0", "Should return results"]
```

### Google Cloud Spanner Operations

Cloud-native distributed database operations with emulator support:

```yaml
testcase: "Spanner Operations Test"
description: "Test Google Cloud Spanner operations with emulator"

variables:
  vars:
    spanner_project: "robogo-test-project"
    spanner_instance: "robogo-test-instance"
    spanner_database: "robogo-test-db"

steps:
  # Connect to Spanner emulator
  - action: spanner
    args: ["connect", "projects/${spanner_project}/instances/${spanner_instance}/databases/${spanner_database}?useEmulator=true"]
    result: connection_result

  # Create table
  - action: spanner
    args: ["execute", "projects/${spanner_project}/instances/${spanner_instance}/databases/${spanner_database}",
           "CREATE TABLE IF NOT EXISTS users (id STRING(MAX) NOT NULL, name STRING(MAX), email STRING(MAX)) PRIMARY KEY (id)"]
    result: create_result

  # Insert data
  - action: spanner
    args: ["execute", "projects/${spanner_project}/instances/${spanner_instance}/databases/${spanner_database}",
           "INSERT INTO users (id, name, email) VALUES (@id, @name, @email)", ["user1", "John Doe", "john@example.com"]]
    result: insert_result

  # Query data
  - action: spanner
    args: ["query", "projects/${spanner_project}/instances/${spanner_instance}/databases/${spanner_database}",
           "SELECT * FROM users WHERE id = @id", ["user1"]]
    result: query_result

  # Validate results
  - action: assert
    args: ["${query_result}", "contains", "John Doe", "User name should be John Doe"]

  # Close connection
  - action: spanner
    args: ["close"]
    result: close_result
```

#### Spanner Features
- **Emulator Support**: Local development with Google Cloud Spanner emulator
- **Distributed Transactions**: Handle complex distributed database operations
- **SQL Dialect**: Full support for Spanner's SQL dialect
- **Parameterized Queries**: Secure query execution with parameter binding
- **Connection Management**: Automatic connection pooling and cleanup

## 📨 Kafka Operations

Robogo provides comprehensive Kafka support for message publishing and consumption with configurable producer and consumer settings:

### Kafka Publishing

Publish messages to Kafka topics with various configuration options:

```yaml
testcase: "Kafka Publisher Test"
description: "Test Kafka message publishing with different configurations"

variables:
  vars:
    kafka_broker: "localhost:9092"
    topic_name: "payment-events"
    message_body: "Payment processed: TXN123456"

steps:
  # Publish with default settings
  - action: kafka
    args:
      - "publish"
      - "${kafka_broker}"
      - "${topic_name}"
      - "${message_body}"
    result: publish_result

  # Validate publish result
  - action: assert
    args: ["${publish_result.status}", "==", "message published"]

  # Publish with custom settings
  - action: kafka
    args:
      - "publish"
      - "${kafka_broker}"
      - "${topic_name}"
      - "High-priority payment: TXN789012"
      -
        acks: "all"
        compression: "snappy"
    result: publish_result_high_priority

  # Validate high-priority publish
  - action: assert
    args: ["${publish_result_high_priority.status}", "==", "message published"]
```

### Kafka Consumption

Consume messages from Kafka topics with consumer group support:

```yaml
testcase: "Kafka Consumer Test"
description: "Test Kafka message consumption with consumer groups"

variables:
  vars:
    kafka_broker: "localhost:9092"
    topic_name: "payment-events"
    consumer_group: "payment-processor-group"

steps:
  # Consume message with consumer group
  - action: kafka
    args:
      - "consume"
      - "${kafka_broker}"
      - "${topic_name}"
      -
        groupID: "${consumer_group}"
        fromOffset: "first"
    result: consume_result

  # Validate consumed message
  - action: assert
    args: ["${consume_result.message}", "contains", "Payment processed", "Should contain payment message"]
  
  - action: assert
    args: ["${consume_result.topic}", "==", "${topic_name}", "Should be from correct topic"]

  # Consume from specific partition
  - action: kafka
    args:
      - "consume"
      - "${kafka_broker}"
      - "${topic_name}"
      -
        partition: "0"
        fromOffset: "last"
    result: partition_consume_result
```

### Kafka Configuration Options

Robogo supports various Kafka configuration options:

#### Producer Settings
- **`acks`**: Message acknowledgment level (`"all"`, `"one"`, `"none"`)
- **`compression`**: Message compression (`"gzip"`, `"snappy"`, `"lz4"`, `"zstd"`)

#### Consumer Settings
- **`groupID`**: Consumer group identifier for load balancing
- **`fromOffset`**: Starting offset (`"first"`, `"last"`)
- **`partition`**: Specific partition to consume from

### Kafka Integration Testing

Test end-to-end Kafka workflows:

```yaml
testcase: "Kafka Integration Test"
description: "Test complete Kafka publish-consume workflow"

variables:
  vars:
    kafka_broker: "localhost:9092"
    topic_name: "test-topic"
    consumer_group: "test-group"
    test_message: "Integration test message from Robogo"

steps:
  # Publish test message
  - action: kafka
    args:
      - "publish"
      - "${kafka_broker}"
      - "${topic_name}"
      - "${test_message}"
      -
        acks: "all"
        compression: "snappy"
    result: publish_result

  # Validate publish success
  - action: assert
    args: ["${publish_result.status}", "==", "message published"]

  # Consume the message
  - action: kafka
    args:
      - "consume"
      - "${kafka_broker}"
      - "${topic_name}"
      -
        groupID: "${consumer_group}"
        fromOffset: "first"
    result: consume_result

  # Validate message content
  - action: assert
    args: ["${consume_result.message}", "==", "${test_message}", "Consumed message should match published message"]

  # Validate metadata
  - action: assert
    args: ["${consume_result.topic}", "==", "${topic_name}", "Topic should match"]
  
  - action: assert
    args: ["${consume_result.partition}", ">=", "0", "Partition should be valid"]
  
  - action: assert
    args: ["${consume_result.offset}", ">=", "0", "Offset should be valid"]
```

### Kafka with Docker

The project includes Docker Compose configuration for local development with Kafka, PostgreSQL, RabbitMQ, and Google Cloud Spanner:

```yaml
# docker-compose.yml
version: '3.8'
services:
  zookeeper:
    image: bitnami/zookeeper:3.8
    container_name: zookeeper
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes
    ports:
      - "2181:2181"

  kafka:
    image: bitnami/kafka:3.6
    container_name: kafka
    depends_on:
      - zookeeper
    environment:
      - KAFKA_BROKER_ID=1
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
    ports:
      - "9092:9092"

  spanner:
    image: gcr.io/cloud-spanner-emulator/emulator:latest
    container_name: spanner-emulator
    ports:
      - "9010:9010"
      - "9020:9020"
    environment:
      - SPANNER_PROJECT_ID=robogo-test-project
      - SPANNER_INSTANCE_ID=robogo-test-instance
      - SPANNER_DATABASE_ID=robogo-test-db
    command: >
      gcloud emulators spanner start
      --host-port=0.0.0.0:9010
      --rest-port=9020
      --project=${SPANNER_PROJECT_ID}

  postgres:
    image: postgres:15
    container_name: postgres
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=robogo_testdb
      - POSTGRES_USER=robogo_testuser
      - POSTGRES_PASSWORD=robogo_testpass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  rabbitmq:
    image: rabbitmq:3-management
    container_name: rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      - RABBITMQ_DEFAULT_USER=robogo_user
      - RABBITMQ_DEFAULT_PASS=robogo_pass

volumes:
  postgres_data:
```

Start all services for testing:
```bash
docker-compose up -d
```

## 📝 Test Case Format

Test cases are written in YAML format with comprehensive features:

```yaml
testcase: "Comprehensive Test"
description: "Test with variables, secrets, control flow, and TDM"

variables:
  vars:
    api_url: "https://api.example.com"
    timeout: 30
  secrets:
    api_key:
      file: "secret.txt"
      mask_output: true

steps:
  # Control flow with loops and retry
  - for:
      condition: "1..3"
      steps:
        - action: get_random
          args: [1000.50]
          result: amount
        
        - action: http_post
          args: 
            - "${api_url}/transactions"
            - '{"amount": "${amount}"}'
          result: response
          retry:
            attempts: 3
            delay: "1s"
            backoff: "exponential"
        
        - action: assert
          args: ["${response.status_code}", "==", "200"]

  # Conditional execution
  - if:
      condition: "${response.status_code} == 200"
      then:
        - action: log
          args: ["Transaction successful"]
      else:
        - action: log
          args: ["Transaction failed"]
```

## 🧪 Example Test Cases

### Core Functionality
- **`examples/sample.robogo`** - Basic functionality demonstration
- **`tests/core/test-syntax.robogo`** - Syntax and basic operations
- **`tests/core/test-variables.robogo`** - Variable management and substitution
- **`tests/core/test-assert.robogo`** - Assertion and validation examples

### Advanced Features
- **`tests/tdm/test-tdm-simple.robogo`** - Simple Test Data Management
- **`tests/tdm/test-tdm.robogo`** - Comprehensive TDM with PostgreSQL integration
- **`tests/core/test-control-flow.robogo`** - Control flow features (if, for, while)
- **`tests/core/test-retry.robogo`** - Retry mechanisms and error handling

### SWIFT and Financial
- **`tests/templates/swift/test-swift-working.robogo`** - SWIFT message generation and testing
- **`tests/templates/swift/test-swift-messages.robogo`** - Advanced SWIFT message examples
- **`tests/templates/swift/test-swift-advanced.robogo`** - Complex SWIFT workflows

### API and Database Testing
- **`tests/integration/test-http.robogo`** - HTTP API testing examples
- **`tests/integration/test-postgres.robogo`** - Database operations and queries
- **`tests/core/test-secrets.robogo`** - Secret management and security
- **`tests/integration/test-parallel.robogo`** - Parallel execution and batch operations

### Random Generation and Utilities
- **`tests/core/test-random-decimals.robogo`** - Enhanced random number generation
- **`tests/core/test-random-ranges.robogo`** - Random value ranges and validation
- **`tests/core/test-random-edge-cases.robogo`** - Edge cases and boundary testing
- **`tests/core/test-time-formats.robogo`** - Time formatting and manipulation

### Error Handling and Validation
- **`tests/edge/test-fail-in-loop.robogo`** - Error handling in loops
- **`tests/edge/test-continue-on-failure.robogo`** - Continue on failure scenarios
- **`tests/edge/test-verbosity.robogo`** - Verbosity levels and logging

### Test Suites
- **`examples/test-suite.robogo`** - Basic test suite with setup/teardown
- **`examples/multi-outcome-suite.robogo`** - Suite with mixed test outcomes
- **`examples/step-results-suite.robogo`** - Suite demonstrating step-level reporting

## 🏗️ Project Structure

```
robogo/
├── cmd/robogo/          # CLI entry point
├── internal/           # Core framework code
│   ├── actions/        # Built-in actions (HTTP, DB, TDM, etc.)
│   ├── parser/         # YAML parsing and test execution
│   ├── runner/         # Test orchestration and TDM integration
│   └── util/           # Utility functions and validation
├── tests/              # Comprehensive test examples
│   ├── core/           # Core functionality tests
│   ├── edge/           # Edge case and error handling tests
│   ├── integration/    # Integration and end-to-end tests
│   ├── tdm/            # Test Data Management tests
│   └── templates/      # Template-based tests (SWIFT, SEPA)
├── examples/           # Basic examples and tutorials
├── docs/              # Documentation and guides
├── prd/               # Product requirements and specifications
├── templates/         # Template files for SWIFT and SEPA messages
├── .vscode/           # VS Code extension and configuration
│   └── extensions/robogo/  # Complete VS Code extension
└── run-extension.ps1  # Extension launcher script
```

## 🛠️ VS Code Extension

Robogo includes a complete VS Code extension for enhanced development experience:

### Features
- **🎯 Rich Documentation & Hover Support** - Comprehensive action documentation with detailed parameters, examples, and best practices
- **🔍 Intelligent Autocomplete** - Smart suggestions for actions, fields, parameters, and TDM structures
- **✅ Real-time Validation** - Syntax validation, action verification, and TDM structure checking
- **🎨 Enhanced Syntax Highlighting** - Color-coded highlighting for actions, control flow, TDM elements, and variables
- **🚀 Test Execution** - One-click test execution with integrated terminal output
- **📊 Step-Level Reporting** - Detailed step-by-step results with timing and error information

### Installation & Setup

#### Quick Setup
```bash
# Run the extension launcher (Windows PowerShell)
./run-extension.ps1
```

#### Manual Setup
```bash
# Navigate to extension directory
cd .vscode/extensions/robogo

# Install dependencies
npm install

# Build the extension
npm run compile

# Launch VS Code with extension
code --new-window --extensionDevelopmentPath="path/to/robogo/.vscode/extensions/robogo"
```

### Extension Commands
- **`robogo.runTest`** - Run the current Robogo test file
- **`robogo.runTestParallel`** - Run test with parallel execution
- **`robogo.listActions`** - Display all available actions with documentation
- **`robogo.validateTDM`** - Validate TDM configuration
- **`robogo.generateTemplate`** - Generate test template
- **`robogo.runWithOutput`** - Run test with specific output format

### Configuration
The extension provides several configuration options:
- **`robogo.executablePath`** - Path to Robogo executable
- **`robogo.outputFormat`** - Default output format (console, json, markdown)
- **`robogo.showDetailedDocumentation`** - Enable detailed hover tooltips
- **`robogo.enableParallelExecution`** - Enable parallel execution by default
- **`robogo.maxConcurrency`** - Maximum concurrency for parallel execution
- **`robogo.enableRealTimeValidation`** - Enable real-time validation
- **`robogo.showVerboseOutput`** - Show verbose output in test execution

## 🔧 Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run specific test
./robogo.exe run tests/templates/swift/test-swift-working.robogo

# Run TDM test
./robogo.exe run tests/tdm/test-tdm-simple.robogo

# Run with specific output format
./robogo.exe run test.robogo --output json

# Run tests in parallel
./robogo.exe run tests/*.robogo --parallel --max-concurrency 4

# Run test suite
./robogo.exe run-suite examples/test-suite.robogo

# Run suite with markdown output
./robogo.exe run-suite examples/test-suite.robogo --output markdown
```

### Build

```bash
go build -o robogo.exe ./cmd/robogo
```

### List Available Actions

```bash
./robogo.exe list
```

### Get Action Completions

```bash
./robogo.exe completions get_random
```

## 📊 Output Formats

Robogo supports multiple output formats with detailed analytics and step-level reporting:

### Console Output (Default)
Human-readable output with colors, formatting, and detailed step information:
```bash
./robogo.exe run test.robogo
```

### JSON Output
Machine-readable format for CI/CD integration with complete step details:
```bash
./robogo.exe run test.robogo --output json
```

### Markdown Output
Documentation-friendly format with collapsible step details:
```bash
./robogo.exe run test.robogo --output markdown
```

### Test Suite Output
Comprehensive suite reporting with step-level summaries:
```bash
./robogo.exe run-suite examples/test-suite.robogo --output markdown
```

## 📚 Documentation

Comprehensive documentation available in the [docs/](docs/) directory:

- **[TDM Implementation Guide](docs/tdm-implementation.md)** - Complete Test Data Management system documentation
- **[TDM Evaluation Summary](docs/tdm-evaluation-summary.md)** - TDM system analysis and evaluation
- **[Framework Comparison](docs/framework-comparison.md)** - Robogo vs Robot Framework, Selenium, Postman, and others
- **[Actions Reference](docs/actions.md)** - Complete list of available actions with examples
- **[Quick Start Guide](docs/quickstart.md)** - Get started quickly with Robogo
- **[Test Cases Guide](docs/test-cases.md)** - Writing effective test cases
- **[CLI Reference](docs/cli-reference.md)** - Command-line interface documentation
- **[Contributing Guide](docs/CONTRIBUTING.md)** - How to contribute to Robogo

## 🎯 Use Cases

### Financial Services
- **SWIFT Message Testing** - Generate and validate SWIFT messages (MT103, MT202, etc.)
- **Payment API Testing** - Test payment processing systems and workflows
- **Banking Integration** - Validate banking APIs, compliance, and regulatory requirements
- **Test Data Management** - Structured data sets for financial testing scenarios

### API Testing
- **REST API Validation** - Comprehensive HTTP API testing with authentication
- **mTLS Security Testing** - Test secure API endpoints with certificate validation
- **Performance Testing** - Load testing, retry mechanisms, and performance validation
- **Data-Driven Testing** - TDM-powered test scenarios with multiple data sets

### Database Testing
- **PostgreSQL Operations** - Database query, transaction, and integration testing
- **Data Validation** - Verify database state, results, and data integrity
- **Integration Testing** - End-to-end database workflows with TDM data sets
- **Data Lifecycle Management** - Setup, teardown, and cleanup operations

### Test Automation
- **CI/CD Integration** - Automated testing in continuous integration pipelines
- **Regression Testing** - Comprehensive test suites with TDM data management
- **Load and Performance** - Scalable testing with retry mechanisms and timing
- **Cross-Platform Testing** - Consistent testing across different environments

## 🚀 Roadmap

### Completed Features ✅
- [x] **Test Data Management (TDM)** - Structured data sets and lifecycle management
- [x] **Enhanced Random Generation** - Decimal support with precision control
- [x] **Comprehensive HTTP Testing** - mTLS, headers, and response validation
- [x] **PostgreSQL Integration** - Database operations with connection pooling
- [x] **Google Cloud Spanner** - Cloud-native distributed database operations with emulator support
- [x] **Advanced Control Flow** - If, for, while loops with retry mechanisms
- [x] **Secret Management** - Secure credential handling with masking
- [x] **VS Code Extension** - Complete extension with syntax highlighting, autocomplete, and validation
- [x] **Parallel Execution** - Test file and step-level parallelism with dependency analysis
- [x] **Batch Operations** - Parallel HTTP requests and database operations
- [x] **Step-Level Reporting** - Detailed step-by-step results in all output formats
- [x] **Test Suite Support** - Suite execution with shared setup/teardown and comprehensive reporting

### Planned Features 🚧
- [ ] **Plugin System** - Custom action development and extensibility
- [ ] **Web Interface** - Browser-based test management and monitoring
- [ ] **Advanced Reporting** - Detailed analytics, dashboards, and metrics
- [ ] **Cloud Integration** - AWS, Azure, GCP support and cloud-native testing
- [ ] **CI/CD Integration** - Jenkins, GitHub Actions, GitLab CI templates
- [ ] **Multi-Database Support** - MySQL, SQLite, MongoDB integration
- [ ] **GraphQL Testing** - Native GraphQL query and mutation testing

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details on:

- Code style and standards
- Testing requirements
- Pull request process
- Issue reporting
- Feature requests

## 📄 License

MIT License - see LICENSE file for details.

---

**Robogo** - Modern test automation for the Go ecosystem, with powerful SWIFT message generation, comprehensive API testing, advanced Test Data Management capabilities, parallel execution for high-performance testing, and a complete VS Code extension for enhanced development experience. Built for financial services, API testing, and enterprise automation needs. 

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