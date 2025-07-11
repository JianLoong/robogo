import * as vscode from 'vscode';
import * as path from 'path';

/**
 * Generates templates and boilerplate code for Robogo tests
 */
export class TemplateGenerator {

    /**
     * Generate template based on type
     */
    async generateTemplate(templateType: string): Promise<void> {
        switch (templateType) {
            case 'testcase':
                await this.generateTestCase();
                break;
            case 'testsuite':
                await this.generateTestSuite();
                break;
            case 'http':
                await this.generateHTTPTest();
                break;
            case 'database':
                await this.generateDatabaseTest();
                break;
            case 'integration':
                await this.generateIntegrationTest();
                break;
            default:
                vscode.window.showErrorMessage(`Unknown template type: ${templateType}`);
        }
    }

    /**
     * Generate basic test case
     */
    async generateTestCase(): Promise<void> {
        const fileName = await vscode.window.showInputBox({
            prompt: 'Enter test case name',
            placeHolder: 'my-test-case',
            validateInput: (value) => {
                if (!value || value.trim() === '') {
                    return 'Test case name cannot be empty';
                }
                if (!/^[a-zA-Z0-9\-_]+$/.test(value)) {
                    return 'Test case name should only contain letters, numbers, hyphens, and underscores';
                }
                return null;
            }
        });

        if (!fileName) return;

        const content = this.getTestCaseTemplate(fileName);
        await this.createFileWithContent(`${fileName}.robogo`, content);
    }

    /**
     * Generate test suite
     */
    async generateTestSuite(): Promise<void> {
        const suiteName = await vscode.window.showInputBox({
            prompt: 'Enter test suite name',
            placeHolder: 'my-test-suite',
            validateInput: (value) => {
                if (!value || value.trim() === '') {
                    return 'Test suite name cannot be empty';
                }
                return null;
            }
        });

        if (!suiteName) return;

        const content = this.getTestSuiteTemplate(suiteName);
        await this.createFileWithContent(`${suiteName}.robogo`, content);
    }

    /**
     * Generate HTTP API test
     */
    async generateHTTPTest(): Promise<void> {
        const testName = await vscode.window.showInputBox({
            prompt: 'Enter API test name',
            placeHolder: 'api-test',
        });

        if (!testName) return;

        const baseUrl = await vscode.window.showInputBox({
            prompt: 'Enter API base URL',
            placeHolder: 'https://api.example.com',
            value: 'https://api.example.com'
        });

        if (!baseUrl) return;

        const content = this.getHTTPTestTemplate(testName, baseUrl);
        await this.createFileWithContent(`${testName}.robogo`, content);
    }

    /**
     * Generate database test
     */
    async generateDatabaseTest(): Promise<void> {
        const testName = await vscode.window.showInputBox({
            prompt: 'Enter database test name',
            placeHolder: 'database-test',
        });

        if (!testName) return;

        const dbType = await vscode.window.showQuickPick(['PostgreSQL', 'Google Cloud Spanner'], {
            placeHolder: 'Select database type'
        });

        if (!dbType) return;

        const content = this.getDatabaseTestTemplate(testName, dbType);
        await this.createFileWithContent(`${testName}.robogo`, content);
    }

    /**
     * Generate integration test
     */
    async generateIntegrationTest(): Promise<void> {
        const testName = await vscode.window.showInputBox({
            prompt: 'Enter integration test name',
            placeHolder: 'integration-test',
        });

        if (!testName) return;

        const content = this.getIntegrationTestTemplate(testName);
        await this.createFileWithContent(`${testName}.robogo`, content);
    }

    /**
     * Create file with content
     */
    private async createFileWithContent(fileName: string, content: string): Promise<void> {
        try {
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                vscode.window.showErrorMessage('No workspace folder open. Please open a folder first.');
                return;
            }

            const filePath = path.join(workspaceFolder.uri.fsPath, fileName);
            const fileUri = vscode.Uri.file(filePath);

            // Check if file already exists
            try {
                await vscode.workspace.fs.stat(fileUri);
                const overwrite = await vscode.window.showWarningMessage(
                    `File ${fileName} already exists. Overwrite?`,
                    'Overwrite',
                    'Cancel'
                );
                if (overwrite !== 'Overwrite') return;
            } catch {
                // File doesn't exist, continue
            }

            // Create and open the file
            await vscode.workspace.fs.writeFile(fileUri, Buffer.from(content, 'utf8'));
            
            const document = await vscode.workspace.openTextDocument(fileUri);
            await vscode.window.showTextDocument(document);

            vscode.window.showInformationMessage(`✅ Created ${fileName} successfully!`);

        } catch (error) {
            vscode.window.showErrorMessage(`Failed to create file: ${error}`);
        }
    }

    /**
     * Get basic test case template
     */
    private getTestCaseTemplate(name: string): string {
        return `testcase: "${this.toTitleCase(name.replace(/[-_]/g, ' '))}"
description: "Description of what this test case validates"

variables:
  vars:
    base_url: "https://api.example.com"
    timeout: "30s"
  secrets:
    api_key:
      file: "secrets/api_key.txt"
      mask_output: true

steps:
  - name: "Setup test data"
    action: log
    args: ["Starting test case: ${name}"]
    
  - name: "Execute main test logic"
    action: http
    args: ["GET", "\${base_url}/health"]
    options:
      timeout: "\${timeout}"
      headers:
        Authorization: "Bearer \${SECRETS.api_key}"
    result: health_response
    
  - name: "Validate response"
    action: assert
    args: ["\${health_response.status_code}", "==", "200", "Health check should return 200"]
    
  - name: "Verify response body"
    action: assert
    args: ["\${health_response.body}", "contains", "healthy", "Response should indicate healthy status"]
    
  - name: "Log success"
    action: log
    args: ["Test case completed successfully"]
`;
    }

    /**
     * Get test suite template
     */
    private getTestSuiteTemplate(name: string): string {
        return `testsuite: "${this.toTitleCase(name.replace(/[-_]/g, ' '))}"
description: "Test suite description and purpose"

# Global setup - runs before all test cases
setup:
  - name: "Initialize test environment"
    action: log
    args: ["Setting up test suite: ${name}"]
    
  - name: "Prepare test data"
    action: variable
    args: ["set", "suite_start_time", "\${get_time}"]

# Global teardown - runs after all test cases
teardown:
  - name: "Clean up test environment"
    action: log
    args: ["Cleaning up test suite: ${name}"]
    
  - name: "Log execution time"
    action: log
    args: ["Suite execution completed at: \${get_time}"]

# Test cases to execute
testcases:
  - test-case-1.robogo
  - test-case-2.robogo
  - test-case-3.robogo

# Parallel execution configuration
parallel:
  enabled: true
  max_concurrency: 3
  test_cases: true
  steps: false
`;
    }

    /**
     * Get HTTP test template
     */
    private getHTTPTestTemplate(name: string, baseUrl: string): string {
        return `testcase: "${this.toTitleCase(name.replace(/[-_]/g, ' '))}"
description: "HTTP API testing with comprehensive validation"

variables:
  vars:
    base_url: "${baseUrl}"
    user_id: "123"
    timeout: "30s"
  secrets:
    api_token:
      file: "secrets/api_token.txt"
      mask_output: true

steps:
  - name: "Get user information"
    action: http
    args: ["GET", "\${base_url}/users/\${user_id}"]
    options:
      timeout: "\${timeout}"
      headers:
        Authorization: "Bearer \${SECRETS.api_token}"
        Content-Type: "application/json"
    result: user_response
    
  - name: "Validate user response status"
    action: assert
    args: ["\${user_response.status_code}", "==", "200", "User API should return 200"]
    
  - name: "Validate user data structure"
    action: assert
    args: ["\${user_response.body}", "contains", "id", "Response should contain user ID"]
    
  - name: "Extract user data"
    action: variable
    args: ["set", "user_name", "\${user_response.body.name}"]
    
  - name: "Update user information"
    action: http
    args: ["PUT", "\${base_url}/users/\${user_id}", "{\\"name\\": \\"Updated Name\\", \\"email\\": \\"updated@example.com\\"}"]
    options:
      headers:
        Authorization: "Bearer \${SECRETS.api_token}"
        Content-Type: "application/json"
    result: update_response
    
  - name: "Validate update response"
    action: assert
    args: ["\${update_response.status_code}", "==", "200", "Update should succeed"]
    
  - name: "Verify updated data"
    action: http
    args: ["GET", "\${base_url}/users/\${user_id}"]
    options:
      headers:
        Authorization: "Bearer \${SECRETS.api_token}"
    result: verification_response
    
  - name: "Confirm changes persisted"
    action: assert
    args: ["\${verification_response.body.name}", "==", "Updated Name", "Name should be updated"]
`;
    }

    /**
     * Get database test template
     */
    private getDatabaseTestTemplate(name: string, dbType: string): string {
        const action = dbType === 'PostgreSQL' ? 'postgres' : 'spanner';
        const connectionString = dbType === 'PostgreSQL' 
            ? 'postgres://username:password@localhost:5432/database'
            : 'projects/your-project/instances/your-instance/databases/your-database';

        return `testcase: "${this.toTitleCase(name.replace(/[-_]/g, ' '))}"
description: "${dbType} database operations testing"

variables:
  vars:
    connection_name: "test_db"
  secrets:
    db_connection:
      value: "${connectionString}"
      mask_output: true

steps:
  - name: "Connect to ${dbType} database"
    action: ${action}
    args: ["connect", "\${SECRETS.db_connection}", "\${connection_name}"]
    
  - name: "Create test table"
    action: ${action}
    args: ["execute", "CREATE TABLE IF NOT EXISTS test_users (id SERIAL PRIMARY KEY, name VARCHAR(100), email VARCHAR(100))", "\${connection_name}"]
    result: create_result
    
  - name: "Insert test data"
    action: ${action}
    args: ["execute", "INSERT INTO test_users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com", "\${connection_name}"]
    result: insert_result
    
  - name: "Query test data"
    action: ${action}
    args: ["query", "SELECT * FROM test_users WHERE name = $1", "John Doe", "\${connection_name}"]
    result: query_result
    
  - name: "Validate query results"
    action: assert
    args: ["\${query_result}", "not_empty", "", "Query should return results"]
    
  - name: "Verify user data"
    action: assert
    args: ["\${query_result[0].name}", "==", "John Doe", "Name should match inserted value"]
    
  - name: "Update test data"
    action: ${action}
    args: ["execute", "UPDATE test_users SET email = $1 WHERE name = $2", "newemail@example.com", "John Doe", "\${connection_name}"]
    result: update_result
    
  - name: "Verify update"
    action: ${action}
    args: ["query", "SELECT email FROM test_users WHERE name = $1", "John Doe", "\${connection_name}"]
    result: updated_data
    
  - name: "Validate updated email"
    action: assert
    args: ["\${updated_data[0].email}", "==", "newemail@example.com", "Email should be updated"]
    
  - name: "Clean up test data"
    action: ${action}
    args: ["execute", "DELETE FROM test_users WHERE name = $1", "John Doe", "\${connection_name}"]
    
  - name: "Close database connection"
    action: ${action}
    args: ["close", "\${connection_name}"]
`;
    }

    /**
     * Get integration test template
     */
    private getIntegrationTestTemplate(name: string): string {
        return `testcase: "${this.toTitleCase(name.replace(/[-_]/g, ' '))}"
description: "Complex integration test with multiple systems"

variables:
  vars:
    api_url: "https://api.example.com"
    queue_url: "amqp://localhost:5672"
    db_name: "integration_test"
  secrets:
    api_key:
      file: "secrets/api_key.txt"
      mask_output: true
    db_password:
      file: "secrets/db_password.txt"
      mask_output: true

templates:
  user_payload: |
    {
      "name": "{{.name}}",
      "email": "{{.email}}",
      "department": "{{.department}}"
    }

steps:
  - name: "Initialize test environment"
    action: log
    args: ["Starting integration test: ${name}"]
    
  - name: "Connect to database"
    action: postgres
    args: ["connect", "postgres://testuser:\${SECRETS.db_password}@localhost:5432/\${db_name}", "main"]
    
  - name: "Connect to message queue"
    action: rabbitmq
    args: ["connect", "\${queue_url}", "main"]
    
  - name: "Generate test user data"
    action: tdm
    args: ["generate", "person", "1"]
    result: test_user
    
  - name: "Create user payload from template"
    action: template
    args: ["user_payload", {
      "name": "\${test_user[0].name}",
      "email": "\${test_user[0].email}",
      "department": "Engineering"
    }]
    result: user_json
    
  - name: "Create user via API"
    action: http
    args: ["POST", "\${api_url}/users", "\${user_json}"]
    options:
      headers:
        Authorization: "Bearer \${SECRETS.api_key}"
        Content-Type: "application/json"
    result: create_response
    
  - name: "Validate user creation"
    action: assert
    args: ["\${create_response.status_code}", "==", "201", "User should be created successfully"]
    
  - name: "Extract user ID"
    action: variable
    args: ["set", "new_user_id", "\${create_response.body.id}"]
    
  - name: "Verify user in database"
    action: postgres
    args: ["query", "SELECT * FROM users WHERE id = $1", "\${new_user_id}", "main"]
    result: db_user
    
  - name: "Validate database record"
    action: assert
    args: ["\${db_user[0].name}", "==", "\${test_user[0].name}", "Database should contain correct user name"]
    
  - name: "Publish user event to queue"
    action: rabbitmq
    args: ["publish", "main", "user.events", "user.created", "{\\"user_id\\": \\"\${new_user_id}\\", \\"action\\": \\"created\\"}"]
    
  - name: "Consume event from queue"
    action: rabbitmq
    args: ["consume", "main", "user.events.queue", "5s"]
    result: queue_message
    
  - name: "Validate queue message"
    action: assert
    args: ["\${queue_message.message}", "contains", "\${new_user_id}", "Queue message should contain user ID"]
    
  - name: "Update user via API"
    action: http
    args: ["PUT", "\${api_url}/users/\${new_user_id}", "{\\"department\\": \\"Marketing\\"}"]
    options:
      headers:
        Authorization: "Bearer \${SECRETS.api_key}"
        Content-Type: "application/json"
    result: update_response
    
  - name: "Verify update in database"
    action: postgres
    args: ["query", "SELECT department FROM users WHERE id = $1", "\${new_user_id}", "main"]
    result: updated_user
    
  - name: "Validate department update"
    action: assert
    args: ["\${updated_user[0].department}", "==", "Marketing", "Department should be updated"]
    
  - name: "Clean up test user"
    action: http
    args: ["DELETE", "\${api_url}/users/\${new_user_id}"]
    options:
      headers:
        Authorization: "Bearer \${SECRETS.api_key}"
    
  - name: "Close database connection"
    action: postgres
    args: ["close", "main"]
    
  - name: "Close message queue connection"
    action: rabbitmq
    args: ["close", "main"]
    
  - name: "Log completion"
    action: log
    args: ["Integration test completed successfully"]

# Enable parallel execution for performance
parallel:
  enabled: true
  max_concurrency: 2
  steps: false  # Keep steps sequential for data consistency
`;
    }

    /**
     * Convert string to title case
     */
    private toTitleCase(str: string): string {
        return str.replace(/\w\S*/g, (txt) => 
            txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase()
        );
    }
}