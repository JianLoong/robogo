import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export function activate(context: vscode.ExtensionContext) {
    console.log('Robogo extension is now active!');

    // Register completion provider for .robogo files
    const completionProvider = vscode.languages.registerCompletionItemProvider(
        [
            { scheme: 'file', language: 'robogo' },
            { scheme: 'file', language: 'yaml' }
        ],
        new RobogoCompletionProvider(),
        ':', ' ', '-', '$', '{' // Trigger on colon, space, dash, and variable start
    );

    context.subscriptions.push(completionProvider);

    // Register hover provider for action documentation
    const hoverProvider = vscode.languages.registerHoverProvider(
        [
            { scheme: 'file', language: 'robogo' },
            { scheme: 'file', language: 'yaml' }
        ],
        new RobogoHoverProvider()
    );

    context.subscriptions.push(hoverProvider);

    // Register signature help provider for action parameters
    const signatureHelpProvider = vscode.languages.registerSignatureHelpProvider(
        [
            { scheme: 'file', language: 'robogo' },
            { scheme: 'file', language: 'yaml' }
        ],
        new RobogoSignatureHelpProvider(),
        '(', ',', ' ' // Trigger on parentheses, comma, and space
    );

    context.subscriptions.push(signatureHelpProvider);

    // Register commands
    const runTestCommand = vscode.commands.registerCommand('robogo.runTest', async () => {
        await runTest();
    });

    const listActionsCommand = vscode.commands.registerCommand('robogo.listActions', async () => {
        await listActions();
    });

    context.subscriptions.push(runTestCommand, listActionsCommand);
}

interface RobogoAction {
    Name: string;
    Description: string;
    Example: string;
    Parameters?: ActionParameter[];
    Returns?: string;
    Notes?: string;
}

interface ActionParameter {
    name: string;
    type: string;
    description: string;
    required: boolean;
    default?: string;
}

class RobogoCompletionProvider implements vscode.CompletionItemProvider {
    async provideCompletionItems(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken,
        context: vscode.CompletionContext
    ): Promise<vscode.CompletionItem[]> {
        const items: vscode.CompletionItem[] = [];

        // Get the line text up to the cursor position
        const linePrefix = document.lineAt(position).text.substr(0, position.character);
        
        // Check if we're in an action field (after "action:")
        if (linePrefix.includes('action:')) {
            try {
                const actions = await this.getActions();
                for (const action of actions) {
                    const item = new vscode.CompletionItem(action.Name, vscode.CompletionItemKind.Function);
                    item.detail = action.Description;
                    item.documentation = new vscode.MarkdownString(
                        `**${action.Name}**\n\n${action.Description}\n\n**Example:**\n\`\`\`yaml\n${action.Example}\n\`\`\``
                    );
                    items.push(item);
                }
            } catch (error) {
                console.error('Failed to get actions:', error);
            }
        }

        // Check if we're in args for get_time action (provide format options)
        if (linePrefix.includes('args:') && this.isInGetTimeContext(document, position)) {
            const timeFormats = [
                { name: 'iso', desc: 'ISO 8601 format (2006-01-02T15:04:05Z07:00)' },
                { name: 'iso_date', desc: 'ISO date only (2006-01-02)' },
                { name: 'iso_time', desc: 'ISO time only (15:04:05)' },
                { name: 'datetime', desc: 'Standard datetime (2006-01-02 15:04:05)' },
                { name: 'date', desc: 'Date only (2006-01-02)' },
                { name: 'time', desc: 'Time only (15:04:05)' },
                { name: 'timestamp', desc: 'Compact timestamp (20060102150405)' },
                { name: 'unix', desc: 'Unix timestamp (seconds since epoch)' },
                { name: 'unix_ms', desc: 'Unix timestamp in milliseconds' }
            ];
            
            for (const format of timeFormats) {
                const item = new vscode.CompletionItem(format.name, vscode.CompletionItemKind.Constant);
                item.detail = `Time format: ${format.desc}`;
                item.insertText = `"${format.name}"`;
                items.push(item);
            }
        }

        // Autocomplete HTTP methods and headers in http action context
        if (linePrefix.includes('args:') && this.isInHTTPContext(document, position)) {
            const httpMethods = ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"];
            for (const method of httpMethods) {
                const item = new vscode.CompletionItem(method, vscode.CompletionItemKind.EnumMember);
                item.detail = 'HTTP method';
                item.insertText = `"${method}"`;
                items.push(item);
            }
            const commonHeaders = ["Content-Type", "Accept", "Authorization", "User-Agent", "Accept-Encoding"];
            for (const header of commonHeaders) {
                const item = new vscode.CompletionItem(header, vscode.CompletionItemKind.Property);
                item.detail = 'HTTP header';
                item.insertText = `"${header}": `;
                items.push(item);
            }
            
            // Add certificate options
            const certOptions = [
                { name: 'cert', desc: 'Client certificate (file path or PEM content)' },
                { name: 'key', desc: 'Client private key (file path or PEM content)' },
                { name: 'ca', desc: 'Custom CA certificate (file path or PEM content)' }
            ];
            for (const option of certOptions) {
                const item = new vscode.CompletionItem(option.name, vscode.CompletionItemKind.Property);
                item.detail = `Certificate option: ${option.desc}`;
                item.insertText = `"${option.name}": `;
                items.push(item);
            }
        }

        // Check if we're at the start of a step (after "-")
        if (linePrefix.trim().endsWith('-') || linePrefix.trim() === '') {
            // Add action field completion
            items.push(
                this.createCompletionItem('name', 'Step name (strongly recommended for clarity)', vscode.CompletionItemKind.Property),
                this.createCompletionItem('action', 'Action to execute', vscode.CompletionItemKind.Property),
                this.createCompletionItem('args', 'Arguments for the action', vscode.CompletionItemKind.Property),
                this.createCompletionItem('result', 'Variable name to store the result', vscode.CompletionItemKind.Property)
            );
        }

        // Add common YAML/Robogo structure completions
        if (linePrefix.trim() === '' || linePrefix.trim().endsWith('-')) {
            items.push(
                this.createCompletionItem('testcase', 'Test case name', vscode.CompletionItemKind.Property),
                this.createCompletionItem('description', 'Test case description', vscode.CompletionItemKind.Property),
                this.createCompletionItem('steps', 'List of test steps', vscode.CompletionItemKind.Property)
            );
        }

        // Add step structure completions
        if (linePrefix.includes('steps:') || linePrefix.includes('- action:')) {
            items.push(
                this.createCompletionItem('name', 'Step name (strongly recommended for clarity)', vscode.CompletionItemKind.Property),
                this.createCompletionItem('action', 'Action to execute', vscode.CompletionItemKind.Property),
                this.createCompletionItem('args', 'Arguments for the action', vscode.CompletionItemKind.Property),
                this.createCompletionItem('result', 'Variable name to store the result', vscode.CompletionItemKind.Property)
            );
        }

        // Variable reference completion (suggest variables when typing ${...})
        if (linePrefix.includes('${')) {
            // Try to find variable names in the document
            const variableNames = this.extractVariableNames(document);
            for (const varName of variableNames) {
                const item = new vscode.CompletionItem(varName, vscode.CompletionItemKind.Variable);
                item.detail = 'Variable';
                item.insertText = varName;
                items.push(item);
            }
        }

        return items;
    }

    private createCompletionItem(label: string, detail: string, kind: vscode.CompletionItemKind): vscode.CompletionItem {
        const item = new vscode.CompletionItem(label, kind);
        item.detail = detail;
        return item;
    }

    private async getActions(): Promise<RobogoAction[]> {
        try {
            const config = vscode.workspace.getConfiguration('robogo');
            const executablePath = config.get<string>('executablePath', 'robogo');
            
            const { stdout } = await execAsync(`${executablePath} list --output json`);
            return JSON.parse(stdout);
        } catch (error) {
            console.error('Failed to get actions:', error);
            // Return built-in actions as fallback (matching current registry)
            return [
                {
                    Name: "assert",
                    Description: "Assert two values are equal",
                    Example: "- action: assert\n  args: [\"expected\", \"actual\", \"message\"]"
                },
                {
                    Name: "concat",
                    Description: "Concatenate strings",
                    Example: "- action: concat\n  args: [\"Hello\", \" \", \"World\"]\n  result: message"
                },
                {
                    Name: "get_random",
                    Description: "Get a random number",
                    Example: "- action: get_random\n  args: [100]\n  result: random_number"
                },
                {
                    Name: "get_time",
                    Description: "Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)",
                    Example: "- action: get_time\n  args: [\"iso\"]\n  result: timestamp"
                },
                {
                    Name: "http",
                    Description: "Perform HTTP request (GET, POST, PUT, DELETE, etc.). Supports client cert (cert/key) and custom CA (ca) options.",
                    Example: "- action: http\n  args: [\"GET\", \"https://secure.example.com/api\", {\"cert\": \"client.crt\", \"key\": \"client.key\", \"ca\": \"ca.crt\", \"Authorization\": \"Bearer ...\"}]\n  result: response"
                },
                {
                    Name: "http_get",
                    Description: "Perform HTTP GET request",
                    Example: "- action: http_get\n  args: [\"https://api.example.com/users\"]\n  result: response"
                },
                {
                    Name: "http_post",
                    Description: "Perform HTTP POST request",
                    Example: "- action: http_post\n  args: [\"https://api.example.com/users\", '{\"name\": \"John\"}']\n  result: response"
                },
                {
                    Name: "length",
                    Description: "Get length of string or array",
                    Example: "- action: length\n  args: [\"Hello World\"]\n  result: str_length"
                },
                {
                    Name: "log",
                    Description: "Log a message",
                    Example: "- action: log\n  args: [\"message\"]"
                },
                {
                    Name: "sleep",
                    Description: "Sleep for a duration",
                    Example: "- action: sleep\n  args: [2]"
                }
            ];
        }
    }

    // Extract variable names from the document (simple heuristic: find all result: ...)
    private extractVariableNames(document: vscode.TextDocument): string[] {
        const text = document.getText();
        const regex = /result:\s*([a-zA-Z_][a-zA-Z0-9_]*)/g;
        const variables = new Set<string>();
        let match;
        while ((match = regex.exec(text)) !== null) {
            variables.add(match[1]);
        }
        return Array.from(variables);
    }

    // Check if we're in a get_time action context
    private isInGetTimeContext(document: vscode.TextDocument, position: vscode.Position): boolean {
        // Check current line and previous few lines for get_time action
        for (let i = Math.max(0, position.line - 3); i <= position.line; i++) {
            const line = document.lineAt(i).text;
            if (line.includes('action: get_time')) {
                return true;
            }
        }
        return false;
    }

    // Check if we're in an HTTP action context
    private isInHTTPContext(document: vscode.TextDocument, position: vscode.Position): boolean {
        // Check current line and previous few lines for http action
        for (let i = Math.max(0, position.line - 3); i <= position.line; i++) {
            const line = document.lineAt(i).text;
            if (line.includes('action: http') || line.includes('action: http_get') || line.includes('action: http_post')) {
                return true;
            }
        }
        return false;
    }
}

class RobogoHoverProvider implements vscode.HoverProvider {
    async provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Hover | undefined> {
        const wordRange = document.getWordRangeAtPosition(position);
        if (!wordRange) {
            return undefined;
        }

        const word = document.getText(wordRange);
        
        // Check if the word is an action
        try {
            const actions: RobogoAction[] = await this.getActions();
            const action = actions.find(a => a.Name === word);
            
            if (action) {
                const markdown = new vscode.MarkdownString();
                markdown.appendMarkdown(`## ${action.Name}\n\n`);
                markdown.appendMarkdown(`${action.Description}\n\n`);
                
                // Add detailed parameter information if available
                if (action.Parameters && action.Parameters.length > 0) {
                    markdown.appendMarkdown(`### Parameters\n\n`);
                    for (const param of action.Parameters) {
                        const required = param.required ? '**Required**' : 'Optional';
                        const defaultValue = param.default ? ` (default: \`${param.default}\`)` : '';
                        markdown.appendMarkdown(`- **\`${param.name}\`** (${param.type}) - ${required}${defaultValue}\n`);
                        markdown.appendMarkdown(`  ${param.description}\n\n`);
                    }
                }
                
                // Add return value information if available
                if (action.Returns) {
                    markdown.appendMarkdown(`### Returns\n\n`);
                    markdown.appendMarkdown(`${action.Returns}\n\n`);
                }
                
                // Add notes if available
                if (action.Notes) {
                    markdown.appendMarkdown(`### Notes\n\n`);
                    markdown.appendMarkdown(`${action.Notes}\n\n`);
                }
                
                markdown.appendMarkdown(`### Example\n\n`);
                markdown.appendMarkdown(`\`\`\`yaml\n${action.Example}\n\`\`\`\n`);
                
                // Add related actions for similar functionality
                const relatedActions = this.getRelatedActions(action.Name, actions);
                if (relatedActions.length > 0) {
                    markdown.appendMarkdown(`### Related Actions\n\n`);
                    for (const related of relatedActions) {
                        markdown.appendMarkdown(`- \`${related.Name}\` - ${related.Description}\n`);
                    }
                }
                
                return new vscode.Hover(markdown, wordRange);
            }
        } catch (error) {
            console.error('Failed to get action info:', error);
        }

        return undefined;
    }
    
    private getRelatedActions(actionName: string, allActions: RobogoAction[]): RobogoAction[] {
        const related: RobogoAction[] = [];
        
        // Define action groups
        const actionGroups = {
            'http': ['http', 'http_get', 'http_post'],
            'time': ['get_time'],
            'string': ['concat', 'length'],
            'testing': ['assert'],
            'utility': ['log', 'sleep', 'get_random']
        };
        
        // Find which group the current action belongs to
        for (const [group, actions] of Object.entries(actionGroups)) {
            if (actions.includes(actionName)) {
                // Add other actions from the same group
                for (const action of actions) {
                    if (action !== actionName) {
                        const found = allActions.find(a => a.Name === action);
                        if (found) {
                            related.push(found);
                        }
                    }
                }
                break;
            }
        }
        
        return related.slice(0, 3); // Limit to 3 related actions
    }

    private async getActions(): Promise<RobogoAction[]> {
        try {
            const config = vscode.workspace.getConfiguration('robogo');
            const executablePath = config.get<string>('executablePath', 'robogo');
            
            const { stdout } = await execAsync(`${executablePath} list --output json`);
            return JSON.parse(stdout);
        } catch (error) {
            console.error('Failed to get actions:', error);
            // Return built-in actions as fallback (matching current registry)
            return [
                {
                    Name: "assert",
                    Description: "Assert two values are equal",
                    Example: "- action: assert\n  args: [\"expected\", \"actual\", \"message\"]"
                },
                {
                    Name: "concat",
                    Description: "Concatenate strings",
                    Example: "- action: concat\n  args: [\"Hello\", \" \", \"World\"]\n  result: message"
                },
                {
                    Name: "get_random",
                    Description: "Get a random number",
                    Example: "- action: get_random\n  args: [100]\n  result: random_number"
                },
                {
                    Name: "get_time",
                    Description: "Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)",
                    Example: "- action: get_time\n  args: [\"iso\"]\n  result: timestamp"
                },
                {
                    Name: "http",
                    Description: "Perform HTTP request (GET, POST, PUT, DELETE, etc.). Supports client cert (cert/key) and custom CA (ca) options.",
                    Example: "- action: http\n  args: [\"GET\", \"https://secure.example.com/api\", {\"cert\": \"client.crt\", \"key\": \"client.key\", \"ca\": \"ca.crt\", \"Authorization\": \"Bearer ...\"}]\n  result: response"
                },
                {
                    Name: "http_get",
                    Description: "Perform HTTP GET request",
                    Example: "- action: http_get\n  args: [\"https://api.example.com/users\"]\n  result: response"
                },
                {
                    Name: "http_post",
                    Description: "Perform HTTP POST request",
                    Example: "- action: http_post\n  args: [\"https://api.example.com/users\", '{\"name\": \"John\"}']\n  result: response"
                },
                {
                    Name: "length",
                    Description: "Get length of string or array",
                    Example: "- action: length\n  args: [\"Hello World\"]\n  result: str_length"
                },
                {
                    Name: "log",
                    Description: "Log a message",
                    Example: "- action: log\n  args: [\"message\"]"
                },
                {
                    Name: "sleep",
                    Description: "Sleep for a duration",
                    Example: "- action: sleep\n  args: [2]"
                }
            ];
        }
    }
}

class RobogoSignatureHelpProvider implements vscode.SignatureHelpProvider {
    async provideSignatureHelp(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken,
        context: vscode.SignatureHelpContext
    ): Promise<vscode.SignatureHelp> {
        const signatureHelp = new vscode.SignatureHelp();
        const wordRange = document.getWordRangeAtPosition(position);
        if (!wordRange) {
            return signatureHelp;
        }

        const word = document.getText(wordRange);
        
        // Check if the word is an action
        try {
            const actions: RobogoAction[] = await this.getActions();
            const action = actions.find(a => a.Name === word);
            
            if (action) {
                const signature = new vscode.SignatureInformation(
                    `${action.Name}(args)`,
                    new vscode.MarkdownString(`**${action.Name}**\n\n${action.Description}\n\n**Example:**\n\`\`\`yaml\n${action.Example}\n\`\`\``)
                );
                
                // Add parameter information
                signature.parameters = [
                    new vscode.ParameterInformation('args', 'Arguments for the action'),
                    new vscode.ParameterInformation('result', 'Variable name to store the result (optional)')
                ];
                
                signatureHelp.signatures = [signature];
                signatureHelp.activeSignature = 0;
                signatureHelp.activeParameter = 0;
            }
        } catch (error) {
            console.error('Failed to get action signature:', error);
        }

        return signatureHelp;
    }

    private async getActions(): Promise<RobogoAction[]> {
        try {
            const config = vscode.workspace.getConfiguration('robogo');
            const executablePath = config.get<string>('executablePath', 'robogo');
            
            const { stdout } = await execAsync(`${executablePath} list --output json`);
            return JSON.parse(stdout);
        } catch (error) {
            console.error('Failed to get actions:', error);
            // Return built-in actions as fallback (matching current registry)
            return [
                {
                    Name: "assert",
                    Description: "Assert two values are equal",
                    Example: "- action: assert\n  args: [\"expected\", \"actual\", \"message\"]"
                },
                {
                    Name: "concat",
                    Description: "Concatenate strings",
                    Example: "- action: concat\n  args: [\"Hello\", \" \", \"World\"]\n  result: message"
                },
                {
                    Name: "get_random",
                    Description: "Get a random number",
                    Example: "- action: get_random\n  args: [100]\n  result: random_number"
                },
                {
                    Name: "get_time",
                    Description: "Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)",
                    Example: "- action: get_time\n  args: [\"iso\"]\n  result: timestamp"
                },
                {
                    Name: "http",
                    Description: "Perform HTTP request (GET, POST, PUT, DELETE, etc.). Supports client cert (cert/key) and custom CA (ca) options.",
                    Example: "- action: http\n  args: [\"GET\", \"https://secure.example.com/api\", {\"cert\": \"client.crt\", \"key\": \"client.key\", \"ca\": \"ca.crt\", \"Authorization\": \"Bearer ...\"}]\n  result: response"
                },
                {
                    Name: "http_get",
                    Description: "Perform HTTP GET request",
                    Example: "- action: http_get\n  args: [\"https://api.example.com/users\"]\n  result: response"
                },
                {
                    Name: "http_post",
                    Description: "Perform HTTP POST request",
                    Example: "- action: http_post\n  args: [\"https://api.example.com/users\", '{\"name\": \"John\"}']\n  result: response"
                },
                {
                    Name: "length",
                    Description: "Get length of string or array",
                    Example: "- action: length\n  args: [\"Hello World\"]\n  result: str_length"
                },
                {
                    Name: "log",
                    Description: "Log a message",
                    Example: "- action: log\n  args: [\"message\"]"
                },
                {
                    Name: "sleep",
                    Description: "Sleep for a duration",
                    Example: "- action: sleep\n  args: [2]"
                }
            ];
        }
    }
}

// Command implementations
async function runTest() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor found');
        return;
    }

    const document = editor.document;
    const fileName = document.fileName;
    
    // Check if it's a supported file type
    const supportedExtensions = ['.robogo', '.yaml', '.yml'];
    const fileExtension = fileName.substring(fileName.lastIndexOf('.'));
    
    if (!supportedExtensions.includes(fileExtension)) {
        vscode.window.showErrorMessage('Current file is not a supported Robogo test file');
        return;
    }

    try {
        const config = vscode.workspace.getConfiguration('robogo');
        const executablePath = config.get<string>('executablePath', 'robogo');
        const outputFormat = config.get<string>('outputFormat', 'console');

        // Show progress
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Running Robogo test...",
            cancellable: false
        }, async (progress) => {
            progress.report({ message: `Running ${fileName}` });

            const command = `${executablePath} run "${fileName}" --output ${outputFormat}`;
            const { stdout, stderr } = await execAsync(command);

            if (stderr) {
                console.error('Robogo stderr:', stderr);
            }

            // Show results
            if (outputFormat === 'json') {
                // For JSON output, show in a new document
                const doc = await vscode.workspace.openTextDocument({
                    content: stdout,
                    language: 'json'
                });
                await vscode.window.showTextDocument(doc);
            } else {
                // For console/markdown output, show in output channel
                const output = vscode.window.createOutputChannel('Robogo Test Results');
                output.appendLine(`Running: ${fileName}`);
                output.appendLine(stdout);
                output.show();
            }

            vscode.window.showInformationMessage(`Test completed: ${fileName}`);
        });

    } catch (error) {
        console.error('Failed to run test:', error);
        vscode.window.showErrorMessage(`Failed to run test: ${error}`);
    }
}

async function listActions() {
    try {
        const config = vscode.workspace.getConfiguration('robogo');
        const executablePath = config.get<string>('executablePath', 'robogo');

        const { stdout } = await execAsync(`${executablePath} list --output console`);
        
        // Show in output channel
        const output = vscode.window.createOutputChannel('Robogo Actions');
        output.appendLine('Available Robogo Actions:');
        output.appendLine(stdout);
        output.show();

    } catch (error) {
        console.error('Failed to list actions:', error);
        vscode.window.showErrorMessage(`Failed to list actions: ${error}`);
    }
}

export function deactivate() {} 