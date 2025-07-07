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

    const runTestParallelCommand = vscode.commands.registerCommand('robogo.runTestParallel', async () => {
        await runTestParallel();
    });

    const listActionsCommand = vscode.commands.registerCommand('robogo.listActions', async () => {
        await listActions();
    });

    const validateTDMCommand = vscode.commands.registerCommand('robogo.validateTDM', async () => {
        await validateTDM();
    });

    const generateTemplateCommand = vscode.commands.registerCommand('robogo.generateTemplate', async () => {
        await generateTemplate();
    });

    const runWithOutputCommand = vscode.commands.registerCommand('robogo.runWithOutput', async () => {
        await runWithOutput();
    });

    const runTestSuiteCommand = vscode.commands.registerCommand('robogo.runTestSuite', async () => {
        await runTestSuite();
    });

    const showDocumentationCommand = vscode.commands.registerCommand('robogo.showDocumentation', async () => {
        await showDocumentation();
    });

    const validateSyntaxCommand = vscode.commands.registerCommand('robogo.validateSyntax', async () => {
        await validateSyntax();
    });

    context.subscriptions.push(
        runTestCommand,
        runTestParallelCommand,
        runTestSuiteCommand,
        listActionsCommand,
        validateTDMCommand,
        generateTemplateCommand,
        runWithOutputCommand,
        showDocumentationCommand,
        validateSyntaxCommand
    );
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
                // Distinguish between testsuite and testcase
                if (parsed.testsuite) {
                    this.validateTestSuite(parsed, document, diagnostics);
                } else if (parsed.testcase) {
                    this.validateTestCase(parsed, document, diagnostics);
                } else {
                    diagnostics.push(this.createDiagnostic(
                        new vscode.Range(0, 0, 0, 0),
                        'Missing required field: testcase or testsuite',
                        vscode.DiagnosticSeverity.Error
                    ));
                }
            }
        } catch (error) {
            // YAML parsing error
            const errorMessage = error instanceof Error ? error.message : 'Unknown YAML error';
            const lineMatch = errorMessage.match(/line (\d+)/);
            const line = lineMatch && lineMatch[1] ? parseInt(lineMatch[1]) - 1 : 0;

            diagnostics.push(this.createDiagnostic(
                new vscode.Range(line, 0, line, 0),
                `YAML parsing error: ${errorMessage}`,
                vscode.DiagnosticSeverity.Error
            ));
        }

        this.diagnosticCollection.set(document.uri, diagnostics);
    }

    // Add this new method for suite validation
    private validateTestSuite(suite: any, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        // Require 'testcases' as an array
        if (!suite.testcases || !Array.isArray(suite.testcases)) {
            diagnostics.push(this.createDiagnostic(
                new vscode.Range(0, 0, 0, 0),
                'Missing or invalid testcases array in testsuite',
                vscode.DiagnosticSeverity.Error
            ));
        }
        // Optionally, validate other suite-level fields here (setup, teardown, variables, etc.)
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

        // Validate TDM structure if present
        if (testCase.data_management) {
            this.validateTDMStructure(testCase.data_management, document, diagnostics);
        }

        // Validate environments if present
        if (testCase.environments) {
            this.validateEnvironments(testCase.environments, document, diagnostics);
        }

        // Validate variables if present
        if (testCase.variables) {
            this.validateVariables(testCase.variables, document, diagnostics);
        }

        // Validate each step
        testCase.steps.forEach((step: any, index: number) => {
            this.validateStep(step, index, document, diagnostics);
        });
    }

    private validateTDMStructure(tdm: any, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const tdmLine = this.findTopLevelFieldLine('data_management', document);
        const range = new vscode.Range(tdmLine, 0, tdmLine, 0);

        // Validate data_sets if present
        if (tdm.data_sets && Array.isArray(tdm.data_sets)) {
            tdm.data_sets.forEach((ds: any, index: number) => {
                if (!ds.name) {
                    diagnostics.push(this.createDiagnostic(
                        range,
                        `Data set ${index + 1} missing required field: name`,
                        vscode.DiagnosticSeverity.Error
                    ));
                }
                if (!ds.data || typeof ds.data !== 'object') {
                    diagnostics.push(this.createDiagnostic(
                        range,
                        `Data set ${ds.name || index + 1} missing required field: data`,
                        vscode.DiagnosticSeverity.Error
                    ));
                }
            });
        }

        // Validate validation rules if present
        if (tdm.validation && Array.isArray(tdm.validation)) {
            tdm.validation.forEach((validation: any, index: number) => {
                if (!validation.name) {
                    diagnostics.push(this.createDiagnostic(
                        range,
                        `Validation rule ${index + 1} missing required field: name`,
                        vscode.DiagnosticSeverity.Error
                    ));
                }
                if (!validation.type) {
                    diagnostics.push(this.createDiagnostic(
                        range,
                        `Validation rule ${validation.name || index + 1} missing required field: type`,
                        vscode.DiagnosticSeverity.Error
                    ));
                }
                if (!validation.field) {
                    diagnostics.push(this.createDiagnostic(
                        range,
                        `Validation rule ${validation.name || index + 1} missing required field: field`,
                        vscode.DiagnosticSeverity.Error
                    ));
                }
            });
        }
    }

    private validateEnvironments(environments: any[], document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const envLine = this.findTopLevelFieldLine('environments', document);
        const range = new vscode.Range(envLine, 0, envLine, 0);

        environments.forEach((env: any, index: number) => {
            if (!env.name) {
                diagnostics.push(this.createDiagnostic(
                    range,
                    `Environment ${index + 1} missing required field: name`,
                    vscode.DiagnosticSeverity.Error
                ));
            }
        });
    }

    private validateVariables(variables: any, document: vscode.TextDocument, diagnostics: vscode.Diagnostic[]) {
        const varsLine = this.findTopLevelFieldLine('variables', document);
        const range = new vscode.Range(varsLine, 0, varsLine, 0);

        // Validate secrets if present
        if (variables.secrets && typeof variables.secrets === 'object') {
            Object.keys(variables.secrets).forEach(secretName => {
                const secret = variables.secrets[secretName];
                if (!secret.value && !secret.file) {
                    diagnostics.push(this.createDiagnostic(
                        range,
                        `Secret '${secretName}' must have either 'value' or 'file' field`,
                        vscode.DiagnosticSeverity.Error
                    ));
                }
            });
        }
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
                vscode.DiagnosticSeverity.Warning
            ));
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
            const line = lines[i] ?? '';
            const trimmedLine = line.trim();

            // Check if we're entering the steps section
            if (trimmedLine === 'steps:' || trimmedLine.startsWith('steps:')) {
                inSteps = true;
                continue;
            }

            // Only count steps when we're in the steps section
            if (inSteps && trimmedLine.startsWith('-')) {
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
            const line = lines[i] ?? '';
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

    private findTopLevelFieldLine(fieldName: string, document: vscode.TextDocument): number {
        // Find top-level field (not indented or at root level)
        for (let i = 0; i < document.lineCount; i++) {
            const line = document.lineAt(i).text;
            const trimmed = line.trim();
            if (trimmed.startsWith(fieldName + ':') && (line.startsWith(fieldName + ':') || line.match(new RegExp(`^\\s*${fieldName}:`)))) {
                return i;
            }
        }
        return 0; // Default to first line if not found
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
    SupportedTypes?: string[];
    UseCases?: string[];
    Examples?: string[];
    Operators?: string[];
    Formats?: string[];
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
        const line = document.lineAt(position);
        const text = line.text;

        // Check if we're completing an action name
        if (text.match(/action:\s*$/)) {
            const actions = await this.getActions();
            actions.forEach(action => {
                const item = new vscode.CompletionItem(action.Name, vscode.CompletionItemKind.Function);
                item.detail = action.Description;
                const documentation = new vscode.MarkdownString();

                // Add parameters if available
                if (action.Parameters && action.Parameters.length > 0) {
                    documentation.appendMarkdown(`**Parameters:**\n`);
                    action.Parameters.forEach(param => {
                        const required = param.required ? 'required' : 'optional';
                        const defaultValue = param.default ? ` (default: ${param.default})` : '';
                        documentation.appendMarkdown(`- \`${param.name}\` (${param.type}, ${required})${defaultValue}: ${param.description}\n`);
                    });
                    documentation.appendMarkdown(`\n`);
                }

                // Add return value if available
                if (action.Returns) {
                    documentation.appendMarkdown(`**Returns:** ${action.Returns}\n\n`);
                }

                // Add examples
                if (action.Examples && action.Examples.length > 0) {
                    documentation.appendMarkdown(`**Examples:**\n`);
                    action.Examples.forEach(example => {
                        documentation.appendMarkdown(`\`\`\`yaml\n${example}\n\`\`\`\n`);
                    });
                } else {
                    documentation.appendMarkdown(`**Example:**\n\`\`\`yaml\n${action.Example}\n\`\`\`\n`);
                }

                // Add notes if available
                if (action.Notes) {
                    documentation.appendMarkdown(`**Notes:** ${action.Notes}\n`);
                }

                item.documentation = documentation;

                item.insertText = action.Name;
                items.push(item);
            });
        }

        // Check if we're completing field names
        if (text.match(/^\s*$/)) {
            const fieldItems = [
                { name: 'testcase', detail: 'Test case name (required)', documentation: 'The name of the test case. This is a required field that identifies the test.' },
                { name: 'description', detail: 'Test case description', documentation: 'Optional description of what the test case does.' },
                { name: 'variables', detail: 'Variables and secrets', documentation: 'Define variables and secrets for use throughout the test case.' },
                { name: 'data_management', detail: 'Test Data Management', documentation: 'Test Data Management configuration including data sets, validation, and environments.' },
                { name: 'steps', detail: 'Test steps array (required)', documentation: 'Array of test steps to execute. Each step can be an action or control flow.' },
                { name: 'name', detail: 'Step or element name', documentation: 'Name identifier for the current element (step, data set, environment, etc.).' },
                { name: 'action', detail: 'Action to execute', documentation: 'The action to execute (log, http, assert, etc.).' },
                { name: 'args', detail: 'Action arguments', documentation: 'Arguments to pass to the action.' },
                { name: 'result', detail: 'Result variable', documentation: 'Variable name to store the action result.' },
                { name: 'if', detail: 'If statement', documentation: 'Conditional execution block with then/else branches.' },
                { name: 'for', detail: 'For loop', documentation: 'Loop execution block with range, array, or count.' },
                { name: 'while', detail: 'While loop', documentation: 'Conditional loop with condition and max iterations.' },
                { name: 'continue_on_failure', detail: 'Continue on failure', documentation: 'Continue test execution even if this step fails.' },
                { name: 'verbose', detail: 'Verbose output', documentation: 'Enable detailed output for this step.' },
                { name: 'retry', detail: 'Retry configuration', documentation: 'Retry settings for this step.' }
            ];

            fieldItems.forEach(field => {
                const item = new vscode.CompletionItem(field.name, vscode.CompletionItemKind.Field);
                item.detail = field.detail;
                item.documentation = new vscode.MarkdownString(field.documentation);
                item.insertText = `${field.name}: `;
                items.push(item);
            });
        }

        // Check if we're completing TDM fields
        if (text.match(/data_management:\s*$/)) {
            const tdmItems = [
                { name: 'environment', detail: 'Environment name', documentation: 'The environment to use for this test case.' },
                { name: 'data_sets', detail: 'Data sets array', documentation: 'Array of data sets with test data.' },
                { name: 'validations', detail: 'Validations array', documentation: 'Array of validation rules for data.' }
            ];

            tdmItems.forEach(field => {
                const item = new vscode.CompletionItem(field.name, vscode.CompletionItemKind.Field);
                item.detail = field.detail;
                item.documentation = new vscode.MarkdownString(field.documentation);
                item.insertText = `${field.name}: `;
                items.push(item);
            });
        }

        // Check if we're completing variable fields
        if (text.match(/variables:\s*$/)) {
            const varItems = [
                { name: 'vars', detail: 'Variables object', documentation: 'Define variables as key-value pairs.' },
                { name: 'secrets', detail: 'Secrets object', documentation: 'Define secrets loaded from files or environment.' }
            ];

            varItems.forEach(field => {
                const item = new vscode.CompletionItem(field.name, vscode.CompletionItemKind.Field);
                item.detail = field.detail;
                item.documentation = new vscode.MarkdownString(field.documentation);
                item.insertText = `${field.name}: `;
                items.push(item);
            });
        }

        // Check if we're completing HTTP methods
        if (text.match(/args:\s*\["GET"|"POST"|"PUT"|"DELETE"|"PATCH"|"HEAD"|"OPTIONS"\]/)) {
            const httpMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'];
            httpMethods.forEach(method => {
                const item = new vscode.CompletionItem(method, vscode.CompletionItemKind.Value);
                item.detail = `HTTP ${method} method`;
                item.insertText = method;
                items.push(item);
            });
        }

        // Check if we're completing comparison operators
        if (text.match(/args:\s*\[.*,\s*$/)) {
            const operators = [
                { name: '==', detail: 'Equal to' },
                { name: '!=', detail: 'Not equal to' },
                { name: '>', detail: 'Greater than' },
                { name: '<', detail: 'Less than' },
                { name: '>=', detail: 'Greater than or equal to' },
                { name: '<=', detail: 'Less than or equal to' },
                { name: 'contains', detail: 'Contains substring' },
                { name: 'starts_with', detail: 'Starts with substring' },
                { name: 'ends_with', detail: 'Ends with substring' }
            ];

            operators.forEach(op => {
                const item = new vscode.CompletionItem(op.name, vscode.CompletionItemKind.Operator);
                item.detail = op.detail;
                item.insertText = op.name;
                items.push(item);
            });
        }

        // Check if we're completing time formats
        if (text.match(/args:\s*\["get_time"\]/)) {
            const timeFormats = [
                { name: 'iso', detail: 'ISO 8601 format' },
                { name: 'datetime', detail: 'Date and time format' },
                { name: 'date', detail: 'Date only format' },
                { name: 'time', detail: 'Time only format' },
                { name: 'unix', detail: 'Unix timestamp (seconds)' },
                { name: 'unix_ms', detail: 'Unix timestamp (milliseconds)' },
                { name: 'timestamp', detail: 'Timestamp format' }
            ];

            timeFormats.forEach(format => {
                const item = new vscode.CompletionItem(format.name, vscode.CompletionItemKind.Value);
                item.detail = format.detail;
                item.insertText = format.name;
                items.push(item);
            });
        }

        return items;
    }

    public async getActions(): Promise<RobogoAction[]> {
        return [
            {
                Name: 'log',
                Description: 'Output messages to the console with optional formatting and verbosity control.',
                Example: '- action: log\n  args: ["Hello, World!"]',
                Parameters: [
                    { name: 'message', type: 'string|number|boolean|object', description: 'Message to log', required: true },
                    { name: '...args', type: 'any', description: 'Additional arguments to log', required: false }
                ],
                Returns: 'The logged message as string',
                SupportedTypes: ['Strings', 'Numbers', 'Booleans', 'Objects', 'Arrays'],
                UseCases: ['Debug information', 'Test progress tracking', 'Error reporting', 'Data inspection', 'Performance logging'],
                Notes: 'Messages are displayed in console output. Supports variable substitution with ${variable} syntax. Objects are pretty-printed as JSON. Use verbose field to control output detail level.'
            },
            {
                Name: 'sleep',
                Description: 'Pause test execution for a specified duration.',
                Example: '- action: sleep\n  args: [2]',
                Parameters: [
                    { name: 'duration', type: 'integer|float|string', description: 'Sleep duration in various formats', required: true }
                ],
                Returns: 'Confirmation message with actual sleep duration',
                Formats: ['Integer seconds: 5', 'Float seconds: 0.5', 'String duration: "2m30s"'],
                UseCases: ['Rate limiting for API calls', 'Waiting for async operations', 'Simulating user delays', 'Polling with intervals', 'Performance testing delays'],
                Notes: 'Blocks test execution for the specified duration. Supports sub-second precision for precise timing. String format follows Go\'s time.ParseDuration. Consider impact on test execution time.'
            },
            {
                Name: 'assert',
                Description: 'Validate conditions using various comparison operators and return detailed results.',
                Example: '- action: assert\n  args: ["value", ">", "0", "Value should be positive"]',
                Parameters: [
                    { name: 'actual', type: 'any', description: 'Actual value to compare', required: true },
                    { name: 'operator', type: 'string', description: 'Comparison operator', required: true },
                    { name: 'expected', type: 'any', description: 'Expected value to compare against', required: true },
                    { name: 'message', type: 'string', description: 'Optional custom error message', required: false }
                ],
                Returns: 'Success message or detailed error with actual vs expected values',
                Operators: ['==', '!=', '>', '<', '>=', '<=', 'contains', 'starts_with', 'ends_with'],
                UseCases: ['Response validation', 'Data verification', 'Condition checking', 'Error detection'],
                Notes: 'Supports automatic type conversion for numeric comparisons. String operations are case-sensitive. Boolean values can be strings ("true"/"false") or actual booleans. Use continue_on_failure to prevent test termination on assertion failure.'
            },
            {
                Name: 'get_time',
                Description: 'Get current timestamp with optional format (iso, datetime, date, time, unix, unix_ms, custom formats).',
                Example: '- action: get_time\n  args: ["iso"]\n  result: timestamp',
                Parameters: [
                    { name: 'format', type: 'string', description: 'Time format specification', required: false, default: 'iso' }
                ],
                Returns: 'Formatted timestamp as string',
                Formats: ['iso', 'datetime', 'date', 'time', 'unix', 'unix_ms', 'timestamp', 'custom Go format'],
                UseCases: ['Timestamp generation', 'Date/time formatting', 'API request timing', 'Log timestamps'],
                Notes: 'All timestamps are in UTC unless format specifies timezone. Unix timestamps are in seconds (unix) or milliseconds (unix_ms). Custom formats use Go\'s time formatting reference date: 2006-01-02 15:04:05. Use result field to store timestamp for later use.'
            },
            {
                Name: 'get_random',
                Description: 'Generate random numbers (integers and decimals with precision control).',
                Example: '- action: get_random\n  args: [10, 50]\n  result: random_range',
                Parameters: [
                    { name: 'max', type: 'number', description: 'Maximum value (for single argument: generates 0 to max)', required: true },
                    { name: 'min', type: 'number', description: 'Minimum value (for two arguments: generates min to max)', required: false },
                    { name: 'precision', type: 'number', description: 'Decimal precision (for decimal ranges)', required: false }
                ],
                Returns: 'Random number as string',
                UseCases: ['Test data generation', 'Random ID creation', 'Load testing', 'Simulation scenarios'],
                Notes: 'Uses cryptographically secure random number generation. Supports both integer and decimal ranges. Decimal precision defaults to 2 places if not specified. Inclusive ranges (min and max are possible values). Backward compatible with single argument format.'
            },
            {
                Name: 'concat',
                Description: 'Concatenate multiple strings or values into a single string.',
                Example: '- action: concat\n  args: ["Hello", " ", "World"]\n  result: greeting',
                Parameters: [
                    { name: '...args', type: 'any', description: 'Variable number of values to concatenate', required: true }
                ],
                Returns: 'Concatenated string',
                SupportedTypes: ['Strings', 'Numbers', 'Booleans', 'Arrays', 'Objects'],
                UseCases: ['Building dynamic messages', 'Creating file paths', 'Constructing API endpoints', 'Formatting log messages', 'Building SWIFT messages'],
                Notes: 'All arguments are converted to strings. Arrays are space-separated. Objects are JSON-encoded. Use result field to store concatenated string. Supports variable substitution with ${variable} syntax.'
            },
            {
                Name: 'length',
                Description: 'Get length of strings or arrays.',
                Example: '- action: length\n  args: ["Hello World"]\n  result: str_length',
                Parameters: [
                    { name: 'value', type: 'any', description: 'Value to measure length of', required: true }
                ],
                Returns: 'Length as string representation of number',
                SupportedTypes: ['Strings', 'Arrays', 'Maps', 'Numbers'],
                UseCases: ['Validate string lengths', 'Check array sizes', 'Count items in collections', 'Data validation'],
                Notes: 'Numbers are converted to string before counting digits. Arrays and maps return element count. Returns string representation of number. Use result field to store length for later use.'
            },
            {
                Name: 'http',
                Description: 'Generic HTTP requests with mTLS support and custom options.',
                Example: '- action: http\n  args: ["GET", "https://api.example.com/data", {"Authorization": "Bearer ..."}]\n  result: response',
                Parameters: [
                    { name: 'method', type: 'string', description: 'HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)', required: true },
                    { name: 'url', type: 'string', description: 'Target URL', required: true },
                    { name: 'body', type: 'string', description: 'Request body (optional for GET/HEAD)', required: false },
                    { name: 'headers', type: 'object', description: 'Custom headers map', required: false },
                    { name: 'options', type: 'object', description: 'TLS options (cert, key, ca)', required: false }
                ],
                Returns: 'JSON response with status_code, headers, body, and duration',
                Examples: [
                    '- action: http\n  args: ["GET", "https://api.example.com/data"]',
                    '- action: http\n  args: ["POST", "https://api.example.com/users", \'{"name": "John"}\', {"Content-Type": "application/json"}]',
                    '- action: http\n  args: ["GET", "https://secure.example.com/api", "", "", {"cert": "client.crt", "key": "client.key"}]'
                ],
                UseCases: ['API testing', 'Secure communication', 'Custom headers', 'Certificate-based auth'],
                Notes: 'Automatically sets Content-Type to application/json for POST/PUT/PATCH with body. Supports both file paths and PEM content for certificates. Default timeout is 30 seconds. Response includes timing information.'
            },
            {
                Name: 'http_get',
                Description: 'Simplified GET requests.',
                Example: '- action: http_get\n  args: ["https://api.example.com/data"]\n  result: response',
                Parameters: [
                    { name: 'url', type: 'string', description: 'Target URL to GET', required: true },
                    { name: 'headers', type: 'object', description: 'Optional custom headers map', required: false }
                ],
                Returns: 'JSON response with status_code, headers, body, and duration',
                Examples: [
                    '- action: http_get\n  args: ["https://api.example.com/data"]',
                    '- action: http_get\n  args: ["https://api.example.com/data", {"Authorization": "Bearer token"}]'
                ],
                Notes: 'Simplified wrapper around HTTPAction for GET requests. Automatically sets method to GET. No request body support (GET requests should not have body).'
            },
            {
                Name: 'http_post',
                Description: 'Simplified POST requests.',
                Example: '- action: http_post\n  args: ["https://api.example.com/data", \'{"key": "value"}\']\n  result: response',
                Parameters: [
                    { name: 'url', type: 'string', description: 'Target URL to POST to', required: true },
                    { name: 'body', type: 'string', description: 'Request body (JSON string or data)', required: true },
                    { name: 'headers', type: 'object', description: 'Optional custom headers map', required: false }
                ],
                Returns: 'JSON response with status_code, headers, body, and duration',
                Examples: [
                    '- action: http_post\n  args: ["https://api.example.com/users", \'{"name": "John", "email": "john@example.com"}\']',
                    '- action: http_post\n  args: ["https://api.example.com/users", \'{"name": "John"}\', {"Content-Type": "application/json", "Authorization": "Bearer token"}]'
                ],
                Notes: 'Simplified wrapper around HTTPAction for POST requests. Automatically sets method to POST. Automatically sets Content-Type to application/json if not specified. Body is required for POST requests.'
            },
            {
                Name: 'postgres',
                Description: 'PostgreSQL operations (query, execute, connect, close).',
                Example: '- action: postgres\n  args: ["query", "postgres://user:pass@localhost/db", "SELECT * FROM users"]\n  result: query_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Database operation (query, execute, connect, close)', required: true },
                    { name: 'connection', type: 'string', description: 'Database connection string or connection object', required: true },
                    { name: 'sql', type: 'string', description: 'SQL query or statement', required: false },
                    { name: 'params', type: 'array', description: 'Query parameters', required: false }
                ],
                Returns: 'Query results or operation status',
                UseCases: ['Database testing', 'Data validation', 'Setup/teardown operations', 'Data verification'],
                Notes: 'Supports connection pooling and parameterized queries. Connection strings should be URL-encoded. Use connect/close for connection management. Supports transaction handling.'
            },
            {
                Name: 'variable',
                Description: 'Variable management operations (set, get, list).',
                Example: '- action: variable\n  args: ["set", "my_var", "my_value"]\n  result: set_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Variable operation (set, get, list)', required: true },
                    { name: 'var_name', type: 'string', description: 'Variable name (for set/get operations)', required: false },
                    { name: 'value', type: 'any', description: 'Variable value (for set operation)', required: false }
                ],
                Returns: 'JSON result with operation status and data',
                Examples: [
                    '- action: variable\n  args: ["set", "api_key", "abc123"]',
                    '- action: variable\n  args: ["set", "user_data", {"name": "John", "age": 30}]',
                    '- action: variable\n  args: ["get", "user_id"]',
                    '- action: variable\n  args: ["list"]'
                ],
                UseCases: ['Dynamic data handling', 'State management', 'Configuration storage', 'Data persistence'],
                Notes: 'Variables persist throughout test execution. Supports complex data types (strings, numbers, objects, arrays). Use ${variable_name} syntax to reference in other actions. Variables are shared across all steps in a test case.'
            },
            {
                Name: 'tdm',
                Description: 'Test Data Management operations (generate, validate, load_dataset, set_environment).',
                Example: '- action: tdm\n  args: ["generate", "user_{index}", 5]\n  result: generated_data',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'TDM operation (generate, validate, load_dataset, set_environment)', required: true },
                    { name: '...args', type: 'any', description: 'Operation-specific arguments', required: false }
                ],
                Returns: 'JSON result with operation status and data',
                Examples: [
                    '- action: tdm\n  args: ["generate", "user_{index}", 5]',
                    '- action: tdm\n  args: ["generate", "user_{index}@example.com", 3]',
                    '- action: tdm\n  args: ["validate", "email_validation"]',
                    '- action: tdm\n  args: ["load_dataset", "test_users"]',
                    '- action: tdm\n  args: ["set_environment", "production"]'
                ],
                UseCases: ['Test data generation', 'Data validation', 'Environment management', 'Dataset loading'],
                Notes: 'Generate creates multiple records based on count. Validate returns detailed validation results. Load_dataset makes data available as variables. Set_environment applies environment-specific overrides.'
            },
            {
                Name: 'control',
                Description: 'Control flow operations (if, for, while).',
                Example: '- action: control\n  args: ["if", "${var} > 5"]\n  result: condition_result',
                Parameters: [
                    { name: 'flowType', type: 'string', description: 'Control flow type (if, for, while)', required: true },
                    { name: 'condition', type: 'string', description: 'Condition to evaluate or loop specification', required: true }
                ],
                Returns: 'Control flow result',
                UseCases: ['Conditional execution', 'Loop processing', 'Flow control'],
                Notes: 'Supports if statements, for loops, and while loops with dependency analysis'
            },
            {
                Name: 'http_batch',
                Description: 'Perform multiple HTTP requests in parallel with concurrency control.',
                Example: '- action: http_batch\n  args: [{"requests": [{"method": "GET", "url": "https://api1.com"}, {"method": "GET", "url": "https://api2.com"}], "concurrency": 3}]\n  result: batch_response',
                Parameters: [
                    { name: 'requests', type: 'object[]', description: 'Array of request configurations', required: true },
                    { name: 'concurrency', type: 'number', description: 'Maximum concurrent requests', required: false, default: '4' },
                    { name: 'timeout', type: 'string', description: 'Request timeout', required: false }
                ],
                Returns: 'JSON object with results for each request',
                Examples: [
                    '- action: http_batch\n  args: [{"requests": [{"method": "GET", "url": "https://api1.com"}, {"method": "GET", "url": "https://api2.com"}], "concurrency": 3}]',
                    '- action: http_batch\n  args: [{"requests": [{"method": "POST", "url": "https://api.com/users", "body": \'{"name": "John"}\'}], "timeout": "30s"}]'
                ],
                UseCases: ['Parallel API testing', 'Load testing', 'Batch operations', 'Performance testing'],
                Notes: 'Executes multiple HTTP requests in parallel with configurable concurrency. Each request can have its own method, URL, body, and headers.'
            },
            {
                Name: 'postgres_batch',
                Description: 'Parallel PostgreSQL database operations with concurrency control.',
                Example: '- action: postgres\n  args: ["batch", "postgres://user:pass@localhost/db", [{"query": "SELECT * FROM users"}, {"query": "SELECT * FROM orders"}], {"concurrency": 5}]\n  result: batch_results',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Must be "batch"', required: true },
                    { name: 'connection', type: 'string', description: 'Database connection string', required: true },
                    { name: 'queries', type: 'object[]', description: 'Array of query configurations', required: true },
                    { name: 'options', type: 'object', description: 'Batch options (concurrency, timeout)', required: false }
                ],
                Returns: 'JSON object with results for each query',
                Examples: [
                    '- action: postgres\n  args: ["batch", "postgres://user:pass@localhost/db", [{"query": "SELECT * FROM users"}, {"query": "SELECT * FROM orders"}], {"concurrency": 5}]',
                    '- action: postgres\n  args: ["batch", "postgres://user:pass@localhost/db", [{"execute": "INSERT INTO logs (message) VALUES ($1)", "params": ["test"]}]]'
                ],
                UseCases: ['Parallel database operations', 'Batch data processing', 'Performance testing', 'Data migration'],
                Notes: 'Executes multiple database operations in parallel. Supports both queries and execute operations. Uses connection pooling for efficiency.'
            },
            {
                Name: 'control',
                Description: 'Control flow operations (if, for, while).',
                Example: '- action: control\n  args: ["if", "${var} > 5"]\n  result: condition_result',
                Parameters: [
                    { name: 'flowType', type: 'string', description: 'Control flow type (if, for, while)', required: true },
                    { name: 'condition', type: 'string', description: 'Condition to evaluate or loop specification', required: true }
                ],
                Returns: 'Result of condition evaluation or loop information',
                Examples: [
                    '- action: control\n  args: ["if", "${value} > 5"]',
                    '- action: control\n  args: ["if", "${response} contains \'success\'"]',
                    '- action: control\n  args: ["for", "1..5"]',
                    '- action: control\n  args: ["for", "[item1, item2, item3]"]',
                    '- action: control\n  args: ["while", "${counter} < 10"]'
                ],
                Operators: ['==', '!=', '>', '<', '>=', '<=', 'contains', 'starts_with', 'ends_with', '&&', '||', '!'],
                UseCases: ['Conditional execution', 'Loop control', 'Flow management', 'Dynamic testing'],
                Notes: 'If conditions return boolean for use in if/else blocks. For loops support range, array, and count formats. While conditions return boolean for loop continuation. Use max_iterations to prevent infinite loops.'
            },
            {
                Name: 'kafka',
                Description: 'Apache Kafka operations for message publishing and consuming.',
                Example: '- action: kafka\n  args: ["publish", "localhost:9092", "test-topic", \'{"key": "value"}\']\n  result: publish_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Kafka operation (publish, consume)', required: true },
                    { name: 'brokers', type: 'string', description: 'Comma-separated list of Kafka brokers', required: true },
                    { name: 'topic', type: 'string', description: 'Kafka topic name', required: true },
                    { name: 'message', type: 'string', description: 'Message to publish (for publish operation)', required: false },
                    { name: 'options', type: 'object', description: 'Kafka configuration options', required: false }
                ],
                Returns: 'Operation result with status and data',
                Examples: [
                    '- action: kafka\n  args: ["publish", "localhost:9092", "test-topic", \'{"key": "value"}\']',
                    '- action: kafka\n  args: ["consume", "localhost:9092", "test-topic", "", {"timeout": "30s"}]'
                ],
                UseCases: ['Message queue testing', 'Event-driven testing', 'Integration testing', 'Performance testing'],
                Notes: 'Supports both publishing and consuming messages. Configure timeout for consume operations. Messages are JSON-encoded strings.'
            },
            {
                Name: 'rabbitmq',
                Description: 'RabbitMQ operations for message publishing and consuming.',
                Example: '- action: rabbitmq\n  args: ["publish", "amqp://localhost", "test-exchange", "test-routing-key", \'{"key": "value"}\']\n  result: publish_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'RabbitMQ operation (publish, consume)', required: true },
                    { name: 'url', type: 'string', description: 'RabbitMQ connection URL', required: true },
                    { name: 'exchange', type: 'string', description: 'Exchange name', required: true },
                    { name: 'routing_key', type: 'string', description: 'Routing key', required: true },
                    { name: 'message', type: 'string', description: 'Message to publish (for publish operation)', required: false },
                    { name: 'options', type: 'object', description: 'RabbitMQ configuration options', required: false }
                ],
                Returns: 'Operation result with status and data',
                Examples: [
                    '- action: rabbitmq\n  args: ["publish", "amqp://localhost", "test-exchange", "test-routing-key", \'{"key": "value"}\']',
                    '- action: rabbitmq\n  args: ["consume", "amqp://localhost", "test-exchange", "test-routing-key", "", {"timeout": "30s"}]'
                ],
                UseCases: ['Message queue testing', 'Event-driven testing', 'Integration testing', 'Performance testing'],
                Notes: 'Supports both publishing and consuming messages. Configure timeout for consume operations. Messages are JSON-encoded strings.'
            },
            {
                Name: 'spanner',
                Description: 'Google Cloud Spanner database operations.',
                Example: '- action: spanner\n  args: ["query", "projects/my-project/instances/my-instance/databases/my-db", "SELECT * FROM users"]\n  result: query_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Spanner operation (query, execute, read)', required: true },
                    { name: 'database', type: 'string', description: 'Spanner database path', required: true },
                    { name: 'sql', type: 'string', description: 'SQL query or statement', required: false },
                    { name: 'params', type: 'array', description: 'Query parameters', required: false },
                    { name: 'options', type: 'object', description: 'Spanner configuration options', required: false }
                ],
                Returns: 'Query results or operation status',
                Examples: [
                    '- action: spanner\n  args: ["query", "projects/my-project/instances/my-instance/databases/my-db", "SELECT * FROM users"]',
                    '- action: spanner\n  args: ["execute", "projects/my-project/instances/my-instance/databases/my-db", "INSERT INTO users (id, name) VALUES (@id, @name)", [{"id": 1, "name": "John"}]]'
                ],
                UseCases: ['Cloud database testing', 'Data validation', 'Setup/teardown operations', 'Data verification'],
                Notes: 'Requires Google Cloud authentication. Supports parameterized queries with named parameters. Use read operation for large result sets.'
            },
            {
                Name: 'template',
                Description: 'Generate content from templates with variable substitution.',
                Example: '- action: template\n  args: ["templates/mt103.tmpl", {"amount": "1000.00", "currency": "EUR"}]\n  result: swift_message',
                Parameters: [
                    { name: 'template_path', type: 'string', description: 'Path to template file', required: true },
                    { name: 'variables', type: 'object', description: 'Variables to substitute in template', required: true }
                ],
                Returns: 'Generated content as string',
                Examples: [
                    '- action: template\n  args: ["templates/mt103.tmpl", {"amount": "1000.00", "currency": "EUR"}]',
                    '- action: template\n  args: ["templates/sepa-credit-transfer.xml.tmpl", {"debtor": "John Doe", "creditor": "Jane Smith", "amount": "500.00"}]'
                ],
                UseCases: ['SWIFT message generation', 'XML document creation', 'Email template generation', 'Report generation'],
                Notes: 'Uses Go template syntax. Variables are substituted using {{.variable_name}} syntax. Supports nested objects and arrays.'
            },
            {
                Name: 'skip',
                Description: 'Skip the current step or test case based on conditions.',
                Example: '- name: "Skip step"\n  action: skip\n  args: ["This step is not needed in current environment"]',
                Parameters: [
                    { name: 'reason', type: 'string', description: 'Reason for skipping', required: true }
                ],
                Returns: 'Skip confirmation message',
                Examples: [
                    '- action: skip\n  args: ["This step is not needed in current environment"]',
                    '- action: skip\n  args: ["Feature not implemented yet"]'
                ],
                UseCases: ['Conditional test execution', 'Feature toggles', 'Environment-specific tests', 'Temporary test disabling'],
                Notes: 'Immediately stops execution of current step or test case. Use skip field in step definition for conditional skipping.'
            },
            {
                Name: 'string',
                Description: 'String manipulation operations.',
                Example: '- action: string\n  args: ["upper", "hello world"]\n  result: uppercase_string',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'String operation (upper, lower, trim, replace, split, join)', required: true },
                    { name: 'input', type: 'string', description: 'Input string', required: true },
                    { name: '...args', type: 'any', description: 'Additional operation-specific arguments', required: false }
                ],
                Returns: 'Manipulated string',
                Examples: [
                    '- action: string\n  args: ["upper", "hello world"]',
                    '- action: string\n  args: ["replace", "hello world", "world", "robogo"]',
                    '- action: string\n  args: ["split", "a,b,c", ","]'
                ],
                UseCases: ['Text processing', 'Data cleaning', 'Format conversion', 'String validation'],
                Notes: 'Supports common string operations like case conversion, trimming, replacement, splitting, and joining.'
            },
            {
                Name: 'time',
                Description: 'Time manipulation and formatting operations.',
                Example: '- action: time\n  args: ["add", "2024-01-01", "24h"]\n  result: future_date',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Time operation (add, sub, format, parse)', required: true },
                    { name: 'time', type: 'string', description: 'Time value or timestamp', required: true },
                    { name: 'duration', type: 'string', description: 'Duration to add/subtract (for add/sub operations)', required: false },
                    { name: 'format', type: 'string', description: 'Output format (for format operation)', required: false }
                ],
                Returns: 'Manipulated or formatted time',
                Examples: [
                    '- action: time\n  args: ["add", "2024-01-01", "24h"]',
                    '- action: time\n  args: ["format", "2024-01-01T12:00:00Z", "2006-01-02"]',
                    '- action: time\n  args: ["parse", "01/02/2024", "01/02/2006"]'
                ],
                UseCases: ['Date/time calculations', 'Time formatting', 'Schedule testing', 'Time-based validation'],
                Notes: 'Supports adding/subtracting durations, formatting timestamps, and parsing time strings. Uses Go time package for operations.'
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
        if (actionMatch && actionMatch[1]) {
            const actionName = actionMatch[1];
            const actions = await this.getActions();
            const action = actions.find(a => a.Name === actionName);

            if (action) {
                const markdown = new vscode.MarkdownString();
                markdown.appendMarkdown(`**${action.Name}**\n\n`);
                markdown.appendMarkdown(`${action.Description}\n\n`);

                // Add parameters if available
                if (action.Parameters && action.Parameters.length > 0) {
                    markdown.appendMarkdown(`**Parameters:**\n`);
                    action.Parameters.forEach(param => {
                        const required = param.required ? 'required' : 'optional';
                        const defaultValue = param.default ? ` (default: ${param.default})` : '';
                        markdown.appendMarkdown(`- \`${param.name}\` (${param.type}, ${required})${defaultValue}: ${param.description}\n`);
                    });
                    markdown.appendMarkdown(`\n`);
                }

                // Add return value if available
                if (action.Returns) {
                    markdown.appendMarkdown(`**Returns:** ${action.Returns}\n\n`);
                }

                // Add supported types if available
                if (action.SupportedTypes && action.SupportedTypes.length > 0) {
                    markdown.appendMarkdown(`**Supported Types:**\n`);
                    action.SupportedTypes.forEach(type => {
                        markdown.appendMarkdown(`- ${type}\n`);
                    });
                    markdown.appendMarkdown(`\n`);
                }

                // Add operators if available
                if (action.Operators && action.Operators.length > 0) {
                    markdown.appendMarkdown(`**Operators:**\n`);
                    action.Operators.forEach(op => {
                        markdown.appendMarkdown(`- \`${op}\`\n`);
                    });
                    markdown.appendMarkdown(`\n`);
                }

                // Add formats if available
                if (action.Formats && action.Formats.length > 0) {
                    markdown.appendMarkdown(`**Formats:**\n`);
                    action.Formats.forEach(format => {
                        markdown.appendMarkdown(`- \`${format}\`\n`);
                    });
                    markdown.appendMarkdown(`\n`);
                }

                // Add examples
                if (action.Examples && action.Examples.length > 0) {
                    markdown.appendMarkdown(`**Examples:**\n`);
                    action.Examples.forEach(example => {
                        markdown.appendMarkdown(`\`\`\`yaml\n${example}\n\`\`\`\n`);
                    });
                    markdown.appendMarkdown(`\n`);
                } else {
                    markdown.appendMarkdown(`**Example:**\n\`\`\`yaml\n${action.Example}\n\`\`\`\n\n`);
                }

                // Add use cases if available
                if (action.UseCases && action.UseCases.length > 0) {
                    markdown.appendMarkdown(`**Use Cases:**\n`);
                    action.UseCases.forEach(useCase => {
                        markdown.appendMarkdown(`- ${useCase}\n`);
                    });
                    markdown.appendMarkdown(`\n`);
                }

                // Add notes if available
                if (action.Notes) {
                    markdown.appendMarkdown(`**Notes:**\n${action.Notes}\n\n`);
                }

                return new vscode.Hover(markdown);
            }
        }

        // Check if we're hovering over field names (TDM, environment, etc.)
        const fieldMatch = text.match(/^(\s*)(\w+):/);
        if (fieldMatch && fieldMatch[2]) {
            const fieldName = fieldMatch[2];
            const fieldDoc = this.getFieldDocumentation(fieldName);
            if (fieldDoc) {
                const markdown = new vscode.MarkdownString();
                markdown.appendMarkdown(`**${fieldDoc.name}**\n\n`);
                markdown.appendMarkdown(`${fieldDoc.description}\n\n`);
                if (fieldDoc.type) {
                    markdown.appendMarkdown(`**Type:** ${fieldDoc.type}\n\n`);
                }
                if (fieldDoc.required !== undefined) {
                    markdown.appendMarkdown(`**Required:** ${fieldDoc.required ? 'Yes' : 'No'}\n\n`);
                }
                if (fieldDoc.example) {
                    markdown.appendMarkdown(`**Example:**\n\`\`\`yaml\n${fieldDoc.example}\n\`\`\`\n`);
                }
                return new vscode.Hover(markdown);
            }
        }

        return undefined;
    }

    private getFieldDocumentation(fieldName: string): any {
        const fieldDocs: { [key: string]: any } = {
            'testcase': {
                name: 'Test Case Name',
                description: 'The name of the test case. This is a required field that identifies the test.',
                type: 'string',
                required: true,
                example: 'testcase: "User Login Test"'
            },
            'description': {
                name: 'Description',
                description: 'Optional description of what the test case does.',
                type: 'string',
                required: false,
                example: 'description: "Test user login functionality"'
            },
            'steps': {
                name: 'Test Steps',
                description: 'Array of test steps to execute. Each step can be an action or control flow.',
                type: 'array',
                required: true,
                example: 'steps:\n  - name: "Login"\n    action: http_post\n    args: ["https://api.example.com/login"]'
            },
            'variables': {
                name: 'Variables',
                description: 'Define variables and secrets for use throughout the test case.',
                type: 'object',
                required: false,
                example: 'variables:\n  vars:\n    base_url: "https://api.example.com"\n  secrets:\n    api_key:\n      file: "secret.txt"'
            },
            'data_management': {
                name: 'Data Management',
                description: 'Test Data Management configuration including data sets, validation, and environments.',
                type: 'object',
                required: false,
                example: 'data_management:\n  environment: "development"\n  data_sets:\n    - name: "test_users"\n      data:\n        user1:\n          name: "John Doe"'
            },
            'environments': {
                name: 'Environments',
                description: 'Environment-specific configurations with variables and overrides.',
                type: 'array',
                required: false,
                example: 'environments:\n  - name: "development"\n    variables:\n      api_base_url: "https://dev-api.example.com"'
            },
            'name': {
                name: 'Name',
                description: 'Name identifier for the current element (step, data set, environment, etc.).',
                type: 'string',
                required: false,
                example: 'name: "Login Step"'
            },
            'action': {
                name: 'Action',
                description: 'The action to execute (log, http, assert, etc.).',
                type: 'string',
                required: true,
                example: 'action: http_post'
            },
            'args': {
                name: 'Arguments',
                description: 'Arguments to pass to the action.',
                type: 'array',
                required: false,
                example: 'args: ["https://api.example.com/data", {"key": "value"}]'
            },
            'result': {
                name: 'Result Variable',
                description: 'Variable name to store the action result.',
                type: 'string',
                required: false,
                example: 'result: response'
            },
            'if': {
                name: 'If Statement',
                description: 'Conditional execution block with then/else branches.',
                type: 'object',
                required: false,
                example: 'if:\n  condition: "${value} > 5"\n  then:\n    - action: log\n      args: ["Value is greater than 5"]'
            },
            'for': {
                name: 'For Loop',
                description: 'Loop execution block with range, array, or count.',
                type: 'object',
                required: false,
                example: 'for:\n  condition: "1..5"\n  steps:\n    - action: log\n      args: ["Iteration ${iteration}"]'
            },
            'while': {
                name: 'While Loop',
                description: 'Conditional loop with condition and max iterations.',
                type: 'object',
                required: false,
                example: 'while:\n  condition: "${counter} < 10"\n  max_iterations: 20\n  steps:\n    - action: log\n      args: ["Counter: ${counter}"]'
            },
            'continue_on_failure': {
                name: 'Continue on Failure',
                description: 'Continue test execution even if this step fails.',
                type: 'boolean',
                required: false,
                example: 'continue_on_failure: true'
            },
            'verbose': {
                name: 'Verbose Output',
                description: 'Enable detailed output for this step.',
                type: 'boolean|string',
                required: false,
                example: 'verbose: true'
            },
            'retry': {
                name: 'Retry Configuration',
                description: 'Retry settings for this step.',
                type: 'object',
                required: false,
                example: 'retry:\n  attempts: 3\n  delay: 1s\n  backoff: exponential'
            },
            'testcases': {
                name: 'Test Cases',
                description: 'Array of test case files to include in the test suite.',
                type: 'array',
                required: false,
                example: 'testcases:\n  - file: test1.robogo\n  - file: test2.robogo'
            },
            'testsuite': {
                name: 'Test Suite',
                description: 'The name of the test suite. This is a required field for test suite files.',
                type: 'string',
                required: true,
                example: 'testsuite: "API Test Suite"'
            },
            'parallel': {
                name: 'Parallel Execution',
                description: 'Enable or disable parallel execution for the test suite.',
                type: 'boolean',
                required: false,
                example: 'parallel: false'
            },
            'options': {
                name: 'Options',
                description: 'Test suite options including max_concurrency and other settings.',
                type: 'object',
                required: false,
                example: 'options:\n  max_concurrency: 3'
            },
            'regular': {
                name: 'Regular Variables',
                description: 'Regular variables section in test suite variables.',
                type: 'object',
                required: false,
                example: 'regular:\n  api_url: "https://api.example.com"'
            }
        };

        return fieldDocs[fieldName];
    }

    private async getActions(): Promise<RobogoAction[]> {
        return [
            {
                Name: 'log',
                Description: 'Output messages to the console with optional formatting and verbosity control.',
                Example: '- action: log\n  args: ["Hello, World!"]',
                Parameters: [
                    { name: 'message', type: 'string|number|boolean|object', description: 'Message to log', required: true },
                    { name: '...args', type: 'any', description: 'Additional arguments to log', required: false }
                ],
                Returns: 'The logged message as string',
                SupportedTypes: ['Strings', 'Numbers', 'Booleans', 'Objects', 'Arrays'],
                UseCases: ['Debug information', 'Test progress tracking', 'Error reporting', 'Data inspection', 'Performance logging'],
                Notes: 'Messages are displayed in console output. Supports variable substitution with ${variable} syntax. Objects are pretty-printed as JSON. Use verbose field to control output detail level.'
            },
            {
                Name: 'sleep',
                Description: 'Pause test execution for a specified duration.',
                Example: '- action: sleep\n  args: [2]',
                Parameters: [
                    { name: 'duration', type: 'integer|float|string', description: 'Sleep duration in various formats', required: true }
                ],
                Returns: 'Confirmation message with actual sleep duration',
                Formats: ['Integer seconds: 5', 'Float seconds: 0.5', 'String duration: "2m30s"'],
                UseCases: ['Rate limiting for API calls', 'Waiting for async operations', 'Simulating user delays', 'Polling with intervals', 'Performance testing delays'],
                Notes: 'Blocks test execution for the specified duration. Supports sub-second precision for precise timing. String format follows Go\'s time.ParseDuration. Consider impact on test execution time.'
            },
            {
                Name: 'assert',
                Description: 'Validate conditions using various comparison operators and return detailed results.',
                Example: '- action: assert\n  args: ["value", ">", "0", "Value should be positive"]',
                Parameters: [
                    { name: 'actual', type: 'any', description: 'Actual value to compare', required: true },
                    { name: 'operator', type: 'string', description: 'Comparison operator', required: true },
                    { name: 'expected', type: 'any', description: 'Expected value to compare against', required: true },
                    { name: 'message', type: 'string', description: 'Optional custom error message', required: false }
                ],
                Returns: 'Success message or detailed error with actual vs expected values',
                Operators: ['==', '!=', '>', '<', '>=', '<=', 'contains', 'starts_with', 'ends_with'],
                UseCases: ['Response validation', 'Data verification', 'Condition checking', 'Error detection'],
                Notes: 'Supports automatic type conversion for numeric comparisons. String operations are case-sensitive. Boolean values can be strings ("true"/"false") or actual booleans. Use continue_on_failure to prevent test termination on assertion failure.'
            },
            {
                Name: 'get_time',
                Description: 'Get current timestamp with optional format (iso, datetime, date, time, unix, unix_ms, custom formats).',
                Example: '- action: get_time\n  args: ["iso"]\n  result: timestamp',
                Parameters: [
                    { name: 'format', type: 'string', description: 'Time format specification', required: false, default: 'iso' }
                ],
                Returns: 'Formatted timestamp as string',
                Formats: ['iso', 'datetime', 'date', 'time', 'unix', 'unix_ms', 'timestamp', 'custom Go format'],
                UseCases: ['Timestamp generation', 'Date/time formatting', 'API request timing', 'Log timestamps'],
                Notes: 'All timestamps are in UTC unless format specifies timezone. Unix timestamps are in seconds (unix) or milliseconds (unix_ms). Custom formats use Go\'s time formatting reference date: 2006-01-02 15:04:05. Use result field to store timestamp for later use.'
            },
            {
                Name: 'get_random',
                Description: 'Generate random numbers (integers and decimals with precision control).',
                Example: '- action: get_random\n  args: [10, 50]\n  result: random_range',
                Parameters: [
                    { name: 'max', type: 'number', description: 'Maximum value (for single argument: generates 0 to max)', required: true },
                    { name: 'min', type: 'number', description: 'Minimum value (for two arguments: generates min to max)', required: false },
                    { name: 'precision', type: 'number', description: 'Decimal precision (for decimal ranges)', required: false }
                ],
                Returns: 'Random number as string',
                UseCases: ['Test data generation', 'Random ID creation', 'Load testing', 'Simulation scenarios'],
                Notes: 'Uses cryptographically secure random number generation. Supports both integer and decimal ranges. Decimal precision defaults to 2 places if not specified. Inclusive ranges (min and max are possible values). Backward compatible with single argument format.'
            },
            {
                Name: 'concat',
                Description: 'Concatenate multiple strings or values into a single string.',
                Example: '- action: concat\n  args: ["Hello", " ", "World"]\n  result: greeting',
                Parameters: [
                    { name: '...args', type: 'any', description: 'Variable number of values to concatenate', required: true }
                ],
                Returns: 'Concatenated string',
                SupportedTypes: ['Strings', 'Numbers', 'Booleans', 'Arrays', 'Objects'],
                UseCases: ['Building dynamic messages', 'Creating file paths', 'Constructing API endpoints', 'Formatting log messages', 'Building SWIFT messages'],
                Notes: 'All arguments are converted to strings. Arrays are space-separated. Objects are JSON-encoded. Use result field to store concatenated string. Supports variable substitution with ${variable} syntax.'
            },
            {
                Name: 'length',
                Description: 'Get length of strings or arrays.',
                Example: '- action: length\n  args: ["Hello World"]\n  result: str_length',
                Parameters: [
                    { name: 'value', type: 'any', description: 'Value to measure length of', required: true }
                ],
                Returns: 'Length as string representation of number',
                SupportedTypes: ['Strings', 'Arrays', 'Maps', 'Numbers'],
                UseCases: ['Validate string lengths', 'Check array sizes', 'Count items in collections', 'Data validation'],
                Notes: 'Numbers are converted to string before counting digits. Arrays and maps return element count. Returns string representation of number. Use result field to store length for later use.'
            },
            {
                Name: 'http',
                Description: 'Generic HTTP requests with mTLS support and custom options.',
                Example: '- action: http\n  args: ["GET", "https://api.example.com/data", {"Authorization": "Bearer ..."}]\n  result: response',
                Parameters: [
                    { name: 'method', type: 'string', description: 'HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)', required: true },
                    { name: 'url', type: 'string', description: 'Target URL', required: true },
                    { name: 'body', type: 'string', description: 'Request body (optional for GET/HEAD)', required: false },
                    { name: 'headers', type: 'object', description: 'Custom headers map', required: false },
                    { name: 'options', type: 'object', description: 'TLS options (cert, key, ca)', required: false }
                ],
                Returns: 'JSON response with status_code, headers, body, and duration',
                Examples: [
                    '- action: http\n  args: ["GET", "https://api.example.com/data"]',
                    '- action: http\n  args: ["POST", "https://api.example.com/users", \'{"name": "John"}\', {"Content-Type": "application/json"}]',
                    '- action: http\n  args: ["GET", "https://secure.example.com/api", "", "", {"cert": "client.crt", "key": "client.key"}]'
                ],
                UseCases: ['API testing', 'Secure communication', 'Custom headers', 'Certificate-based auth'],
                Notes: 'Automatically sets Content-Type to application/json for POST/PUT/PATCH with body. Supports both file paths and PEM content for certificates. Default timeout is 30 seconds. Response includes timing information.'
            },
            {
                Name: 'http_get',
                Description: 'Simplified GET requests.',
                Example: '- action: http_get\n  args: ["https://api.example.com/data"]\n  result: response',
                Parameters: [
                    { name: 'url', type: 'string', description: 'Target URL to GET', required: true },
                    { name: 'headers', type: 'object', description: 'Optional custom headers map', required: false }
                ],
                Returns: 'JSON response with status_code, headers, body, and duration',
                Examples: [
                    '- action: http_get\n  args: ["https://api.example.com/data"]',
                    '- action: http_get\n  args: ["https://api.example.com/data", {"Authorization": "Bearer token"}]'
                ],
                Notes: 'Simplified wrapper around HTTPAction for GET requests. Automatically sets method to GET. No request body support (GET requests should not have body).'
            },
            {
                Name: 'http_post',
                Description: 'Simplified POST requests.',
                Example: '- action: http_post\n  args: ["https://api.example.com/data", \'{"key": "value"}\']\n  result: response',
                Parameters: [
                    { name: 'url', type: 'string', description: 'Target URL to POST to', required: true },
                    { name: 'body', type: 'string', description: 'Request body (JSON string or data)', required: true },
                    { name: 'headers', type: 'object', description: 'Optional custom headers map', required: false }
                ],
                Returns: 'JSON response with status_code, headers, body, and duration',
                Examples: [
                    '- action: http_post\n  args: ["https://api.example.com/users", \'{"name": "John", "email": "john@example.com"}\']',
                    '- action: http_post\n  args: ["https://api.example.com/users", \'{"name": "John"}\', {"Content-Type": "application/json", "Authorization": "Bearer token"}]'
                ],
                Notes: 'Simplified wrapper around HTTPAction for POST requests. Automatically sets method to POST. Automatically sets Content-Type to application/json if not specified. Body is required for POST requests.'
            },
            {
                Name: 'postgres',
                Description: 'PostgreSQL operations (query, execute, connect, close).',
                Example: '- action: postgres\n  args: ["query", "postgres://user:pass@localhost/db", "SELECT * FROM users"]\n  result: query_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Database operation (query, execute, connect, close)', required: true },
                    { name: 'connection', type: 'string', description: 'Database connection string or connection object', required: true },
                    { name: 'sql', type: 'string', description: 'SQL query or statement', required: false },
                    { name: 'params', type: 'array', description: 'Query parameters', required: false }
                ],
                Returns: 'Query results or operation status',
                UseCases: ['Database testing', 'Data validation', 'Setup/teardown operations', 'Data verification'],
                Notes: 'Supports connection pooling and parameterized queries. Connection strings should be URL-encoded. Use connect/close for connection management. Supports transaction handling.'
            },
            {
                Name: 'variable',
                Description: 'Variable management operations (set, get, list).',
                Example: '- action: variable\n  args: ["set", "my_var", "my_value"]\n  result: set_result',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'Variable operation (set, get, list)', required: true },
                    { name: 'var_name', type: 'string', description: 'Variable name (for set/get operations)', required: false },
                    { name: 'value', type: 'any', description: 'Variable value (for set operation)', required: false }
                ],
                Returns: 'JSON result with operation status and data',
                Examples: [
                    '- action: variable\n  args: ["set", "api_key", "abc123"]',
                    '- action: variable\n  args: ["set", "user_data", {"name": "John", "age": 30}]',
                    '- action: variable\n  args: ["get", "user_id"]',
                    '- action: variable\n  args: ["list"]'
                ],
                UseCases: ['Dynamic data handling', 'State management', 'Configuration storage', 'Data persistence'],
                Notes: 'Variables persist throughout test execution. Supports complex data types (strings, numbers, objects, arrays). Use ${variable_name} syntax to reference in other actions. Variables are shared across all steps in a test case.'
            },
            {
                Name: 'tdm',
                Description: 'Test Data Management operations (generate, validate, load_dataset, set_environment).',
                Example: '- action: tdm\n  args: ["generate", "user_{index}", 5]\n  result: generated_data',
                Parameters: [
                    { name: 'operation', type: 'string', description: 'TDM operation (generate, validate, load_dataset, set_environment)', required: true },
                    { name: '...args', type: 'any', description: 'Operation-specific arguments', required: false }
                ],
                Returns: 'JSON result with operation status and data',
                Examples: [
                    '- action: tdm\n  args: ["generate", "user_{index}", 5]',
                    '- action: tdm\n  args: ["generate", "user_{index}@example.com", 3]',
                    '- action: tdm\n  args: ["validate", "email_validation"]',
                    '- action: tdm\n  args: ["load_dataset", "test_users"]',
                    '- action: tdm\n  args: ["set_environment", "production"]'
                ],
                UseCases: ['Test data generation', 'Data validation', 'Environment management', 'Dataset loading'],
                Notes: 'Generate creates multiple records based on count. Validate returns detailed validation results. Load_dataset makes data available as variables. Set_environment applies environment-specific overrides.'
            },
            {
                Name: 'control',
                Description: 'Control flow operations (if, for, while).',
                Example: '- action: control\n  args: ["if", "${var} > 5"]\n  result: condition_result',
                Parameters: [
                    { name: 'flowType', type: 'string', description: 'Control flow type (if, for, while)', required: true },
                    { name: 'condition', type: 'string', description: 'Condition to evaluate or loop specification', required: true }
                ],
                Returns: 'Result of condition evaluation or loop information',
                Examples: [
                    '- action: control\n  args: ["if", "${value} > 5"]',
                    '- action: control\n  args: ["if", "${response} contains \'success\'"]',
                    '- action: control\n  args: ["for", "1..5"]',
                    '- action: control\n  args: ["for", "[item1, item2, item3]"]',
                    '- action: control\n  args: ["while", "${counter} < 10"]'
                ],
                Operators: ['==', '!=', '>', '<', '>=', '<=', 'contains', 'starts_with', 'ends_with', '&&', '||', '!'],
                UseCases: ['Conditional execution', 'Loop control', 'Flow management', 'Dynamic testing'],
                Notes: 'If conditions return boolean for use in if/else blocks. For loops support range, array, and count formats. While conditions return boolean for loop continuation. Use max_iterations to prevent infinite loops.'
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

async function runTestParallel() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const config = vscode.workspace.getConfiguration('robogo');
    const executablePath = config.get<string>('executablePath', 'robogo');
    const maxConcurrency = config.get<number>('maxConcurrency', 4);

    try {
        const terminal = vscode.window.createTerminal('Robogo Parallel Test');
        terminal.sendText(`${executablePath} run "${editor.document.fileName}" --parallel --concurrency ${maxConcurrency}`);
        terminal.show();
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to run test in parallel: ${error}`);
    }
}

async function validateTDM() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const config = vscode.workspace.getConfiguration('robogo');
    const executablePath = config.get<string>('executablePath', 'robogo');

    try {
        // For now, we'll just run the test to validate TDM
        // In the future, this could be a dedicated validation command
        const terminal = vscode.window.createTerminal('Robogo TDM Validation');
        terminal.sendText(`${executablePath} run "${editor.document.fileName}" --output json`);
        terminal.show();
        vscode.window.showInformationMessage('TDM validation started. Check terminal for results.');
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to validate TDM: ${error}`);
    }
}

async function generateTemplate() {
    const template = `testcase: "New Test Case"
description: "Description of the test case"

variables:
  vars:
    base_url: "https://api.example.com"
  secrets:
    api_key:
      file: "secret.txt"
      mask_output: true

data_management:
  environment: "development"
  isolation: true
  cleanup: true
  data_sets:
    - name: "test_data"
      description: "Test data set"
      version: "1.0"
      data:
        user1:
          name: "John Doe"
          email: "john@example.com"
      schema:
        name: "string"
        email: "email"
      required: ["name", "email"]
      unique: ["email"]

steps:
  - name: "Get current timestamp"
    action: get_time
    args: ["iso"]
    result: timestamp
  
  - name: "Log start time"
    action: log
    args: ["Test started at: \${timestamp}"]
  
  - name: "Make API request"
    action: http_get
    args: ["\${base_url}/users"]
    result: response
  
  - name: "Assert response status"
    action: assert
    args: ["\${response.status_code}", "==", "200", "API should return 200"]
  
  - name: "Log completion"
    action: log
    args: ["Test completed successfully"]
`;

    const document = await vscode.workspace.openTextDocument({
        content: template,
        language: 'robogo'
    });
    await vscode.window.showTextDocument(document);
    vscode.window.showInformationMessage('Generated new test template');
}

async function runWithOutput() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const config = vscode.workspace.getConfiguration('robogo');
    const executablePath = config.get<string>('executablePath', 'robogo');

    // Show quick pick for output format
    const outputFormat = await vscode.window.showQuickPick(
        ['console', 'json', 'markdown'],
        {
            placeHolder: 'Select output format'
        }
    );

    if (!outputFormat) {
        return;
    }

    try {
        const terminal = vscode.window.createTerminal('Robogo Test');
        terminal.sendText(`${executablePath} run "${editor.document.fileName}" --output ${outputFormat}`);
        terminal.show();
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to run test with output format: ${error}`);
    }
}

async function runTestSuite() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const document = editor.document;
    if (!document.fileName.endsWith('.robogo')) {
        vscode.window.showErrorMessage('Not a Robogo file');
        return;
    }

    const config = vscode.workspace.getConfiguration('robogo');
    const executablePath = config.get<string>('executablePath', 'robogo');

    try {
        const terminal = vscode.window.createTerminal('Robogo Test Suite');
        terminal.sendText(`${executablePath} run-suite "${document.fileName}"`);
        terminal.show();
        vscode.window.showInformationMessage('Test suite execution started');
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to run test suite: ${error}`);
    }
}

async function showDocumentation() {
    const actions = await new RobogoCompletionProvider().getActions();

    const markdown = new vscode.MarkdownString();
    markdown.appendMarkdown('# Robogo Actions Documentation\n\n');

    // Group actions by category
    const categories: { [key: string]: RobogoAction[] } = {};
    actions.forEach(action => {
        const category = getActionCategory(action.Name);
        if (!categories[category]) {
            categories[category] = [];
        }
        categories[category].push(action);
    });

    Object.keys(categories).sort().forEach(category => {
        markdown.appendMarkdown(`## ${category}\n\n`);
        categories[category]?.forEach(action => {
            markdown.appendMarkdown(`### ${action.Name}\n\n`);
            markdown.appendMarkdown(`${action.Description}\n\n`);
            markdown.appendMarkdown(`**Example:**\n\`\`\`yaml\n${action.Example}\n\`\`\`\n\n`);

            if (action.Parameters && action.Parameters.length > 0) {
                markdown.appendMarkdown(`**Parameters:**\n`);
                action.Parameters.forEach(param => {
                    const required = param.required ? 'required' : 'optional';
                    markdown.appendMarkdown(`- \`${param.name}\` (${param.type}, ${required}): ${param.description}\n`);
                });
                markdown.appendMarkdown(`\n`);
            }

            if (action.Returns) {
                markdown.appendMarkdown(`**Returns:** ${action.Returns}\n\n`);
            }

            if (action.Notes) {
                markdown.appendMarkdown(`**Notes:** ${action.Notes}\n\n`);
            }
        });
    });

    // Show documentation in a new document
    const doc = await vscode.workspace.openTextDocument({
        content: markdown.value,
        language: 'markdown'
    });
    await vscode.window.showTextDocument(doc);
}

async function validateSyntax() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const document = editor.document;
    if (!document.fileName.endsWith('.robogo')) {
        vscode.window.showErrorMessage('Not a Robogo file');
        return;
    }

    try {
        const text = document.getText();
        const parsed = yaml.load(text) as any;

        if (!parsed) {
            vscode.window.showErrorMessage('Invalid YAML syntax');
            return;
        }

        // Basic validation
        const errors: string[] = [];

        if (!parsed.testcase && !parsed.testsuite) {
            errors.push('Missing required field: testcase or testsuite');
        }

        if (parsed.testcase && !parsed.steps) {
            errors.push('Missing required field: steps');
        }

        if (parsed.steps && !Array.isArray(parsed.steps)) {
            errors.push('Steps must be an array');
        }

        if (errors.length > 0) {
            vscode.window.showErrorMessage(`Validation errors:\n${errors.join('\n')}`);
        } else {
            vscode.window.showInformationMessage('Syntax validation passed');
        }
    } catch (error) {
        vscode.window.showErrorMessage(`Syntax error: ${error}`);
    }
}

function getActionCategory(actionName: string): string {
    const categoryMap: { [key: string]: string } = {
        'log': 'Basic',
        'sleep': 'Basic',
        'assert': 'Basic',
        'get_time': 'Basic',
        'get_random': 'Basic',
        'concat': 'Basic',
        'length': 'Basic',
        'http': 'HTTP',
        'http_get': 'HTTP',
        'http_post': 'HTTP',
        'postgres': 'Database',
        'spanner': 'Database',
        'kafka': 'Messaging',
        'rabbitmq': 'Messaging',
        'variable': 'Variables',
        'tdm': 'TDM',
        'control': 'Control Flow',
        'template': 'Templates',
        'skip': 'Control Flow',
        'string': 'Utilities',
        'time': 'Utilities'
    };

    return categoryMap[actionName] || 'Other';
}

export function deactivate() { } 