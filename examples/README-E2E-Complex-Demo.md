# E2E Complex Demo Suite

This comprehensive end-to-end test suite demonstrates the advanced capabilities of Robogo, showcasing real-world testing scenarios with multiple technologies and complex workflows.

## Overview

The E2E Complex Demo Suite consists of a main test suite file and five individual test cases that cover different aspects of modern application testing:

- **HTTP API Testing** - RESTful API operations, authentication, and response validation
- **Database Testing** - PostgreSQL CRUD operations, transactions, and complex queries
- **Template/String Operations** - SWIFT message generation, SEPA transfers, and data transformation
- **Messaging/Data Flow** - Message publishing, consumption, routing, and error handling
- **Control Flow** - Advanced control structures, loops, conditionals, and error handling

## Files Structure

```
examples/
├── e2e-complex-demo.robogo          # Main test suite
├── e2e-http-api-test.robogo         # HTTP API comprehensive test
├── e2e-database-test.robogo         # Database operations test
├── e2e-template-test.robogo         # Template and string operations test
├── e2e-messaging-test.robogo        # Messaging and data flow test
├── e2e-control-flow-test.robogo     # Control flow operations test
└── README-E2E-Complex-Demo.md       # This documentation
```

## Features Demonstrated

### 1. Test Suite Management
- **Parallel/Sequential Execution** - Configurable test execution modes
- **Data Management** - TDM (Test Data Management) with data sets and validation
- **Variable Management** - Regular and secret variables with file-based storage
- **Setup/Teardown** - Comprehensive initialization and cleanup procedures

### 2. HTTP API Testing
- **RESTful Operations** - GET, POST, PUT, DELETE with proper status codes
- **Authentication** - Bearer token authentication
- **Response Validation** - Status codes, content validation, and error handling
- **Batch Operations** - Multiple API calls in sequence
- **Error Scenarios** - Invalid endpoints and error response validation

### 3. Database Testing
- **Connection Management** - Database connectivity and health checks
- **CRUD Operations** - Create, Read, Update, Delete operations
- **Complex Queries** - Aggregations, joins, and data validation
- **Transactions** - Commit and rollback scenarios
- **Data Cleanup** - Proper test data management and cleanup

### 4. Template and String Operations
- **SWIFT Message Generation** - MT103, MT202 message formats
- **SEPA Transfer XML** - Complete SEPA credit transfer document
- **String Manipulation** - Concatenation, length operations, and formatting
- **Data Transformation** - Complex message building with variables
- **Error Handling** - Invalid operations and edge cases

### 5. Messaging and Data Flow
- **Message Publishing** - Single and batch message operations
- **Message Consumption** - Retrieving and processing messages
- **Data Transformation** - Message routing and transformation pipelines
- **Error Handling** - Invalid topics and recovery mechanisms
- **Monitoring** - Performance metrics and health checks

### 6. Control Flow Operations
- **Conditionals** - If/else, nested conditions, and switch statements
- **Loops** - For, while, for-each with break and continue
- **Error Handling** - Try-catch blocks and error recovery
- **Parallel Execution** - Concurrent task execution
- **Performance Testing** - Loop performance and timeout controls

## Configuration

### Required Variables

The test suite uses the following variables that should be configured:

```yaml
variables:
  regular:
    api_base_url: "https://jsonplaceholder.typicode.com"  # Test API endpoint
    database_url: "postgres://testuser:testpass@localhost:5432/testdb"
    kafka_brokers: "localhost:9092"
    rabbitmq_url: "amqp://localhost"
    swift_bank_bic: "DEUTDEFF"
    swift_currency: "EUR"
    test_amount: "1000.00"
    test_user_count: 5
  secrets:
    api_key:
      file: "secret.txt"
      mask_output: true
    database_password:
      file: "db_secret.txt"
      mask_output: true
```

### Required Secret Files

Create the following secret files in your project root:

- `secret.txt` - API key for HTTP operations
- `db_secret.txt` - Database password
- `kafka_secret.txt` - Kafka API key (if using Kafka)
- `rabbitmq_secret.txt` - RabbitMQ password (if using RabbitMQ)

## Running the Tests

### Prerequisites

1. **Database Setup** - PostgreSQL instance running with test database
2. **API Endpoints** - Test API endpoints available (or mock services)
3. **Secret Files** - All required secret files created with proper values
4. **Robogo Installation** - Robogo framework installed and configured

### Execution Commands

```bash
# Run the entire test suite
robogo run examples/e2e-complex-demo.robogo

# Run individual test cases
robogo run examples/e2e-http-api-test.robogo
robogo run examples/e2e-database-test.robogo
robogo run examples/e2e-template-test.robogo
robogo run examples/e2e-messaging-test.robogo
robogo run examples/e2e-control-flow-test.robogo

# Run with specific options
robogo run examples/e2e-complex-demo.robogo --parallel --max-concurrency 5
```

### Expected Output

The test suite will produce detailed output including:

- Test execution progress with emojis and timestamps
- Step-by-step validation results
- Comprehensive test summaries
- Performance metrics and statistics
- Error handling and recovery information

## Key Learning Points

### 1. Test Organization
- **Modular Design** - Separate test cases for different concerns
- **Reusable Components** - Common setup and teardown procedures
- **Data Isolation** - Proper test data management and cleanup

### 2. Error Handling
- **Graceful Failures** - Tests continue even when some operations fail
- **Recovery Mechanisms** - Automatic cleanup and error recovery
- **Validation** - Comprehensive assertion and validation checks

### 3. Performance Considerations
- **Parallel Execution** - Configurable concurrency for faster execution
- **Resource Management** - Proper cleanup of databases, queues, and files
- **Monitoring** - Performance metrics and health checks

### 4. Security Best Practices
- **Secret Management** - File-based secrets with output masking
- **Authentication** - Proper API authentication and authorization
- **Data Protection** - Secure handling of sensitive information

## Customization

### Adding New Test Cases

1. Create a new `.robogo` file in the `examples/` directory
2. Follow the naming convention: `e2e-{feature}-test.robogo`
3. Add the test case to the main suite's `testcases` section
4. Update this README with documentation

### Modifying Existing Tests

1. **Variables** - Update variable values for your environment
2. **Endpoints** - Change API endpoints to match your services
3. **Data Sets** - Modify test data to match your requirements
4. **Assertions** - Adjust validation criteria as needed

### Environment-Specific Configuration

Create environment-specific variable files:

```yaml
# development.robogo
variables:
  regular:
    api_base_url: "http://localhost:8080"
    database_url: "postgres://dev:dev@localhost:5432/devdb"

# production.robogo
variables:
  regular:
    api_base_url: "https://api.production.com"
    database_url: "postgres://prod:prod@prod-db:5432/proddb"
```

## Troubleshooting

### Common Issues

1. **Database Connection Failures**
   - Verify PostgreSQL is running
   - Check connection string format
   - Ensure database exists and user has permissions

2. **API Endpoint Errors**
   - Verify API endpoints are accessible
   - Check authentication credentials
   - Validate request/response formats

3. **Secret File Issues**
   - Ensure secret files exist and are readable
   - Check file permissions
   - Verify secret values are correct

4. **Template Generation Errors**
   - Check template file existence
   - Validate JSON data format
   - Ensure all required variables are set

### Debug Mode

Run tests with debug output:

```bash
robogo run examples/e2e-complex-demo.robogo --verbose --debug
```

## Contributing

When adding new features to this demo suite:

1. **Follow Naming Conventions** - Use consistent file and variable naming
2. **Add Documentation** - Update this README with new features
3. **Include Error Handling** - Always handle potential failure scenarios
4. **Test Thoroughly** - Ensure all test cases pass before committing
5. **Update Examples** - Keep examples current with framework changes

## Conclusion

This E2E Complex Demo Suite provides a comprehensive example of how to use Robogo for real-world testing scenarios. It demonstrates best practices for test organization, error handling, performance optimization, and security considerations.

Use this suite as a reference for building your own comprehensive test automation solutions with Robogo. 