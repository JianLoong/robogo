import * as vscode from 'vscode';
import { ConfigurationManager } from './configurationManager';
import { RobogoActionRegistry } from './actionRegistry';

/**
 * Core language server for Robogo extension
 * Provides action definitions, validation, and language services
 */
export class RobogoLanguageServer {
    private actionRegistry: RobogoActionRegistry;

    constructor(private config: ConfigurationManager) {
        this.actionRegistry = new RobogoActionRegistry();
    }

    /**
     * Get action registry
     */
    getActionRegistry(): RobogoActionRegistry {
        return this.actionRegistry;
    }

    /**
     * Get action documentation
     */
    getActionDocumentation(actionName: string): string | undefined {
        return this.actionRegistry.getActionDocumentation(actionName);
    }

    /**
     * Get action completions
     */
    getActionCompletions(): vscode.CompletionItem[] {
        return this.actionRegistry.getAllActions().map(action => {
            const item = new vscode.CompletionItem(action.name, vscode.CompletionItemKind.Function);
            item.detail = action.description;
            item.documentation = new vscode.MarkdownString(action.documentation);
            item.insertText = new vscode.SnippetString(action.snippet);
            item.sortText = `0_${action.name}`;
            return item;
        });
    }

    /**
     * Get variable completions from context
     */
    getVariableCompletions(document: vscode.TextDocument): vscode.CompletionItem[] {
        const variables = this.extractVariablesFromDocument(document);
        return variables.map(variable => {
            const item = new vscode.CompletionItem(variable, vscode.CompletionItemKind.Variable);
            item.detail = 'Variable';
            item.insertText = variable;
            item.sortText = `1_${variable}`;
            return item;
        });
    }

    /**
     * Get secret completions
     */
    getSecretCompletions(document: vscode.TextDocument): vscode.CompletionItem[] {
        const secrets = this.extractSecretsFromDocument(document);
        return secrets.map(secret => {
            const item = new vscode.CompletionItem(`SECRETS.${secret}`, vscode.CompletionItemKind.Constant);
            item.detail = 'Secret Variable';
            item.insertText = `SECRETS.${secret}`;
            item.sortText = `2_${secret}`;
            item.documentation = new vscode.MarkdownString('🔒 Secret variable (masked in output)');
            return item;
        });
    }

    /**
     * Get field completions based on context
     */
    getFieldCompletions(context: string): vscode.CompletionItem[] {
        const fields: { [key: string]: vscode.CompletionItem[] } = {
            'testcase': [
                this.createFieldCompletion('testcase', 'Test case name', '"${1:Test Name}"'),
                this.createFieldCompletion('description', 'Test description', '"${1:Test description}"'),
                this.createFieldCompletion('variables', 'Variable definitions', '{}'),
                this.createFieldCompletion('steps', 'Test steps', '[]'),
                this.createFieldCompletion('templates', 'Template definitions', '{}'),
                this.createFieldCompletion('parallel', 'Parallel execution config', '{}'),
            ],
            'testsuite': [
                this.createFieldCompletion('testsuite', 'Test suite name', '"${1:Suite Name}"'),
                this.createFieldCompletion('description', 'Suite description', '"${1:Suite description}"'),
                this.createFieldCompletion('setup', 'Setup steps', '[]'),
                this.createFieldCompletion('teardown', 'Teardown steps', '[]'),
                this.createFieldCompletion('testcases', 'Test case files', '[]'),
                this.createFieldCompletion('parallel', 'Parallel execution config', '{}'),
            ],
            'variables': [
                this.createFieldCompletion('vars', 'Regular variables', '{}'),
                this.createFieldCompletion('secrets', 'Secret variables', '{}'),
            ],
            'step': [
                this.createFieldCompletion('name', 'Step name', '"${1:Step Name}"'),
                this.createFieldCompletion('action', 'Action to execute', '"${1:action}"'),
                this.createFieldCompletion('args', 'Action arguments', '[]'),
                this.createFieldCompletion('result', 'Result variable name', '"${1:result}"'),
                this.createFieldCompletion('if', 'Conditional execution', '"${1:condition}"'),
                this.createFieldCompletion('retry', 'Retry configuration', '{}'),
                this.createFieldCompletion('timeout', 'Step timeout', '"${1:30s}"'),
            ]
        };

        return fields[context] || [];
    }

    /**
     * Validate document syntax
     */
    validateDocument(document: vscode.TextDocument): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];
        const text = document.getText();
        const lines = text.split('\n');

        try {
            // Basic YAML structure validation
            this.validateYAMLStructure(lines, diagnostics);
            
            // Robogo-specific validation
            this.validateRobogoSyntax(lines, diagnostics);
            
            // Action validation
            this.validateActions(lines, diagnostics);
            
            // Variable validation
            this.validateVariables(lines, diagnostics);
            
        } catch (error) {
            const diagnostic = new vscode.Diagnostic(
                new vscode.Range(0, 0, 0, 0),
                `Validation error: ${error}`,
                vscode.DiagnosticSeverity.Error
            );
            diagnostics.push(diagnostic);
        }

        return diagnostics;
    }

    /**
     * Extract variables from document
     */
    private extractVariablesFromDocument(document: vscode.TextDocument): string[] {
        const text = document.getText();
        const variables = new Set<string>();
        
        // Extract from variables section
        const varsMatch = text.match(/vars:\s*\n([\s\S]*?)(?=\n\S|\n*$)/);
        if (varsMatch) {
            const varLines = varsMatch[1].split('\n');
            for (const line of varLines) {
                const match = line.match(/^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:/);
                if (match) {
                    variables.add(match[1]);
                }
            }
        }

        // Extract from ${} references
        const varReferences = text.match(/\$\{([^}]+)\}/g);
        if (varReferences) {
            for (const ref of varReferences) {
                const varName = ref.slice(2, -1).split('.')[0];
                if (!varName.startsWith('SECRETS.') && varName !== '__robogo_steps') {
                    variables.add(varName);
                }
            }
        }

        return Array.from(variables);
    }

    /**
     * Extract secrets from document
     */
    private extractSecretsFromDocument(document: vscode.TextDocument): string[] {
        const text = document.getText();
        const secrets = new Set<string>();
        
        // Extract from secrets section
        const secretsMatch = text.match(/secrets:\s*\n([\s\S]*?)(?=\n\S|\n*$)/);
        if (secretsMatch) {
            const secretLines = secretsMatch[1].split('\n');
            for (const line of secretLines) {
                const match = line.match(/^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:/);
                if (match) {
                    secrets.add(match[1]);
                }
            }
        }

        return Array.from(secrets);
    }

    /**
     * Create field completion item
     */
    private createFieldCompletion(name: string, description: string, snippet: string): vscode.CompletionItem {
        const item = new vscode.CompletionItem(name, vscode.CompletionItemKind.Field);
        item.detail = description;
        item.insertText = new vscode.SnippetString(`${name}: ${snippet}`);
        item.sortText = `3_${name}`;
        return item;
    }

    /**
     * Validate YAML structure
     */
    private validateYAMLStructure(lines: string[], diagnostics: vscode.Diagnostic[]): void {
        let indentLevel = 0;
        let inList = false;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed === '' || trimmed.startsWith('#')) continue;

            // Check indentation
            const currentIndent = line.length - line.trimStart().length;
            
            // Validate list items
            if (trimmed.startsWith('-')) {
                if (!inList && currentIndent > 0) {
                    const diagnostic = new vscode.Diagnostic(
                        new vscode.Range(i, 0, i, line.length),
                        'Inconsistent list indentation',
                        vscode.DiagnosticSeverity.Warning
                    );
                    diagnostics.push(diagnostic);
                }
                inList = true;
            }

            // Check for missing colons in key-value pairs
            if (trimmed.includes(':') && !trimmed.startsWith('-')) {
                const colonIndex = trimmed.indexOf(':');
                if (colonIndex === trimmed.length - 1 && i < lines.length - 1) {
                    const nextLine = lines[i + 1];
                    if (nextLine.trim() !== '' && !nextLine.startsWith(' ') && !nextLine.startsWith('\t')) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i + 1, 0, i + 1, nextLine.length),
                            'Expected indented block after colon',
                            vscode.DiagnosticSeverity.Error
                        );
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }
    }

    /**
     * Validate Robogo-specific syntax
     */
    private validateRobogoSyntax(lines: string[], diagnostics: vscode.Diagnostic[]): void {
        let hasTestCaseOrSuite = false;
        let hasSteps = false;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Check for required root fields
            if (trimmed.startsWith('testcase:') || trimmed.startsWith('testsuite:')) {
                hasTestCaseOrSuite = true;
            }

            if (trimmed.startsWith('steps:') || trimmed.startsWith('testcases:')) {
                hasSteps = true;
            }

            // Validate variable references
            const varReferences = trimmed.match(/\$\{([^}]+)\}/g);
            if (varReferences) {
                for (const ref of varReferences) {
                    const varName = ref.slice(2, -1);
                    if (varName === '') {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(ref), i, line.indexOf(ref) + ref.length),
                            'Empty variable reference',
                            vscode.DiagnosticSeverity.Error
                        );
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }

        // Check for required structure
        if (!hasTestCaseOrSuite) {
            const diagnostic = new vscode.Diagnostic(
                new vscode.Range(0, 0, 0, 0),
                'File must contain either "testcase:" or "testsuite:" declaration',
                vscode.DiagnosticSeverity.Error
            );
            diagnostics.push(diagnostic);
        }
    }

    /**
     * Validate actions
     */
    private validateActions(lines: string[], diagnostics: vscode.Diagnostic[]): void {
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed.startsWith('action:')) {
                const actionMatch = trimmed.match(/action:\s*["']?([^"'\s]+)["']?/);
                if (actionMatch) {
                    const actionName = actionMatch[1];
                    if (!this.actionRegistry.hasAction(actionName)) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(actionName), i, line.indexOf(actionName) + actionName.length),
                            `Unknown action: ${actionName}`,
                            vscode.DiagnosticSeverity.Error
                        );
                        diagnostic.source = 'robogo';
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }
    }

    /**
     * Validate variables
     */
    private validateVariables(lines: string[], diagnostics: vscode.Diagnostic[]): void {
        const definedVars = new Set<string>();
        const referencedVars = new Set<string>();

        // First pass: collect defined variables
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Variables in vars section
            if (trimmed.match(/^\s*[a-zA-Z_][a-zA-Z0-9_]*\s*:/)) {
                const varMatch = trimmed.match(/^([a-zA-Z_][a-zA-Z0-9_]*)\s*:/);
                if (varMatch) {
                    definedVars.add(varMatch[1]);
                }
            }
        }

        // Second pass: collect referenced variables
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const varReferences = line.match(/\$\{([^}]+)\}/g);
            
            if (varReferences) {
                for (const ref of varReferences) {
                    const varName = ref.slice(2, -1).split('.')[0];
                    if (!varName.startsWith('SECRETS.') && varName !== '__robogo_steps') {
                        referencedVars.add(varName);
                        
                        // Check if variable is defined
                        if (!definedVars.has(varName)) {
                            const diagnostic = new vscode.Diagnostic(
                                new vscode.Range(i, line.indexOf(ref), i, line.indexOf(ref) + ref.length),
                                `Undefined variable: ${varName}`,
                                vscode.DiagnosticSeverity.Warning
                            );
                            diagnostic.source = 'robogo';
                            diagnostics.push(diagnostic);
                        }
                    }
                }
            }
        }
    }
}