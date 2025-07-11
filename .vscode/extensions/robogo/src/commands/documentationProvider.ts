import * as vscode from 'vscode';
import { RobogoLanguageServer } from '../core/languageServer';

/**
 * Provides documentation and help for Robogo
 */
export class DocumentationProvider {

    constructor(private languageServer: RobogoLanguageServer) {}

    /**
     * Show main documentation
     */
    async showDocumentation(): Promise<void> {
        const panel = vscode.window.createWebviewPanel(
            'robogoDocumentation',
            'Robogo Documentation',
            vscode.ViewColumn.Beside,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = this.getDocumentationHTML();
    }

    /**
     * Show action-specific help
     */
    async showActionHelp(actionName: string): Promise<void> {
        const action = this.languageServer.getActionRegistry().getAction(actionName);
        if (!action) {
            vscode.window.showErrorMessage(`Action '${actionName}' not found.`);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'robogoActionHelp',
            `Robogo Action: ${actionName}`,
            vscode.ViewColumn.Beside,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = this.getActionHelpHTML(action);
    }

    /**
     * Show examples
     */
    async showExamples(): Promise<void> {
        const panel = vscode.window.createWebviewPanel(
            'robogoExamples',
            'Robogo Examples',
            vscode.ViewColumn.Beside,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = this.getExamplesHTML();
    }

    /**
     * Generate main documentation HTML
     */
    private getDocumentationHTML(): string {
        return `
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Robogo Documentation</title>
            <style>
                body {
                    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                    line-height: 1.6;
                    color: #333;
                    max-width: 1200px;
                    margin: 0 auto;
                    padding: 20px;
                }
                .header {
                    text-align: center;
                    margin-bottom: 40px;
                    padding: 20px;
                    background: linear-gradient(135deg, #007ACC, #005A9E);
                    color: white;
                    border-radius: 10px;
                }
                .section {
                    margin: 30px 0;
                    padding: 20px;
                    border: 1px solid #ddd;
                    border-radius: 8px;
                    background: #f9f9f9;
                }
                .section h2 {
                    color: #007ACC;
                    margin-top: 0;
                }
                .feature-grid {
                    display: grid;
                    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
                    gap: 20px;
                    margin: 20px 0;
                }
                .feature-card {
                    padding: 20px;
                    background: white;
                    border: 1px solid #ddd;
                    border-radius: 8px;
                    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                }
                .feature-card h3 {
                    color: #007ACC;
                    margin-top: 0;
                }
                .code-block {
                    background: #f4f4f4;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                    padding: 15px;
                    font-family: 'Courier New', monospace;
                    overflow-x: auto;
                    margin: 10px 0;
                }
                .quick-links {
                    display: flex;
                    gap: 15px;
                    flex-wrap: wrap;
                    margin: 20px 0;
                }
                .quick-link {
                    padding: 10px 20px;
                    background: #007ACC;
                    color: white;
                    text-decoration: none;
                    border-radius: 5px;
                    transition: background 0.3s;
                }
                .quick-link:hover {
                    background: #005A9E;
                }
                .tip {
                    background: #e7f3ff;
                    border-left: 4px solid #007ACC;
                    padding: 15px;
                    margin: 15px 0;
                }
                .warning {
                    background: #fff3cd;
                    border-left: 4px solid #ffc107;
                    padding: 15px;
                    margin: 15px 0;
                }
            </style>
        </head>
        <body>
            <div class="header">
                <h1>🎯 Robogo Test Automation Framework</h1>
                <p>Modern, git-driven test automation with YAML-based test cases</p>
            </div>

            <div class="quick-links">
                <a href="#getting-started" class="quick-link">Getting Started</a>
                <a href="#actions" class="quick-link">Actions Reference</a>
                <a href="#examples" class="quick-link">Examples</a>
                <a href="#advanced" class="quick-link">Advanced Features</a>
            </div>

            <div class="section" id="getting-started">
                <h2>🚀 Getting Started</h2>
                <p>Robogo provides a powerful yet simple way to write test automation using YAML syntax. Here's how to create your first test:</p>
                
                <h3>Basic Test Case Structure</h3>
                <div class="code-block">testcase: "My First Test"
description: "A simple test to get started"

variables:
  vars:
    api_url: "https://api.example.com"
  secrets:
    api_key:
      file: "secrets/api_key.txt"
      mask_output: true

steps:
  - name: "Check API health"
    action: http
    args: ["GET", "\${api_url}/health"]
    result: health_response
    
  - name: "Validate response"
    action: assert
    args: ["\${health_response.status_code}", "==", "200"]</div>

                <div class="tip">
                    <strong>💡 Tip:</strong> Use the "Robogo: Generate Template" command to quickly create new test files!
                </div>
            </div>

            <div class="section" id="key-features">
                <h2>✨ Key Features</h2>
                <div class="feature-grid">
                    <div class="feature-card">
                        <h3>🌐 HTTP Testing</h3>
                        <p>Full HTTP client with support for all methods, custom headers, mTLS, and comprehensive response handling.</p>
                    </div>
                    <div class="feature-card">
                        <h3>🗄️ Database Support</h3>
                        <p>Native support for PostgreSQL and Google Cloud Spanner with parameterized queries and connection management.</p>
                    </div>
                    <div class="feature-card">
                        <h3>📨 Messaging</h3>
                        <p>Kafka and RabbitMQ integration for publish/consume testing scenarios.</p>
                    </div>
                    <div class="feature-card">
                        <h3>🔒 Secret Management</h3>
                        <p>Secure handling of sensitive data with automatic output masking and file-based secrets.</p>
                    </div>
                    <div class="feature-card">
                        <h3>📄 Templates</h3>
                        <p>Built-in support for SWIFT message generation and custom template rendering.</p>
                    </div>
                    <div class="feature-card">
                        <h3>⚡ Parallel Execution</h3>
                        <p>Run tests and steps in parallel for improved performance with configurable concurrency.</p>
                    </div>
                </div>
            </div>

            <div class="section" id="actions">
                <h2>🎯 Core Actions</h2>
                <p>Robogo provides a comprehensive set of built-in actions:</p>
                
                <h3>HTTP & API Testing</h3>
                <ul>
                    <li><strong>http</strong> - Execute HTTP requests with full feature support</li>
                </ul>
                
                <h3>Database Operations</h3>
                <ul>
                    <li><strong>postgres</strong> - PostgreSQL database operations</li>
                    <li><strong>spanner</strong> - Google Cloud Spanner operations</li>
                </ul>
                
                <h3>Messaging</h3>
                <ul>
                    <li><strong>kafka</strong> - Kafka publish/consume operations</li>
                    <li><strong>rabbitmq</strong> - RabbitMQ message operations</li>
                </ul>
                
                <h3>Control Flow</h3>
                <ul>
                    <li><strong>if</strong> - Conditional execution</li>
                    <li><strong>for</strong> - Loop over collections</li>
                    <li><strong>while</strong> - Loop with conditions</li>
                </ul>
                
                <h3>Validation & Utilities</h3>
                <ul>
                    <li><strong>assert</strong> - Validate conditions and test results</li>
                    <li><strong>log</strong> - Output messages for debugging</li>
                    <li><strong>variable</strong> - Manage variables during execution</li>
                    <li><strong>template</strong> - Render templates with data</li>
                    <li><strong>tdm</strong> - Test Data Management operations</li>
                </ul>
            </div>

            <div class="section" id="variables">
                <h2>📦 Variables & Secrets</h2>
                <p>Robogo provides powerful variable management with support for regular variables and secure secrets:</p>
                
                <h3>Regular Variables</h3>
                <div class="code-block">variables:
  vars:
    api_url: "https://api.example.com"
    user_id: 123
    config:
      timeout: "30s"
      retries: 3</div>
                
                <h3>Secret Variables</h3>
                <div class="code-block">variables:
  secrets:
    api_token:
      file: "secrets/token.txt"
      mask_output: true
    db_password:
      value: "secret123"
      mask_output: true</div>
      
                <h3>Variable Usage</h3>
                <ul>
                    <li><code>\${variable_name}</code> - Access regular variables</li>
                    <li><code>\${SECRETS.secret_name}</code> - Access secret variables</li>
                    <li><code>\${response.body.field}</code> - Dot notation for nested data</li>
                    <li><code>\${__robogo_steps[0].result}</code> - Access step execution history</li>
                </ul>
            </div>

            <div class="section" id="advanced">
                <h2>⚙️ Advanced Features</h2>
                
                <h3>Parallel Execution</h3>
                <div class="code-block">parallel:
  enabled: true
  max_concurrency: 4
  test_cases: true
  steps: true
  http_requests: true</div>
                
                <h3>Retry Configuration</h3>
                <div class="code-block">- name: "Retry example"
  action: http
  args: ["GET", "https://api.example.com/data"]
  retry:
    attempts: 3
    delay: "1s"
    exponential_backoff: true</div>
                
                <h3>Conditional Execution</h3>
                <div class="code-block">- name: "Conditional step"
  action: log
  args: ["This runs conditionally"]
  if: "\${response.status_code} == 200"</div>
            </div>

            <div class="section" id="best-practices">
                <h2>💡 Best Practices</h2>
                <ul>
                    <li><strong>Descriptive Names:</strong> Use clear, descriptive names for test cases and steps</li>
                    <li><strong>Secret Security:</strong> Always use the SECRETS namespace for sensitive data</li>
                    <li><strong>Error Handling:</strong> Include proper assertions to validate expected behavior</li>
                    <li><strong>Modular Tests:</strong> Break complex tests into smaller, focused test cases</li>
                    <li><strong>Test Data:</strong> Use TDM for generating realistic test data</li>
                    <li><strong>Parallel Execution:</strong> Enable parallel execution for improved performance</li>
                    <li><strong>Documentation:</strong> Include clear descriptions for test cases and complex steps</li>
                </ul>
                
                <div class="warning">
                    <strong>⚠️ Security Note:</strong> Never commit secrets or sensitive data to version control. Always use file-based secrets or environment variables.
                </div>
            </div>

            <div class="section" id="getting-help">
                <h2>🆘 Getting Help</h2>
                <p>Need more help? Here are your options:</p>
                <ul>
                    <li><strong>Command Palette:</strong> Use "Robogo: Show Action Help" for specific action documentation</li>
                    <li><strong>Hover Help:</strong> Hover over actions and variables for instant documentation</li>
                    <li><strong>Examples:</strong> Use "Robogo: Show Examples" for more code samples</li>
                    <li><strong>Validation:</strong> Real-time validation helps catch errors as you type</li>
                    <li><strong>Templates:</strong> Generate boilerplate code with "Robogo: Generate Template"</li>
                </ul>
            </div>
        </body>
        </html>
        `;
    }

    /**
     * Generate action help HTML
     */
    private getActionHelpHTML(action: any): string {
        return `
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Robogo Action: ${action.name}</title>
            <style>
                body {
                    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                    line-height: 1.6;
                    color: #333;
                    max-width: 800px;
                    margin: 0 auto;
                    padding: 20px;
                }
                .header {
                    background: linear-gradient(135deg, #007ACC, #005A9E);
                    color: white;
                    padding: 20px;
                    border-radius: 10px;
                    margin-bottom: 30px;
                }
                .section {
                    margin: 20px 0;
                    padding: 20px;
                    border: 1px solid #ddd;
                    border-radius: 8px;
                    background: #f9f9f9;
                }
                .parameter {
                    margin: 10px 0;
                    padding: 15px;
                    background: white;
                    border: 1px solid #ddd;
                    border-radius: 5px;
                }
                .parameter-name {
                    font-weight: bold;
                    color: #007ACC;
                }
                .required {
                    color: #d73a49;
                    font-weight: bold;
                }
                .optional {
                    color: #6f42c1;
                }
                .code-block {
                    background: #f4f4f4;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                    padding: 15px;
                    font-family: 'Courier New', monospace;
                    overflow-x: auto;
                    margin: 10px 0;
                }
                .category-badge {
                    background: #007ACC;
                    color: white;
                    padding: 5px 10px;
                    border-radius: 15px;
                    font-size: 12px;
                    font-weight: bold;
                }
            </style>
        </head>
        <body>
            <div class="header">
                <h1>🎯 ${action.name}</h1>
                <p>${action.description}</p>
                <span class="category-badge">${action.category}</span>
            </div>

            <div class="section">
                <h2>📋 Parameters</h2>
                ${action.parameters.map((param: any) => `
                    <div class="parameter">
                        <div class="parameter-name">
                            ${param.name} 
                            <span class="${param.required ? 'required' : 'optional'}">
                                (${param.type}) ${param.required ? '* required' : 'optional'}
                            </span>
                        </div>
                        <div>${param.description}</div>
                        ${param.default !== undefined ? `<div><strong>Default:</strong> <code>${param.default}</code></div>` : ''}
                        ${param.examples ? `<div><strong>Examples:</strong> ${param.examples.map((ex: string) => `<code>${ex}</code>`).join(', ')}</div>` : ''}
                    </div>
                `).join('')}
            </div>

            <div class="section">
                <h2>💡 Examples</h2>
                ${action.examples.map((example: string, index: number) => `
                    <h3>Example ${index + 1}</h3>
                    <div class="code-block">${example}</div>
                `).join('')}
            </div>

            <div class="section">
                <h2>📖 Full Documentation</h2>
                <div>${action.documentation}</div>
            </div>
        </body>
        </html>
        `;
    }

    /**
     * Generate examples HTML
     */
    private getExamplesHTML(): string {
        return `
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Robogo Examples</title>
            <style>
                body {
                    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                    line-height: 1.6;
                    color: #333;
                    max-width: 1000px;
                    margin: 0 auto;
                    padding: 20px;
                }
                .header {
                    text-align: center;
                    background: linear-gradient(135deg, #007ACC, #005A9E);
                    color: white;
                    padding: 20px;
                    border-radius: 10px;
                    margin-bottom: 30px;
                }
                .example {
                    margin: 30px 0;
                    padding: 20px;
                    border: 1px solid #ddd;
                    border-radius: 8px;
                    background: #f9f9f9;
                }
                .example h3 {
                    color: #007ACC;
                    margin-top: 0;
                }
                .code-block {
                    background: #f4f4f4;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                    padding: 15px;
                    font-family: 'Courier New', monospace;
                    overflow-x: auto;
                    margin: 15px 0;
                }
                .description {
                    margin: 15px 0;
                    padding: 10px;
                    background: #e7f3ff;
                    border-left: 4px solid #007ACC;
                }
            </style>
        </head>
        <body>
            <div class="header">
                <h1>📚 Robogo Examples</h1>
                <p>Practical examples to get you started</p>
            </div>

            <div class="example">
                <h3>🌐 Simple HTTP API Test</h3>
                <div class="description">Basic API testing with authentication and validation</div>
                <div class="code-block">testcase: "User API Test"
description: "Test user creation and retrieval"

variables:
  vars:
    base_url: "https://jsonplaceholder.typicode.com"
  secrets:
    auth_token:
      value: "your-secret-token"
      mask_output: true

steps:
  - name: "Get user information"
    action: http
    args: ["GET", "\${base_url}/users/1"]
    result: user_response
    
  - name: "Validate response status"
    action: assert
    args: ["\${user_response.status_code}", "==", "200"]
    
  - name: "Validate user data"
    action: assert
    args: ["\${user_response.body.name}", "!=", "", "User should have a name"]</div>
            </div>

            <div class="example">
                <h3>🗄️ Database Testing</h3>
                <div class="description">Database operations with parameterized queries</div>
                <div class="code-block">testcase: "Database Operations Test"
description: "Test database CRUD operations"

variables:
  secrets:
    db_connection:
      value: "postgres://user:pass@localhost:5432/testdb"
      mask_output: true

steps:
  - name: "Connect to database"
    action: postgres
    args: ["connect", "\${SECRETS.db_connection}", "main"]
    
  - name: "Insert test data"
    action: postgres
    args: ["execute", "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com", "main"]
    result: insert_result
    
  - name: "Query inserted data"
    action: postgres
    args: ["query", "SELECT * FROM users WHERE email = $1", "john@example.com", "main"]
    result: user_data
    
  - name: "Validate query results"
    action: assert
    args: ["\${user_data[0].name}", "==", "John Doe"]</div>
            </div>

            <div class="example">
                <h3>🔄 Control Flow Example</h3>
                <div class="description">Using conditional logic and loops</div>
                <div class="code-block">testcase: "Control Flow Test"
description: "Demonstrate conditional execution and loops"

variables:
  vars:
    user_ids: [1, 2, 3, 4, 5]
    api_url: "https://jsonplaceholder.typicode.com"

steps:
  - name: "Process multiple users"
    action: for
    args: ["user_id", "\${user_ids}"]
    steps:
      - name: "Get user \${user_id}"
        action: http
        args: ["GET", "\${api_url}/users/\${user_id}"]
        result: user_response
        
      - name: "Check if user is active"
        action: if
        args: ["\${user_response.status_code}", "==", "200"]
        then:
          - name: "Log active user"
            action: log
            args: ["User \${user_id} is active: \${user_response.body.name}"]
        else:
          - name: "Log inactive user"
            action: log
            args: ["User \${user_id} is not found"]</div>
            </div>

            <div class="example">
                <h3>📄 Template Usage</h3>
                <div class="description">Generate SWIFT messages using templates</div>
                <div class="code-block">testcase: "SWIFT Message Generation"
description: "Generate MT103 SWIFT message"

variables:
  vars:
    transaction_data:
      TransactionID: "TXN123456"
      Amount: "1000.00"
      Currency: "USD"
      Sender:
        BIC: "BANKUSXX"
        Account: "123456789"
        Name: "Test Bank Inc"
      Beneficiary:
        Account: "987654321"
        Name: "John Doe"

steps:
  - name: "Generate MT103 message"
    action: template
    args: ["templates/mt103.tmpl", "\${transaction_data}"]
    result: swift_message
    
  - name: "Validate message format"
    action: assert
    args: ["\${swift_message}", "contains", ":20:TXN123456"]
    
  - name: "Log generated message"
    action: log
    args: ["Generated SWIFT message: \${swift_message}"]</div>
            </div>

            <div class="example">
                <h3>📨 Messaging Example</h3>
                <div class="description">Kafka publish and consume operations</div>
                <div class="code-block">testcase: "Kafka Messaging Test"
description: "Test Kafka publish/consume workflow"

variables:
  vars:
    kafka_broker: "localhost:9092"
    topic_name: "test-topic"
    test_message: '{"event": "user_created", "user_id": 123}'

steps:
  - name: "Publish message to Kafka"
    action: kafka
    args: ["publish", "\${kafka_broker}", "\${topic_name}", "\${test_message}"]
    
  - name: "Consume message from Kafka"
    action: kafka
    args: ["consume", "\${kafka_broker}", "\${topic_name}", "10s"]
    result: consumed_message
    
  - name: "Validate consumed message"
    action: assert
    args: ["\${consumed_message.message}", "contains", "user_created"]</div>
            </div>

            <div class="example">
                <h3>🧪 Test Data Management</h3>
                <div class="description">Generate realistic test data</div>
                <div class="code-block">testcase: "TDM Example"
description: "Generate and use test data"

steps:
  - name: "Generate test users"
    action: tdm
    args: ["generate", "person", "3"]
    result: test_users
    
  - name: "Process each generated user"
    action: for
    args: ["user", "\${test_users}"]
    steps:
      - name: "Create user account"
        action: http
        args: ["POST", "https://api.example.com/users", '{"name": "\${user.name}", "email": "\${user.email}"}']
        result: create_response
        
      - name: "Validate user creation"
        action: assert
        args: ["\${create_response.status_code}", "==", "201"]</div>
            </div>

            <div class="example">
                <h3>⚡ Parallel Execution</h3>
                <div class="description">High-performance testing with parallel execution</div>
                <div class="code-block">testcase: "Parallel Execution Test"
description: "Demonstrate parallel test execution"

parallel:
  enabled: true
  max_concurrency: 3
  steps: true

variables:
  vars:
    endpoints: 
      - "/users"
      - "/posts"
      - "/comments"
    base_url: "https://jsonplaceholder.typicode.com"

steps:
  - name: "Test multiple endpoints in parallel"
    action: for
    args: ["endpoint", "\${endpoints}"]
    steps:
      - name: "Test \${endpoint} endpoint"
        action: http
        args: ["GET", "\${base_url}\${endpoint}"]
        result: "response_\${endpoint}"
        
      - name: "Validate \${endpoint} response"
        action: assert
        args: ["\${response_\${endpoint}.status_code}", "==", "200"]</div>
            </div>
        </body>
        </html>
        `;
    }
}