import * as vscode from 'vscode';

/**
 * Action definition interface
 */
export interface ActionDefinition {
    name: string;
    description: string;
    documentation: string;
    snippet: string;
    parameters: ActionParameter[];
    examples: string[];
    category: string;
}

/**
 * Action parameter definition
 */
export interface ActionParameter {
    name: string;
    type: string;
    description: string;
    required: boolean;
    default?: any;
    examples?: string[];
}

/**
 * Registry of all Robogo actions with their definitions
 */
export class RobogoActionRegistry {
    private actions: Map<string, ActionDefinition> = new Map();

    constructor() {
        this.initializeActions();
    }

    /**
     * Get all actions
     */
    getAllActions(): ActionDefinition[] {
        return Array.from(this.actions.values());
    }

    /**
     * Check if action exists
     */
    hasAction(name: string): boolean {
        return this.actions.has(name);
    }

    /**
     * Get action definition
     */
    getAction(name: string): ActionDefinition | undefined {
        return this.actions.get(name);
    }

    /**
     * Get action documentation
     */
    getActionDocumentation(name: string): string | undefined {
        const action = this.actions.get(name);
        return action?.documentation;
    }

    /**
     * Get actions by category
     */
    getActionsByCategory(category: string): ActionDefinition[] {
        return Array.from(this.actions.values()).filter(action => action.category === category);
    }

    /**
     * Initialize all action definitions
     */
    private initializeActions(): void {
        // HTTP Actions
        this.addAction({
            name: 'http',
            description: 'Execute HTTP requests with full feature support',
            documentation: `
**HTTP Action**

Execute HTTP requests with comprehensive features including mTLS, custom headers, and response handling.

**Supported Methods:** GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

**Parameters:**
- \`method\`: HTTP method (required)
- \`url\`: Target URL (required)  
- \`body\`: Request body (optional)
- \`headers\`: Custom headers (optional)
- \`timeout\`: Request timeout (optional)

**Examples:**
\`\`\`yaml
# Simple GET request
- name: "Get user data"
  action: http
  args: ["GET", "https://api.example.com/users/123"]
  result: user_response

# POST with body and headers
- name: "Create user"
  action: http
  args: ["POST", "https://api.example.com/users", "{\\"name\\": \\"John\\"}"]
  options:
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer \${api_token}"
  result: create_response
\`\`\`
            `,
            snippet: 'action: http\n  args: ["${1:GET}", "${2:https://api.example.com}"]${3:\n  result: ${4:response}}',
            parameters: [
                { name: 'method', type: 'string', description: 'HTTP method', required: true, examples: ['GET', 'POST', 'PUT', 'DELETE'] },
                { name: 'url', type: 'string', description: 'Target URL', required: true },
                { name: 'body', type: 'string', description: 'Request body', required: false },
            ],
            examples: [
                'action: http\n  args: ["GET", "https://api.example.com/users"]',
                'action: http\n  args: ["POST", "https://api.example.com/users", "{\\"name\\": \\"John\\"}"]'
            ],
            category: 'HTTP'
        });

        // Database Actions
        this.addAction({
            name: 'postgres',
            description: 'Execute PostgreSQL database operations',
            documentation: `
**PostgreSQL Action**

Execute PostgreSQL database operations with connection management and parameterized queries.

**Subcommands:**
- \`connect\`: Establish database connection
- \`query\`: Execute SELECT queries
- \`execute\`: Execute INSERT/UPDATE/DELETE
- \`close\`: Close connection

**Parameters:**
- \`subcommand\`: Operation to perform (required)
- \`connection_string\`: Database connection string (for connect)
- \`query\`: SQL query to execute (for query/execute)
- \`connection_name\`: Named connection reference (optional)

**Examples:**
\`\`\`yaml
# Connect to database
- name: "Connect to database"
  action: postgres
  args: ["connect", "postgres://user:pass@localhost/db", "main"]

# Execute query
- name: "Get users"
  action: postgres
  args: ["query", "SELECT * FROM users WHERE id = $1", "123", "main"]
  result: users
\`\`\`
            `,
            snippet: 'action: postgres\n  args: ["${1:query}", "${2:SELECT * FROM table}"]${3:\n  result: ${4:result}}',
            parameters: [
                { name: 'subcommand', type: 'string', description: 'Database operation', required: true, examples: ['connect', 'query', 'execute', 'close'] },
                { name: 'query', type: 'string', description: 'SQL query', required: false },
            ],
            examples: [
                'action: postgres\n  args: ["query", "SELECT * FROM users"]',
                'action: postgres\n  args: ["connect", "postgres://user:pass@localhost/db"]'
            ],
            category: 'Database'
        });

        this.addAction({
            name: 'spanner',
            description: 'Execute Google Cloud Spanner operations',
            documentation: `
**Google Cloud Spanner Action**

Execute operations against Google Cloud Spanner database with connection management.

**Subcommands:**
- \`connect\`: Establish Spanner connection
- \`query\`: Execute SELECT queries  
- \`execute\`: Execute DML operations
- \`close\`: Close connection

**Parameters:**
- \`subcommand\`: Operation to perform (required)
- \`connection_string\`: Spanner connection string (for connect)
- \`query\`: SQL query to execute (for query/execute)
- \`connection_name\`: Named connection reference (optional)

**Examples:**
\`\`\`yaml
# Connect to Spanner
- name: "Connect to Spanner"
  action: spanner  
  args: ["connect", "projects/my-project/instances/my-instance/databases/my-db", "main"]

# Execute query
- name: "Query data"
  action: spanner
  args: ["query", "SELECT * FROM Users WHERE UserId = @userId", "123", "main"]
  result: spanner_result
\`\`\`
            `,
            snippet: 'action: spanner\n  args: ["${1:query}", "${2:SELECT * FROM table}"]${3:\n  result: ${4:result}}',
            parameters: [
                { name: 'subcommand', type: 'string', description: 'Spanner operation', required: true, examples: ['connect', 'query', 'execute', 'close'] },
                { name: 'query', type: 'string', description: 'SQL query', required: false },
            ],
            examples: [
                'action: spanner\n  args: ["query", "SELECT * FROM Users"]',
                'action: spanner\n  args: ["connect", "projects/proj/instances/inst/databases/db"]'
            ],
            category: 'Database'
        });

        // Messaging Actions
        this.addAction({
            name: 'kafka',
            description: 'Kafka message publish/consume operations',
            documentation: `
**Kafka Action**

Publish and consume messages from Apache Kafka topics.

**Subcommands:**
- \`publish\`: Publish message to topic
- \`consume\`: Consume message from topic
- \`connect\`: Establish Kafka connection

**Parameters:**
- \`subcommand\`: Kafka operation (required)
- \`broker\`: Kafka broker address (for connect)
- \`topic\`: Topic name (for publish/consume)
- \`message\`: Message content (for publish)
- \`timeout\`: Consumer timeout (optional)

**Examples:**
\`\`\`yaml
# Publish message
- name: "Publish to Kafka"
  action: kafka
  args: ["publish", "localhost:9092", "test-topic", "Hello World"]

# Consume message  
- name: "Consume from Kafka"
  action: kafka
  args: ["consume", "localhost:9092", "test-topic", "30s"]
  result: kafka_message
\`\`\`
            `,
            snippet: 'action: kafka\n  args: ["${1:publish}", "${2:localhost:9092}", "${3:topic}", "${4:message}"]${5:\n  result: ${6:result}}',
            parameters: [
                { name: 'subcommand', type: 'string', description: 'Kafka operation', required: true, examples: ['publish', 'consume', 'connect'] },
                { name: 'broker', type: 'string', description: 'Kafka broker', required: true },
                { name: 'topic', type: 'string', description: 'Topic name', required: true },
            ],
            examples: [
                'action: kafka\n  args: ["publish", "localhost:9092", "test-topic", "message"]',
                'action: kafka\n  args: ["consume", "localhost:9092", "test-topic"]'
            ],
            category: 'Messaging'
        });

        this.addAction({
            name: 'rabbitmq',
            description: 'RabbitMQ message operations',
            documentation: `
**RabbitMQ Action**

Manage RabbitMQ connections and message operations.

**Subcommands:**
- \`connect\`: Establish RabbitMQ connection
- \`publish\`: Publish message to exchange
- \`consume\`: Consume message from queue
- \`close\`: Close connection

**Parameters:**
- \`subcommand\`: RabbitMQ operation (required)
- \`connection_string\`: AMQP connection string (for connect)
- \`exchange\`: Exchange name (for publish)
- \`routing_key\`: Routing key (for publish)
- \`queue\`: Queue name (for consume)
- \`message\`: Message body (for publish)

**Examples:**
\`\`\`yaml
# Connect to RabbitMQ
- name: "Connect to RabbitMQ"
  action: rabbitmq
  args: ["connect", "amqp://guest:guest@localhost:5672/", "main"]

# Publish message
- name: "Publish message"
  action: rabbitmq
  args: ["publish", "main", "test-exchange", "test.route", "Hello RabbitMQ"]
\`\`\`
            `,
            snippet: 'action: rabbitmq\n  args: ["${1:publish}", "${2:connection}", "${3:exchange}", "${4:routing_key}", "${5:message}"]${6:\n  result: ${7:result}}',
            parameters: [
                { name: 'subcommand', type: 'string', description: 'RabbitMQ operation', required: true, examples: ['connect', 'publish', 'consume', 'close'] },
                { name: 'connection_string', type: 'string', description: 'AMQP connection string', required: false },
            ],
            examples: [
                'action: rabbitmq\n  args: ["connect", "amqp://localhost:5672/"]',
                'action: rabbitmq\n  args: ["publish", "conn", "exchange", "key", "message"]'
            ],
            category: 'Messaging'
        });

        // Control Flow Actions
        this.addAction({
            name: 'if',
            description: 'Conditional execution of steps',
            documentation: `
**If Action**

Execute steps conditionally based on expressions.

**Supported Operators:**
- \`==\`, \`!=\`: Equality/inequality
- \`>\`, \`<\`, \`>=\`, \`<=\`: Comparison
- \`contains\`, \`not_contains\`: String contains
- \`starts_with\`, \`ends_with\`: String matching
- \`matches\`, \`not_matches\`: Regex matching

**Parameters:**
- \`condition\`: Boolean expression to evaluate (required)
- \`then\`: Steps to execute if true (required)
- \`else\`: Steps to execute if false (optional)

**Examples:**
\`\`\`yaml
# Simple condition
- name: "Check response code"
  action: if
  args: ["\${response.status_code}", "==", "200"]
  then:
    - name: "Success"
      action: log
      args: ["Request successful"]
  else:
    - name: "Failure"  
      action: log
      args: ["Request failed"]
\`\`\`
            `,
            snippet: 'action: if\n  args: ["${1:condition}", "${2:==}", "${3:value}"]\n  then:\n    - name: "${4:Then step}"\n      action: ${5:log}\n      args: ["${6:message}"]',
            parameters: [
                { name: 'left', type: 'any', description: 'Left operand', required: true },
                { name: 'operator', type: 'string', description: 'Comparison operator', required: true, examples: ['==', '!=', '>', '<', 'contains'] },
                { name: 'right', type: 'any', description: 'Right operand', required: true },
            ],
            examples: [
                'action: if\n  args: ["${status}", "==", "200"]\n  then:\n    - action: log\n      args: ["Success"]'
            ],
            category: 'Control Flow'
        });

        this.addAction({
            name: 'for',
            description: 'Loop over collections or ranges',
            documentation: `
**For Action**

Iterate over collections, ranges, or lists with loop variable access.

**Loop Types:**
- Collection iteration: \`for item in collection\`
- Range iteration: \`for i in range(start, end)\`
- List iteration: \`for item in [1, 2, 3]\`

**Parameters:**
- \`variable\`: Loop variable name (required)
- \`collection\`: Collection to iterate over (required)
- \`steps\`: Steps to execute in each iteration (required)

**Examples:**
\`\`\`yaml
# Iterate over list
- name: "Process users"
  action: for
  args: ["user", "\${users}"]
  steps:
    - name: "Process user"
      action: log
      args: ["Processing \${user.name}"]

# Range iteration  
- name: "Retry loop"
  action: for
  args: ["i", "range(1, 4)"]
  steps:
    - name: "Attempt \${i}"
      action: http
      args: ["GET", "https://api.example.com/data"]
\`\`\`
            `,
            snippet: 'action: for\n  args: ["${1:item}", "${2:collection}"]\n  steps:\n    - name: "${3:Step name}"\n      action: ${4:log}\n      args: ["${5:Processing} \\${${1:item}}"]',
            parameters: [
                { name: 'variable', type: 'string', description: 'Loop variable name', required: true },
                { name: 'collection', type: 'any', description: 'Collection to iterate', required: true },
            ],
            examples: [
                'action: for\n  args: ["item", "${list}"]\n  steps:\n    - action: log\n      args: ["${item}"]'
            ],
            category: 'Control Flow'
        });

        // Utility Actions
        this.addAction({
            name: 'assert',
            description: 'Assert conditions and validate results',
            documentation: `
**Assert Action**

Validate conditions and fail tests if assertions are not met.

**Supported Operators:**
- \`==\`, \`!=\`: Equality/inequality
- \`>\`, \`<\`, \`>=\`, \`<=\`: Numeric comparison
- \`contains\`, \`not_contains\`: String/array contains
- \`starts_with\`, \`ends_with\`: String prefix/suffix
- \`matches\`, \`not_matches\`: Regex matching
- \`empty\`, \`not_empty\`: Empty/non-empty check

**Parameters:**
- \`actual\`: Actual value to test (required)
- \`operator\`: Assertion operator (required)
- \`expected\`: Expected value (required for most operators)
- \`message\`: Custom assertion message (optional)

**Examples:**
\`\`\`yaml
# Basic equality assertion
- name: "Assert status code"
  action: assert
  args: ["\${response.status_code}", "==", "200", "Expected successful response"]

# String contains assertion
- name: "Assert response contains data"
  action: assert  
  args: ["\${response.body}", "contains", "user_data"]
\`\`\`
            `,
            snippet: 'action: assert\n  args: ["${1:actual}", "${2:==}", "${3:expected}"${4:, "${5:message}"}]',
            parameters: [
                { name: 'actual', type: 'any', description: 'Actual value', required: true },
                { name: 'operator', type: 'string', description: 'Assertion operator', required: true, examples: ['==', '!=', '>', 'contains', 'matches'] },
                { name: 'expected', type: 'any', description: 'Expected value', required: false },
                { name: 'message', type: 'string', description: 'Assertion message', required: false },
            ],
            examples: [
                'action: assert\n  args: ["${status}", "==", "200"]',
                'action: assert\n  args: ["${body}", "contains", "success"]'
            ],
            category: 'Validation'
        });

        this.addAction({
            name: 'variable',
            description: 'Variable management operations',
            documentation: `
**Variable Action**

Manage variables during test execution with dynamic assignment.

**Subcommands:**
- \`set\`: Set variable value

**Parameters:**
- \`subcommand\`: Variable operation (required)
- \`name\`: Variable name (required)
- \`value\`: Variable value (required for set)

**Examples:**
\`\`\`yaml
# Set simple variable
- name: "Set user ID"
  action: variable
  args: ["set", "user_id", "12345"]

# Set complex variable
- name: "Set user object"
  action: variable
  args: ["set", "user", "{\\"id\\": 123, \\"name\\": \\"John\\"}"]
\`\`\`
            `,
            snippet: 'action: variable\n  args: ["set", "${1:name}", "${2:value}"]',
            parameters: [
                { name: 'subcommand', type: 'string', description: 'Variable operation', required: true, examples: ['set'] },
                { name: 'name', type: 'string', description: 'Variable name', required: true },
                { name: 'value', type: 'any', description: 'Variable value', required: true },
            ],
            examples: [
                'action: variable\n  args: ["set", "name", "value"]',
                'action: variable\n  args: ["set", "config", "${response.config}"]'
            ],
            category: 'Variables'
        });

        // Template Actions
        this.addAction({
            name: 'template',
            description: 'Render templates with data',
            documentation: `
**Template Action**

Render Go templates with variable substitution for message generation.

**Template Sources:**
- File-based: Templates loaded from files
- Context-based: Templates defined in test case

**Parameters:**
- \`template_path\`: Path to template file or template name (required)
- \`data\`: Data object for template rendering (required)

**Examples:**
\`\`\`yaml
# Render SWIFT MT103 template
- name: "Generate MT103 message"
  action: template
  args: ["templates/mt103.tmpl", {
    "TransactionID": "TXN123",
    "Amount": "1000.00",
    "Currency": "USD",
    "Sender": {"BIC": "BANKUSXX", "Name": "Test Bank"}
  }]
  result: mt103_message

# Use inline template
- name: "Generate custom message"
  action: template
  args: ["greeting_template", {"name": "John", "time": "morning"}]
  result: greeting
\`\`\`
            `,
            snippet: 'action: template\n  args: ["${1:template_path}", {\n    "${2:field}": "${3:value}"\n  }]\n  result: ${4:result}',
            parameters: [
                { name: 'template_path', type: 'string', description: 'Template file path or name', required: true },
                { name: 'data', type: 'object', description: 'Template data object', required: true },
            ],
            examples: [
                'action: template\n  args: ["templates/mt103.tmpl", {"amount": "100"}]',
                'action: template\n  args: ["greeting", {"name": "John"}]'
            ],
            category: 'Templates'
        });

        // TDM Actions
        this.addAction({
            name: 'tdm',
            description: 'Test Data Management operations',
            documentation: `
**TDM (Test Data Management) Action**

Generate and manage test data with structured data sets.

**Subcommands:**
- \`generate\`: Generate test data based on specification

**Data Types:**
- \`person\`: Personal data (name, email, phone, etc.)
- \`address\`: Address information
- \`company\`: Company/organization data
- \`finance\`: Financial data (accounts, transactions)
- \`custom\`: Custom data specification

**Parameters:**
- \`subcommand\`: TDM operation (required)
- \`data_type\`: Type of data to generate (required)
- \`count\`: Number of records to generate (optional, default: 1)
- \`specification\`: Custom data specification (optional)

**Examples:**
\`\`\`yaml
# Generate person data
- name: "Generate test users"
  action: tdm
  args: ["generate", "person", "5"]
  result: test_users

# Generate financial data
- name: "Generate transactions"
  action: tdm
  args: ["generate", "finance", "10", {"type": "transaction", "currency": "USD"}]
  result: transactions
\`\`\`
            `,
            snippet: 'action: tdm\n  args: ["generate", "${1:person}", "${2:1}"]\n  result: ${3:test_data}',
            parameters: [
                { name: 'subcommand', type: 'string', description: 'TDM operation', required: true, examples: ['generate'] },
                { name: 'data_type', type: 'string', description: 'Data type to generate', required: true, examples: ['person', 'address', 'company', 'finance'] },
                { name: 'count', type: 'number', description: 'Number of records', required: false, default: 1 },
            ],
            examples: [
                'action: tdm\n  args: ["generate", "person", "5"]',
                'action: tdm\n  args: ["generate", "finance", "10"]'
            ],
            category: 'Data'
        });

        // Utility Actions
        this.addAction({
            name: 'log',
            description: 'Log messages and debug information',
            documentation: `
**Log Action**

Output log messages for debugging and information purposes.

**Log Levels:**
- \`info\`: Informational messages (default)
- \`debug\`: Debug information
- \`warn\`: Warning messages
- \`error\`: Error messages

**Parameters:**
- \`message\`: Log message (required)
- \`level\`: Log level (optional, default: info)

**Examples:**
\`\`\`yaml
# Simple log message
- name: "Log status"
  action: log
  args: ["Test execution started"]

# Log with level
- name: "Debug info"
  action: log
  args: ["Variable value: \${user_id}", "debug"]
\`\`\`
            `,
            snippet: 'action: log\n  args: ["${1:message}"${2:, "${3:info}"}]',
            parameters: [
                { name: 'message', type: 'string', description: 'Log message', required: true },
                { name: 'level', type: 'string', description: 'Log level', required: false, default: 'info', examples: ['info', 'debug', 'warn', 'error'] },
            ],
            examples: [
                'action: log\n  args: ["Starting test"]',
                'action: log\n  args: ["Debug info", "debug"]'
            ],
            category: 'Utility'
        });
    }

    /**
     * Add action definition
     */
    private addAction(action: ActionDefinition): void {
        this.actions.set(action.name, action);
    }
}