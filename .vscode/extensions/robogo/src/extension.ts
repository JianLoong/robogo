import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';
import * as yaml from 'js-yaml';

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

    // Register diagnostic provider for real-time validation
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('robogo');
    context.subscriptions.push(diagnosticCollection);

    const diagnosticProvider = new RobogoDiagnosticProvider(diagnosticCollection);
    context.subscriptions.push(diagnosticProvider);

    // Register commands
    const runTestCommand = vscode.commands.registerCommand('robogo.runTest', async () => {
        await runTest();
    });

    const listActionsCommand = vscode.commands.registerCommand('robogo.listActions', async () => {
        await listActions();
    });

    context.subscriptions.push(runTestCommand, listActionsCommand);
}

// Robogo Diagnostic Provider for real-time validation
class RobogoDiagnosticProvider {
    private diagnosticCollection: vscode.DiagnosticCollection;
    private disposables: vscode.Disposable[] = [];

    constructor(diagnosticCollection: vscode.DiagnosticCollection) {
        this.diagnosticCollection = diagnosticCollection;
        this.setupValidation();
    }

    private setupValidation() {
        // Validate on document change
        const changeDisposable = vscode.workspace.onDidChangeTextDocument((event) => {
            if (this.isRobogoFile(event.document)) {
                this.validateDocument(event.document);
            }
        });

        // Validate on document open
        const openDisposable = vscode.workspace.onDidOpenTextDocument((document) => {
            if (this.isRobogoFile(document)) {
                this.validateDocument(document);
            }
        });

        // Validate on document save
        const saveDisposable = vscode.workspace.onDidSaveTextDocument((document) => {
            if (this.isRobogoFile(document)) {
                this.validateDocument(document);
            }
        });

        this.disposables.push(changeDisposable, openDisposable, saveDisposable);
    }

    private isRobogoFile(document: vscode.TextDocument): boolean {
        return document.languageId === 'robogo' ||
            document.languageId === 'yaml' ||
            document.fileName.endsWith('.robogo') ||
            document.fileName.endsWith('.yaml') ||
            document.fileName.endsWith('.yml');
    }

    private async validateDocument(document: vscode.TextDocument) {
        const diagnostics: vscode.Diagnostic[] = [];

        try {
            // Parse YAML
            const text = document.getText();
            const parsed = yaml.load(text) as any;

            if (!parsed) {
                diagnostics.push(this.createDiagnostic(
                    new vscode.Range(0, 0, 0, 0),
                    'Empty or invalid YAML file',
                    vscode.DiagnosticSeverity.Error
                ));
            } else {
                // Validate test case structure
                this.validateTestCase(parsed, document, diagnostics);
            }
        } catch (error) {
            // YAML parsing error
            const errorMessage = error instanceof Error ? error.message : 'Unknown YAML error';
            const lineMatch = errorMessage.match(/line (\d+)/);
            const line = lineMatch ? parseInt(lineMatch[1]) - 1 : 0;

            diagnostics.push(this.createDiagnostic(
                new vscode.Range(line, 0, line, 0),
                `YAML parsing error: ${errorMessage}`,
                vscode.DiagnosticSeverity.Error
            ));
        }

        this.diagnosticCollection.set(document.uri, diagnostics);
    }

    private validateTestCase(testCase: any, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        // Check required fields
        if (!testCase.testcase) {
            diagnostics.push(this.createDiagnostic(
                new vscode.Range(0, 0, 0, 0),
                'Missing required field: testcase',
                vscode.DiagnosticSeverity.Error
            ));
        }

        if (!testCase.steps || !Array.isArray(testCase.steps)) {
            diagnostics.push(this.createDiagnostic(
                new vscode.Range(0, 0, 0, 0),
                'Missing or invalid steps array',
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        // Validate each step
        testCase.steps.forEach((step: any, index: number) => {
            this.validateStep(step, index, document, diagnostics);
        });
    }

    private validateStep(step: any, stepIndex: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const stepLine = this.findStepLine(stepIndex, document);
        const range = new vscode.Range(stepLine, 0, stepLine, 0);

        // Check if step has action or control flow
        const hasAction = step.action;
        const hasControlFlow = step.if || step.for || step.while;

        if (!hasAction && !hasControlFlow) {
            diagnostics.push(this.createDiagnostic(
                range,
                'Step must have either an action or control flow (if/for/while)',
                vscode.DiagnosticSeverity.Error
            ));
        }

        if (hasAction && hasControlFlow) {
            diagnostics.push(this.createDiagnostic(
                range,
                'Step cannot have both action and control flow',
                vscode.DiagnosticSeverity.Error
            ));
        }

        // Validate action if present
        if (hasAction) {
            this.validateAction(step, stepIndex, document, diagnostics);
        }

        // Validate control flow if present
        if (hasControlFlow) {
            this.validateControlFlow(step, range, diagnostics);
        }
    }

    private validateAction(step: any, stepIndex: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const action = step.action;
        const args = step.args;
        const stepLine = this.findStepLine(stepIndex, document);
        const actionLine = this.findStepFieldLine(stepIndex, 'action', document);
        const argsLine = this.findStepFieldLine(stepIndex, 'args', document);

        // Check if action is valid
        if (typeof action !== 'string' || action.trim() === '') {
            const range = new vscode.Range(actionLine, 0, actionLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Action must be a non-empty string',
                vscode.DiagnosticSeverity.Error
            ));
        }

        // Validate args
        if (args !== undefined && !Array.isArray(args)) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Args must be an array',
                vscode.DiagnosticSeverity.Error
            ));
        }

        // Validate specific actions
        if (action === 'assert') {
            this.validateAssertAction(args, argsLine, document, diagnostics);
        } else if (action === 'http' || action === 'http_get' || action === 'http_post') {
            this.validateHttpAction(action, args, argsLine, document, diagnostics);
        } else if (action === 'postgres') {
            this.validatePostgresAction(args, argsLine, document, diagnostics);
        } else if (action === 'get_time') {
            this.validateGetTimeAction(args, argsLine, document, diagnostics);
        } else if (action === 'get_random') {
            this.validateGetRandomAction(args, argsLine, document, diagnostics);
        } else if (action === 'sleep') {
            this.validateSleepAction(args, argsLine, document, diagnostics);
        }

        // Check for unknown actions
        const range = new vscode.Range(actionLine, 0, actionLine, 0);
        this.validateActionExistsSync(action, range, diagnostics);
    }

    private validateAssertAction(args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (!Array.isArray(args) || args.length < 3) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Assert action requires at least 3 arguments: value, operator, expected',
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        const operator = args[1];
        const validOperators = ['==', '!=', '>', '<', '>=', '<=', 'contains', 'not_contains', 'starts_with', 'ends_with', 'matches', 'not_matches', 'empty', 'not_empty', '%'];

        // Special handling for modulo operations: [value, %, divisor, operator, expected]
        if (operator === '%') {
            if (args.length < 5) {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    'Modulo assertion requires 5 arguments: value, %, divisor, operator, expected',
                    vscode.DiagnosticSeverity.Error
                ));
                return;
            }
            const moduloOperator = args[3];
            if (typeof moduloOperator !== 'string' || !validOperators.includes(moduloOperator)) {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    `Invalid modulo assertion operator: ${moduloOperator}. Valid operators: ${validOperators.join(', ')}`,
                    vscode.DiagnosticSeverity.Error
                ));
            }
            return;
        }

        if (typeof operator !== 'string' || !validOperators.includes(operator)) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                `Invalid assertion operator: ${operator}. Valid operators: ${validOperators.join(', ')}`,
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    private validateHttpAction(action: string, args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (!Array.isArray(args) || args.length === 0) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                `${action} action requires at least one argument (URL)`,
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        const url = args[0];
        if (typeof url !== 'string' || !url.startsWith('http')) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'First argument must be a valid HTTP URL',
                vscode.DiagnosticSeverity.Warning
            ));
        }
    }

    private validatePostgresAction(args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (!Array.isArray(args) || args.length < 2) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Postgres action requires at least 2 arguments: operation, connection_string',
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        const operation = args[0];
        const validOperations = ['connect', 'query', 'execute', 'close'];

        if (typeof operation !== 'string' || !validOperations.includes(operation)) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                `Invalid postgres operation: ${operation}. Valid operations: ${validOperations.join(', ')}`,
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    private validateGetTimeAction(args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (args && args.length > 0) {
            const format = args[0];
            const validFormats = ['iso', 'iso_date', 'iso_time', 'datetime', 'date', 'time', 'timestamp', 'unix', 'unix_ms'];

            if (typeof format !== 'string') {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    'Time format must be a string',
                    vscode.DiagnosticSeverity.Error
                ));
            } else if (!validFormats.includes(format)) {
                // Custom formats are supported, so this is just an info message
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    `Custom time format: ${format}. Predefined formats: ${validFormats.join(', ')}`,
                    vscode.DiagnosticSeverity.Information
                ));
            }
        }
    }

    private validateGetRandomAction(args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (!Array.isArray(args) || args.length < 1) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'get_random action requires at least one argument (max value) or two arguments (min, max)',
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        // Check if we have range arguments (min, max)
        if (args.length >= 2) {
            const [min, max] = args;
            if (typeof min !== 'number' || typeof max !== 'number') {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    'get_random min and max must be numbers',
                    vscode.DiagnosticSeverity.Error
                ));
            } else if (min > max) {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    'get_random min must be less than or equal to max',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        } else {
            // Single argument - backward compatibility (0 to max)
            const max = args[0];
            if (typeof max !== 'number' || max <= 0) {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    'get_random max value must be a positive number',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }
    }

    private validateSleepAction(args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (!Array.isArray(args) || args.length === 0) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Sleep action requires duration argument',
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        const duration = args[0];
        if (typeof duration !== 'number' || duration < 0) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Sleep duration must be a positive number',
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    private validateControlFlow(step: any, range: vscode.Range, diagnostics: vscode.Diagnostic[]) {
        // Validate if statement
        if (step.if) {
            if (!step.if.condition) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'If statement must have a condition',
                    vscode.DiagnosticSeverity.Error
                ));
            }
            if (!step.if.then || !Array.isArray(step.if.then)) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'If statement must have a then block with steps array',
                    vscode.DiagnosticSeverity.Error
                ));
            }
            // Validate else block if present
            if (step.if.else && !Array.isArray(step.if.else)) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'If statement else block must be a steps array',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }

        // Validate for loop
        if (step.for) {
            if (!step.for.condition) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'For loop must have a condition (range, array, or count)',
                    vscode.DiagnosticSeverity.Error
                ));
            }
            if (!step.for.steps || !Array.isArray(step.for.steps)) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'For loop must have a steps array',
                    vscode.DiagnosticSeverity.Error
                ));
            }
            // Validate max_iterations if present
            if (step.for.max_iterations && typeof step.for.max_iterations !== 'number') {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'For loop max_iterations must be a number',
                    vscode.DiagnosticSeverity.Warning
                ));
            }
        }

        // Validate while loop
        if (step.while) {
            if (!step.while.condition) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'While loop must have a condition',
                    vscode.DiagnosticSeverity.Error
                ));
            }
            if (!step.while.steps || !Array.isArray(step.while.steps)) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'While loop must have a steps array',
                    vscode.DiagnosticSeverity.Error
                ));
            }
            // Validate max_iterations (required for while loops)
            if (!step.while.max_iterations) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'While loop should have max_iterations to prevent infinite loops',
                    vscode.DiagnosticSeverity.Warning
                ));
            } else if (typeof step.while.max_iterations !== 'number') {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'While loop max_iterations must be a number',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }
    }

    private findStepLine(stepIndex: number, document: vscode.TextDocument): number {
        const text = document.getText();
        const lines = text.split('\n');
        let stepCount = 0;
        let inSteps = false;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();

            // Check if we're entering the steps section
            if (line === 'steps:' || line.startsWith('steps:')) {
                inSteps = true;
                continue;
            }

            // Only count steps when we're in the steps section
            if (inSteps && line.startsWith('-')) {
                if (stepCount === stepIndex) {
                    return i;
                }
                stepCount++;
            }
        }

        return 0;
    }

    private findStepFieldLine(stepIndex: number, fieldName: string, document: vscode.TextDocument): number {
        const text = document.getText();
        const lines = text.split('\n');
        let stepCount = 0;
        let inSteps = false;
        let inTargetStep = false;
        let currentIndent = 0;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmedLine = line.trim();

            // Skip empty lines
            if (trimmedLine === '') {
                continue;
            }

            // Check if we're entering the steps section
            if (trimmedLine === 'steps:' || trimmedLine.startsWith('steps:')) {
                inSteps = true;
                continue;
            }

            // Calculate indentation level
            const indent = line.length - line.trimStart().length;

            // Check if we're entering the target step (starts with - and proper indentation)
            if (inSteps && trimmedLine.startsWith('-') && indent <= 2) {
                if (stepCount === stepIndex) {
                    inTargetStep = true;
                    currentIndent = indent;
                } else {
                    inTargetStep = false;
                }
                stepCount++;
                continue;
            }

            // If we're in the target step, look for the field
            if (inTargetStep && trimmedLine.startsWith(fieldName + ':')) {
                return i;
            }

            // If we hit another step or section, we're no longer in the target step
            if (inTargetStep && (trimmedLine.startsWith('-') || (trimmedLine.includes(':') && indent <= currentIndent && !trimmedLine.startsWith('  ')))) {
                inTargetStep = false;
            }
        }

        // If we can't find the specific field, return the step line as fallback
        return this.findStepLine(stepIndex, document);
    }

    private createDiagnostic(range: vscode.Range, message: string, severity: vscode.DiagnosticSeverity): vscode.Diagnostic {
        return new vscode.Diagnostic(range, message, severity);
    }

    private validateActionExistsSync(action: string, range: vscode.Range, diagnostics: vscode.Diagnostic[]) {
        const validActions = [
            'log', 'sleep', 'assert', 'get_time', 'get_random', 'concat', 'length',
            'http', 'http_get', 'http_post', 'postgres', 'variable', 'control'
        ];

        if (!validActions.includes(action)) {
            diagnostics.push(this.createDiagnostic(
                range,
                `Unknown action: ${action}. Valid actions: ${validActions.join(', ')}`,
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    dispose() {
        this.disposables.forEach(d => d.dispose());
        this.diagnosticCollection.dispose();
    }
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

        // Check if we're at the start of a step (after "-")
        if (linePrefix.trim().endsWith('-') || linePrefix.trim() === '') {
            // Add action field completion
            items.push(
                this.createCompletionItem('name', 'Step name (strongly recommended for clarity)', vscode.CompletionItemKind.Property),
                this.createCompletionItem('action', 'Action to execute', vscode.CompletionItemKind.Property),
                this.createCompletionItem('args', 'Arguments for the action', vscode.CompletionItemKind.Property),
                this.createCompletionItem('result', 'Variable name to store the result', vscode.CompletionItemKind.Property),
                this.createCompletionItem('if', 'If statement for conditional execution', vscode.CompletionItemKind.Property),
                this.createCompletionItem('for', 'For loop for repeated execution', vscode.CompletionItemKind.Property),
                this.createCompletionItem('while', 'While loop for conditional repetition', vscode.CompletionItemKind.Property),
                this.createCompletionItem('continue_on_failure', 'Continue execution even if this step fails', vscode.CompletionItemKind.Property),
                this.createCompletionItem('verbose', 'Enable verbose output for this step', vscode.CompletionItemKind.Property),
                this.createCompletionItem('retry', 'Retry configuration for this step', vscode.CompletionItemKind.Property)
            );
        }

        // Add common YAML/Robogo structure completions
        if (linePrefix.trim() === '' || linePrefix.trim().endsWith('-')) {
            items.push(
                this.createCompletionItem('testcase', 'Test case name', vscode.CompletionItemKind.Property),
                this.createCompletionItem('description', 'Test case description', vscode.CompletionItemKind.Property),
                this.createCompletionItem('variables', 'Variables section with vars and secrets', vscode.CompletionItemKind.Property),
                this.createCompletionItem('steps', 'Test steps array', vscode.CompletionItemKind.Property),
                this.createCompletionItem('timeout', 'Test timeout duration', vscode.CompletionItemKind.Property),
                this.createCompletionItem('verbose', 'Global verbosity setting', vscode.CompletionItemKind.Property)
            );
        }

        // Add control flow structure completions
        if (linePrefix.includes('if:') || linePrefix.includes('for:') || linePrefix.includes('while:')) {
            items.push(
                this.createCompletionItem('condition', 'Condition for control flow', vscode.CompletionItemKind.Property),
                this.createCompletionItem('then', 'Steps to execute when condition is true', vscode.CompletionItemKind.Property),
                this.createCompletionItem('else', 'Steps to execute when condition is false', vscode.CompletionItemKind.Property),
                this.createCompletionItem('steps', 'Steps to execute in loop', vscode.CompletionItemKind.Property),
                this.createCompletionItem('max_iterations', 'Maximum iterations to prevent infinite loops', vscode.CompletionItemKind.Property)
            );
        }

        return items;
    }

    private createCompletionItem(label: string, detail: string, kind: vscode.CompletionItemKind): vscode.CompletionItem {
        const item = new vscode.CompletionItem(label, kind);
        item.detail = detail;
        return item;
    }

    private async getActions(): Promise<RobogoAction[]> {
        return [
            {
                Name: 'log',
                Description: 'Log a message to the console',
                Example: '- action: log\n  args: ["Hello, World!"]'
            },
            {
                Name: 'sleep',
                Description: 'Sleep for a specified duration in seconds',
                Example: '- action: sleep\n  args: [2]'
            },
            {
                Name: 'assert',
                Description: 'Assert a condition using comparison operators',
                Example: '- action: assert\n  args: ["value", "==", "expected", "Custom message"]'
            },
            {
                Name: 'get_time',
                Description: 'Get current timestamp with optional format',
                Example: '- action: get_time\n  args: ["iso"]\n  result: timestamp'
            },
            {
                Name: 'get_random',
                Description: 'Generate a random number (0 to max) or in range (min to max)',
                Example: '- action: get_random\n  args: [100]\n  result: random_number'
            },
            {
                Name: 'concat',
                Description: 'Concatenate multiple strings',
                Example: '- action: concat\n  args: ["Hello", " ", "World"]\n  result: greeting'
            },
            {
                Name: 'length',
                Description: 'Get the length of a string or array',
                Example: '- action: length\n  args: ["Hello World"]\n  result: str_length'
            },
            {
                Name: 'http',
                Description: 'Perform HTTP requests with full control',
                Example: '- action: http\n  args: ["GET", "https://api.example.com/data"]\n  result: response'
            },
            {
                Name: 'http_get',
                Description: 'Perform simplified GET requests',
                Example: '- action: http_get\n  args: ["https://api.example.com/data"]\n  result: response'
            },
            {
                Name: 'http_post',
                Description: 'Perform simplified POST requests',
                Example: '- action: http_post\n  args: ["https://api.example.com/data", "{\\"key\\": \\"value\\"}"]\n  result: response'
            },
            {
                Name: 'postgres',
                Description: 'Perform PostgreSQL database operations',
                Example: '- action: postgres\n  args: ["query", "postgresql://user:pass@localhost/db", "SELECT * FROM users"]\n  result: result'
            }
        ];
    }
}

class RobogoHoverProvider implements vscode.HoverProvider {
    async provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Hover | undefined> {
        const line = document.lineAt(position);
        const text = line.text;

        // Check if we're hovering over an action name
        const actionMatch = text.match(/action:\s*(\w+)/);
        if (actionMatch) {
            const actionName = actionMatch[1];
            const actions = await this.getActions();
            const action = actions.find(a => a.Name === actionName);

            if (action) {
                const markdown = new vscode.MarkdownString();
                markdown.appendMarkdown(`**${action.Name}**\n\n`);
                markdown.appendMarkdown(`${action.Description}\n\n`);
                markdown.appendMarkdown(`**Example:**\n\`\`\`yaml\n${action.Example}\n\`\`\``);
                return new vscode.Hover(markdown);
            }
        }

        return undefined;
    }

    private async getActions(): Promise<RobogoAction[]> {
        return [
            {
                Name: 'log',
                Description: 'Log a message to the console',
                Example: '- action: log\n  args: ["Hello, World!"]'
            },
            {
                Name: 'sleep',
                Description: 'Sleep for a specified duration in seconds',
                Example: '- action: sleep\n  args: [2]'
            },
            {
                Name: 'assert',
                Description: 'Assert a condition using comparison operators',
                Example: '- action: assert\n  args: ["value", "==", "expected", "Custom message"]'
            },
            {
                Name: 'get_time',
                Description: 'Get current timestamp with optional format',
                Example: '- action: get_time\n  args: ["iso"]\n  result: timestamp'
            },
            {
                Name: 'get_random',
                Description: 'Generate a random number (0 to max) or in range (min to max)',
                Example: '- action: get_random\n  args: [100]\n  result: random_number'
            },
            {
                Name: 'concat',
                Description: 'Concatenate multiple strings',
                Example: '- action: concat\n  args: ["Hello", " ", "World"]\n  result: greeting'
            },
            {
                Name: 'length',
                Description: 'Get the length of a string or array',
                Example: '- action: length\n  args: ["Hello World"]\n  result: str_length'
            },
            {
                Name: 'http',
                Description: 'Perform HTTP requests with full control',
                Example: '- action: http\n  args: ["GET", "https://api.example.com/data"]\n  result: response'
            },
            {
                Name: 'http_get',
                Description: 'Perform simplified GET requests',
                Example: '- action: http_get\n  args: ["https://api.example.com/data"]\n  result: response'
            },
            {
                Name: 'http_post',
                Description: 'Perform simplified POST requests',
                Example: '- action: http_post\n  args: ["https://api.example.com/data", "{\\"key\\": \\"value\\"}"]\n  result: response'
            },
            {
                Name: 'postgres',
                Description: 'Perform PostgreSQL database operations',
                Example: '- action: postgres\n  args: ["query", "postgresql://user:pass@localhost/db", "SELECT * FROM users"]\n  result: result'
            }
        ];
    }
}

async function runTest() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const document = editor.document;
    if (!document.fileName.endsWith('.robogo') && !document.fileName.endsWith('.yaml') && !document.fileName.endsWith('.yml')) {
        vscode.window.showErrorMessage('Current file is not a Robogo test file');
        return;
    }

    try {
        const config = vscode.workspace.getConfiguration('robogo');
        const executablePath = config.get<string>('executablePath', 'robogo');

        const terminal = vscode.window.createTerminal('Robogo Test');
        terminal.show();
        terminal.sendText(`${executablePath} run "${document.fileName}"`);

        vscode.window.showInformationMessage(`Running test: ${document.fileName}`);
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to run test: ${error}`);
    }
}

async function listActions() {
    try {
        const config = vscode.workspace.getConfiguration('robogo');
        const executablePath = config.get<string>('executablePath', 'robogo');

        const { stdout } = await execAsync(`${executablePath} list-actions`);

        const document = await vscode.workspace.openTextDocument({
            content: stdout,
            language: 'markdown'
        });

        await vscode.window.showTextDocument(document);
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to list actions: ${error}`);
    }
}

export function deactivate() { } 