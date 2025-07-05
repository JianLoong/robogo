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

        // Validate variables section
        if (testCase.variables) {
            this.validateVariables(testCase.variables, document, diagnostics);
        }

        // Validate variable references in steps
        this.validateVariableReferences(testCase.steps, testCase.variables, document, diagnostics);
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

        // Validate retry configuration
        if (step.retry) {
            this.validateRetry(step.retry, range, diagnostics);
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
        } else if (action === 'random') {
            this.validateRandomAction(args, argsLine, document, diagnostics);
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
            
            if (typeof format !== 'string' || !validFormats.includes(format)) {
                const range = new vscode.Range(argsLine, 0, argsLine, 0);
                diagnostics.push(this.createDiagnostic(
                    range,
                    `Invalid time format: ${format}. Valid formats: ${validFormats.join(', ')}`,
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }
    }

    private validateRandomAction(args: any[], argsLine: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (!Array.isArray(args) || args.length < 2) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Random action requires at least 2 arguments: min, max',
                vscode.DiagnosticSeverity.Error
            ));
            return;
        }

        const [min, max] = args;
        if (typeof min !== 'number' || typeof max !== 'number') {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Random min and max must be numbers',
                vscode.DiagnosticSeverity.Error
            ));
        } else if (min >= max) {
            const range = new vscode.Range(argsLine, 0, argsLine, 0);
            diagnostics.push(this.createDiagnostic(
                range,
                'Random min must be less than max',
                vscode.DiagnosticSeverity.Error
            ));
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
        if (step.if) {
            if (!step.if.condition || !step.if.then) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'If statement must have condition and then blocks',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }

        if (step.for) {
            if (!step.for.condition || !step.for.steps) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'For loop must have condition and steps blocks',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }

        if (step.while) {
            if (!step.while.condition || !step.while.steps) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    'While loop must have condition and steps blocks',
                    vscode.DiagnosticSeverity.Error
                ));
            }
        }
    }

    private validateRetry(retry: any, range: vscode.Range, diagnostics: vscode.Diagnostic[]) {
        if (retry.attempts && (typeof retry.attempts !== 'number' || retry.attempts < 1)) {
            diagnostics.push(this.createDiagnostic(
                range,
                'Retry attempts must be a positive number',
                vscode.DiagnosticSeverity.Error
            ));
        }

        if (retry.backoff && !['fixed', 'linear', 'exponential'].includes(retry.backoff)) {
            diagnostics.push(this.createDiagnostic(
                range,
                'Retry backoff must be: fixed, linear, or exponential',
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    private validateVariables(variables: any, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        if (variables.secrets) {
            Object.entries(variables.secrets).forEach(([name, secret]: [string, any]) => {
                if (typeof secret === 'object' && secret.file && secret.value) {
                    const line = this.findVariableLine(name, document);
                    diagnostics.push(this.createDiagnostic(
                        new vscode.Range(line, 0, line, 0),
                        `Secret '${name}' has both file and value - value takes precedence`,
                        vscode.DiagnosticSeverity.Warning
                    ));
                }
            });
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

    private findVariableLine(varName: string, document: vscode.TextDocument): number {
        const text = document.getText();
        const lines = text.split('\n');
        
        for (let i = 0; i < lines.length; i++) {
            if (lines[i].includes(varName + ':')) {
                return i;
            }
        }
        
        return 0;
    }

    private validateVariableReferences(steps: any[], variables: any, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const definedVars = new Set<string>();
        
        // Collect defined variables
        if (variables) {
            if (variables.vars) {
                Object.keys(variables.vars).forEach(varName => definedVars.add(varName));
            }
            if (variables.secrets) {
                Object.keys(variables.secrets).forEach(varName => definedVars.add(varName));
            }
        }

        // Check variable references in steps
        steps.forEach((step: any, stepIndex: number) => {
            this.checkStepVariableReferences(step, definedVars, stepIndex, document, diagnostics);
        });
    }

    private checkStepVariableReferences(step: any, definedVars: Set<string>, stepIndex: number, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const stepText = JSON.stringify(step);
        const varRegex = /\$\{([^}]+)\}/g;
        let match;

        while ((match = varRegex.exec(stepText)) !== null) {
            const varName = match[1];
            // Handle dot notation (e.g., response.status_code)
            const baseVarName = varName.split('.')[0];
            
            if (!definedVars.has(baseVarName)) {
                const stepLine = this.findStepLine(stepIndex, document);
                diagnostics.push(this.createDiagnostic(
                    new vscode.Range(stepLine, 0, stepLine, 0),
                    `Undefined variable: ${baseVarName}`,
                    vscode.DiagnosticSeverity.Warning
                ));
            }
        }
    }

    private createDiagnostic(range: vscode.Range, message: string, severity: vscode.DiagnosticSeverity): vscode.Diagnostic {
        const diagnostic = new vscode.Diagnostic(range, message, severity);
        diagnostic.source = 'robogo';
        return diagnostic;
    }

    private validateActionExistsSync(action: string, range: vscode.Range, diagnostics: vscode.Diagnostic[]) {
        // Use a predefined list of valid actions for immediate validation
        const validActions = [
            'assert', 'http', 'http_get', 'http_post', 'postgres', 'get_time', 
            'random', 'sleep', 'log', 'set_variable', 'get_variable', 'list_variables',
            'get_secret', 'set_secret', 'list_secrets', 'get_swift_message', 'validate_swift'
        ];
        
        if (!validActions.includes(action)) {
            diagnostics.push(this.createDiagnostic(
                range,
                `Unknown action: ${action}. Available actions: ${validActions.join(', ')}`,
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    private async validateActionExists(action: string, range: vscode.Range, diagnostics: vscode.Diagnostic[]) {
        try {
            const actions = await this.getAvailableActions();
            if (!actions.includes(action)) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    `Unknown action: ${action}. Available actions: ${actions.slice(0, 10).join(', ')}${actions.length > 10 ? '...' : ''}`,
                    vscode.DiagnosticSeverity.Warning
                ));
            }
        } catch (error) {
            // Silently fail - action validation is not critical
        }
    }

    private async getAvailableActions(): Promise<string[]> {
        try {
            const config = vscode.workspace.getConfiguration('robogo');
            const executablePath = config.get<string>('executablePath', 'robogo');
            
            const { stdout } = await execAsync(`${executablePath} list-actions`);
            const lines = stdout.split('\n');
            return lines
                .filter(line => line.trim() && !line.startsWith('Available'))
                .map(line => line.split(/\s+/)[0])
                .filter(Boolean);
        } catch (error) {
            // Return common actions if we can't get the list
            return [
                'assert', 'http', 'http_get', 'http_post', 'postgres', 'get_time', 
                'random', 'sleep', 'log', 'set_variable', 'get_variable', 'list_variables'
            ];
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

        // Autocomplete PostgreSQL operations in postgres action context
        if (linePrefix.includes('args:') && this.isInPostgresContext(document, position)) {
            const postgresOps = [
                { name: 'connect', desc: 'Connect to PostgreSQL database' },
                { name: 'query', desc: 'Execute a SELECT query' },
                { name: 'execute', desc: 'Execute INSERT/UPDATE/DELETE statement' },
                { name: 'close', desc: 'Close database connection' }
            ];
            for (const op of postgresOps) {
                const item = new vscode.CompletionItem(op.name, vscode.CompletionItemKind.Function);
                item.detail = `PostgreSQL operation: ${op.desc}`;
                item.insertText = `"${op.name}"`;
                items.push(item);
            }
        }

        // Autocomplete variable operations in variable action context
        if (linePrefix.includes('args:') && this.isInVariableContext(document, position)) {
            const variableOps = [
                { name: 'set_variable', desc: 'Set a variable to a value' },
                { name: 'get_variable', desc: 'Get a variable value' },
                { name: 'list_variables', desc: 'List all variables' }
            ];
            for (const op of variableOps) {
                const item = new vscode.CompletionItem(op.name, vscode.CompletionItemKind.Function);
                item.detail = `Variable operation: ${op.desc}`;
                item.insertText = `"${op.name}"`;
                items.push(item);
            }
        }

        // Autocomplete assertion operators in assert action context
        if (linePrefix.includes('args:') && this.isInAssertContext(document, position)) {
            const assertionOps = [
                { name: '==', desc: 'Equal to' },
                { name: '!=', desc: 'Not equal to' },
                { name: '>', desc: 'Greater than' },
                { name: '<', desc: 'Less than' },
                { name: '>=', desc: 'Greater than or equal to' },
                { name: '<=', desc: 'Less than or equal to' },
                { name: 'contains', desc: 'String contains substring' },
                { name: 'not_contains', desc: 'String does not contain substring' },
                { name: 'starts_with', desc: 'String starts with prefix' },
                { name: 'ends_with', desc: 'String ends with suffix' }
            ];
            for (const op of assertionOps) {
                const item = new vscode.CompletionItem(op.name, vscode.CompletionItemKind.Operator);
                item.detail = `Assertion operator: ${op.desc}`;
                item.insertText = `"${op.name}"`;
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
                this.createCompletionItem('variables', 'Test case variables and secrets', vscode.CompletionItemKind.Property),
                this.createCompletionItem('verbose', 'Global verbosity setting (true/false or basic/detailed/debug)', vscode.CompletionItemKind.Property),
                this.createCompletionItem('steps', 'List of test steps', vscode.CompletionItemKind.Property)
            );
        }

        // Add variables section completions
        if (linePrefix.includes('variables:') || linePrefix.includes('vars:') || linePrefix.includes('secrets:')) {
            items.push(
                this.createCompletionItem('vars', 'Regular variables', vscode.CompletionItemKind.Property),
                this.createCompletionItem('secrets', 'Secret variables', vscode.CompletionItemKind.Property)
            );
        }

        // Add step structure completions
        if (linePrefix.includes('steps:') || linePrefix.includes('- action:')) {
            items.push(
                this.createCompletionItem('name', 'Step name (strongly recommended for clarity)', vscode.CompletionItemKind.Property),
                this.createCompletionItem('action', 'Action to execute', vscode.CompletionItemKind.Property),
                this.createCompletionItem('args', 'Arguments for the action', vscode.CompletionItemKind.Property),
                this.createCompletionItem('result', 'Variable name to store the result', vscode.CompletionItemKind.Property),
                this.createCompletionItem('verbose', 'Enable verbose output (true/false or basic/detailed/debug)', vscode.CompletionItemKind.Property)
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

        // Verbosity value completion
        if (linePrefix.includes('verbose:')) {
            const verbosityOptions = [
                { value: 'true', desc: 'Enable basic verbose output' },
                { value: 'false', desc: 'Disable verbose output' },
                { value: '"basic"', desc: 'Basic verbose output (action + duration)' },
                { value: '"detailed"', desc: 'Detailed verbose output (args + duration + output)' },
                { value: '"debug"', desc: 'Debug verbose output (all details + verbosity level)' }
            ];
            
            for (const option of verbosityOptions) {
                const item = new vscode.CompletionItem(option.value, vscode.CompletionItemKind.Value);
                item.detail = `Verbosity: ${option.desc}`;
                item.insertText = option.value;
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
            let executablePath = config.get<string>('executablePath', 'robogo');
            
            // Handle Windows executable extension
            if (process.platform === 'win32' && !executablePath.endsWith('.exe')) {
                executablePath = `${executablePath}.exe`;
            }
            
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
                },
                {
                    Name: "postgres",
                    Description: "PostgreSQL database operations (query, execute, connect, close)",
                    Example: "- action: postgres\n  args: [\"query\", \"postgres://user:pass@localhost/db\", \"SELECT * FROM users\"]\n  result: query_result"
                },
                {
                    Name: "variable",
                    Description: "Variable management operations (set_variable, get_variable, list_variables)",
                    Example: "- action: variable\n  args: [\"set_variable\", \"my_var\", \"my_value\"]\n  result: set_result"
                },
                {
                    Name: "control",
                    Description: "Control flow operations (if, for, while)",
                    Example: "- action: control\n  args: [\"if\", \"condition\"]\n  result: condition_result"
                }
            ];
        }
    }

    // Extract variable names from the document (from result fields and variables section)
    private extractVariableNames(document: vscode.TextDocument): string[] {
        const text = document.getText();
        const variables = new Set<string>();
        
        // Find variables from result fields
        const resultRegex = /result:\s*([a-zA-Z_][a-zA-Z0-9_]*)/g;
        let match;
        while ((match = resultRegex.exec(text)) !== null) {
            variables.add(match[1]);
        }
        
        // Find variables from variables section
        const varsRegex = /vars:\s*\n\s*([a-zA-Z_][a-zA-Z0-9_]*):/g;
        while ((match = varsRegex.exec(text)) !== null) {
            variables.add(match[1]);
        }
        
        // Find variables from secrets section
        const secretsRegex = /secrets:\s*\n\s*([a-zA-Z_][a-zA-Z0-9_]*):/g;
        while ((match = secretsRegex.exec(text)) !== null) {
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

    // Check if we're in a PostgreSQL action context
    private isInPostgresContext(document: vscode.TextDocument, position: vscode.Position): boolean {
        // Check current line and previous few lines for postgres action
        for (let i = Math.max(0, position.line - 3); i <= position.line; i++) {
            const line = document.lineAt(i).text;
            if (line.includes('action: postgres')) {
                return true;
            }
        }
        return false;
    }

    // Check if we're in a variable action context
    private isInVariableContext(document: vscode.TextDocument, position: vscode.Position): boolean {
        // Check current line and previous few lines for variable action
        for (let i = Math.max(0, position.line - 3); i <= position.line; i++) {
            const line = document.lineAt(i).text;
            if (line.includes('action: variable')) {
                return true;
            }
        }
        return false;
    }

    // Check if we're in an assert action context
    private isInAssertContext(document: vscode.TextDocument, position: vscode.Position): boolean {
        // Check current line and previous few lines for assert action
        for (let i = Math.max(0, position.line - 3); i <= position.line; i++) {
            const line = document.lineAt(i).text;
            if (line.includes('action: assert')) {
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
                    Description: "Assert a condition using comparison operators (==, !=, >, <, >=, <=, contains, starts_with, ends_with)",
                    Example: "- action: assert\n  args:\n    - \"value\"\n    - \">\"\n    - \"0\"\n    - \"Value should be positive\""
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
            let executablePath = config.get<string>('executablePath', 'robogo');
            
            // Handle Windows executable extension
            if (process.platform === 'win32' && !executablePath.endsWith('.exe')) {
                executablePath = `${executablePath}.exe`;
            }
            
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