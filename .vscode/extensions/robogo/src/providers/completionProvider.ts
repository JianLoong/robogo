import * as vscode from 'vscode';
import { RobogoLanguageServer } from '../core/languageServer';

/**
 * Provides intelligent completion suggestions for Robogo files
 */
export class CompletionProvider implements 
    vscode.CompletionItemProvider, 
    vscode.DocumentSymbolProvider,
    vscode.DefinitionProvider,
    vscode.CodeLensProvider {

    constructor(private languageServer: RobogoLanguageServer) {}

    /**
     * Provide completion items
     */
    async provideCompletionItems(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken,
        context: vscode.CompletionContext
    ): Promise<vscode.CompletionItem[]> {
        const line = document.lineAt(position);
        const lineText = line.text;
        const beforeCursor = lineText.substring(0, position.character);
        const afterCursor = lineText.substring(position.character);

        // Determine completion context
        const completionContext = this.getCompletionContext(document, position);
        
        let completions: vscode.CompletionItem[] = [];

        // Action completions
        if (this.shouldProvideActionCompletions(beforeCursor, completionContext)) {
            completions.push(...this.languageServer.getActionCompletions());
        }

        // Field completions based on context
        if (this.shouldProvideFieldCompletions(beforeCursor, completionContext)) {
            completions.push(...this.languageServer.getFieldCompletions(completionContext.type));
        }

        // Variable completions
        if (this.shouldProvideVariableCompletions(beforeCursor)) {
            completions.push(...this.languageServer.getVariableCompletions(document));
            completions.push(...this.languageServer.getSecretCompletions(document));
        }

        // Template completions
        if (this.shouldProvideTemplateCompletions(beforeCursor, completionContext)) {
            completions.push(...this.getTemplateCompletions());
        }

        // Value completions
        if (this.shouldProvideValueCompletions(beforeCursor, completionContext)) {
            completions.push(...this.getValueCompletions(completionContext.field));
        }

        // Snippet completions
        if (this.shouldProvideSnippetCompletions(beforeCursor, completionContext)) {
            completions.push(...this.getSnippetCompletions());
        }

        return completions;
    }

    /**
     * Provide document symbols for outline
     */
    async provideDocumentSymbols(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): Promise<vscode.DocumentSymbol[]> {
        const symbols: vscode.DocumentSymbol[] = [];
        const text = document.getText();
        const lines = text.split('\n');

        let currentTestCase: vscode.DocumentSymbol | null = null;
        let currentSteps: vscode.DocumentSymbol | null = null;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Test case/suite
            if (trimmed.startsWith('testcase:') || trimmed.startsWith('testsuite:')) {
                const name = trimmed.split(':')[1]?.trim().replace(/['"]/g, '') || 'Unnamed';
                const range = new vscode.Range(i, 0, i, line.length);
                currentTestCase = new vscode.DocumentSymbol(
                    name,
                    trimmed.startsWith('testcase:') ? 'Test Case' : 'Test Suite',
                    vscode.SymbolKind.Class,
                    range,
                    range
                );
                symbols.push(currentTestCase);
            }

            // Steps section
            if (trimmed.startsWith('steps:') && currentTestCase) {
                const range = new vscode.Range(i, 0, i, line.length);
                currentSteps = new vscode.DocumentSymbol(
                    'Steps',
                    'Test Steps',
                    vscode.SymbolKind.Method,
                    range,
                    range
                );
                currentTestCase.children.push(currentSteps);
            }

            // Individual steps
            if (trimmed.startsWith('- name:') && currentSteps) {
                const nameMatch = trimmed.match(/- name:\s*["']?([^"']+)["']?/);
                if (nameMatch) {
                    const stepName = nameMatch[1];
                    const range = new vscode.Range(i, 0, i, line.length);
                    const stepSymbol = new vscode.DocumentSymbol(
                        stepName,
                        'Step',
                        vscode.SymbolKind.Function,
                        range,
                        range
                    );
                    currentSteps.children.push(stepSymbol);
                }
            }

            // Variables section
            if (trimmed.startsWith('variables:') && currentTestCase) {
                const range = new vscode.Range(i, 0, i, line.length);
                const varsSymbol = new vscode.DocumentSymbol(
                    'Variables',
                    'Variable Definitions',
                    vscode.SymbolKind.Namespace,
                    range,
                    range
                );
                currentTestCase.children.push(varsSymbol);
            }
        }

        return symbols;
    }

    /**
     * Provide definition locations
     */
    async provideDefinition(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Definition | null> {
        const wordRange = document.getWordRangeAtPosition(position);
        if (!wordRange) return null;

        const word = document.getText(wordRange);
        const line = document.lineAt(position);

        // Variable definition lookup
        if (line.text.includes('${' + word + '}') || line.text.includes('SECRETS.' + word)) {
            const varDefinition = this.findVariableDefinition(document, word);
            if (varDefinition) {
                return varDefinition;
            }
        }

        // Action definition lookup
        if (line.text.includes('action:') && line.text.includes(word)) {
            const actionDef = this.languageServer.getActionRegistry().getAction(word);
            if (actionDef) {
                // Return hover info as definition (VS Code will show it)
                return new vscode.Location(document.uri, wordRange);
            }
        }

        return null;
    }

    /**
     * Provide code lenses for test execution
     */
    async provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): Promise<vscode.CodeLens[]> {
        const codeLenses: vscode.CodeLens[] = [];
        const text = document.getText();
        const lines = text.split('\n');

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Test case execution
            if (trimmed.startsWith('testcase:')) {
                const range = new vscode.Range(i, 0, i, line.length);
                
                // Run test
                codeLenses.push(new vscode.CodeLens(range, {
                    title: '▶️ Run Test',
                    command: 'robogo.runTest',
                    arguments: [document.uri]
                }));

                // Run test parallel
                codeLenses.push(new vscode.CodeLens(range, {
                    title: '⚡ Run Parallel',
                    command: 'robogo.runTestParallel',
                    arguments: [document.uri]
                }));

                // Debug test
                codeLenses.push(new vscode.CodeLens(range, {
                    title: '🐛 Debug',
                    command: 'robogo.debugTest',
                    arguments: [document.uri]
                }));
            }

            // Test suite execution
            if (trimmed.startsWith('testsuite:')) {
                const range = new vscode.Range(i, 0, i, line.length);
                
                codeLenses.push(new vscode.CodeLens(range, {
                    title: '▶️ Run Suite',
                    command: 'robogo.runTestSuite',
                    arguments: [document.uri]
                }));
            }

            // Step execution
            if (trimmed.startsWith('- name:')) {
                const range = new vscode.Range(i, 0, i, line.length);
                
                codeLenses.push(new vscode.CodeLens(range, {
                    title: '▶️ Run Step',
                    command: 'robogo.runStep',
                    arguments: [document.uri, i]
                }));
            }
        }

        return codeLenses;
    }

    /**
     * Get completion context based on cursor position
     */
    private getCompletionContext(document: vscode.TextDocument, position: vscode.Position): CompletionContext {
        const text = document.getText();
        const beforeCursor = text.substring(0, document.offsetAt(position));
        
        // Determine if we're in a test case or test suite
        const isTestSuite = /testsuite:/m.test(beforeCursor);
        const isTestCase = /testcase:/m.test(beforeCursor);
        
        // Determine current section
        let currentSection = 'root';
        if (/variables:\s*$/.test(beforeCursor)) currentSection = 'variables';
        if (/steps:\s*$/.test(beforeCursor)) currentSection = 'steps';
        if (/- name:/.test(beforeCursor)) currentSection = 'step';
        
        // Determine current field context
        const line = document.lineAt(position).text;
        let currentField = '';
        if (line.includes('action:')) currentField = 'action';
        if (line.includes('args:')) currentField = 'args';
        if (line.includes('result:')) currentField = 'result';
        
        return {
            type: isTestSuite ? 'testsuite' : (isTestCase ? 'testcase' : 'unknown'),
            section: currentSection,
            field: currentField,
            indentLevel: this.getIndentLevel(document.lineAt(position).text)
        };
    }

    /**
     * Check if should provide action completions
     */
    private shouldProvideActionCompletions(beforeCursor: string, context: CompletionContext): boolean {
        return beforeCursor.trim().endsWith('action:') || 
               (beforeCursor.includes('action:') && beforeCursor.trim().endsWith(' '));
    }

    /**
     * Check if should provide field completions
     */
    private shouldProvideFieldCompletions(beforeCursor: string, context: CompletionContext): boolean {
        const line = beforeCursor.split('\n').pop() || '';
        return line.trim() === '' || line.endsWith(':') || line.endsWith(' ');
    }

    /**
     * Check if should provide variable completions
     */
    private shouldProvideVariableCompletions(beforeCursor: string): boolean {
        return beforeCursor.includes('${') || beforeCursor.endsWith('SECRETS.');
    }

    /**
     * Check if should provide template completions
     */
    private shouldProvideTemplateCompletions(beforeCursor: string, context: CompletionContext): boolean {
        return context.field === 'args' && beforeCursor.includes('template');
    }

    /**
     * Check if should provide value completions
     */
    private shouldProvideValueCompletions(beforeCursor: string, context: CompletionContext): boolean {
        return context.field !== '' && beforeCursor.endsWith('"');
    }

    /**
     * Check if should provide snippet completions
     */
    private shouldProvideSnippetCompletions(beforeCursor: string, context: CompletionContext): boolean {
        const line = beforeCursor.split('\n').pop() || '';
        return line.trim() === '' && context.indentLevel === 0;
    }

    /**
     * Get template completions
     */
    private getTemplateCompletions(): vscode.CompletionItem[] {
        const templates = [
            'templates/mt103.tmpl',
            'templates/mt202.tmpl',
            'templates/mt900.tmpl',
            'templates/mt910.tmpl',
            'templates/sepa-credit-transfer.xml.tmpl'
        ];

        return templates.map(template => {
            const item = new vscode.CompletionItem(template, vscode.CompletionItemKind.File);
            item.detail = 'Template File';
            item.insertText = `"${template}"`;
            return item;
        });
    }

    /**
     * Get value completions based on field
     */
    private getValueCompletions(field: string): vscode.CompletionItem[] {
        const valueMap: { [key: string]: string[] } = {
            'action': ['http', 'postgres', 'spanner', 'kafka', 'rabbitmq', 'assert', 'log', 'variable', 'template'],
            'method': ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'],
            'level': ['info', 'debug', 'warn', 'error'],
            'operator': ['==', '!=', '>', '<', '>=', '<=', 'contains', 'not_contains', 'starts_with', 'ends_with']
        };

        const values = valueMap[field] || [];
        return values.map(value => {
            const item = new vscode.CompletionItem(value, vscode.CompletionItemKind.Value);
            item.insertText = `"${value}"`;
            return item;
        });
    }

    /**
     * Get snippet completions
     */
    private getSnippetCompletions(): vscode.CompletionItem[] {
        const snippets = [
            {
                name: 'testcase',
                description: 'Basic test case template',
                snippet: 'testcase: "${1:Test Name}"\ndescription: "${2:Test description}"\n\nvariables:\n  vars:\n    ${3:variable_name}: "${4:value}"\n\nsteps:\n  - name: "${5:Step name}"\n    action: ${6:http}\n    args: [${7:"GET", "https://api.example.com"}]\n    result: ${8:response}'
            },
            {
                name: 'testsuite',
                description: 'Test suite template',
                snippet: 'testsuite: "${1:Suite Name}"\ndescription: "${2:Suite description}"\n\nsetup:\n  - name: "${3:Setup step}"\n    action: log\n    args: ["Setting up..."]\n\nteardown:\n  - name: "${4:Teardown step}"\n    action: log\n    args: ["Cleaning up..."]\n\ntestcases:\n  - ${5:test1.robogo}\n  - ${6:test2.robogo}'
            },
            {
                name: 'http-step',
                description: 'HTTP request step',
                snippet: '- name: "${1:HTTP request}"\n  action: http\n  args: ["${2:GET}", "${3:https://api.example.com}"]\n  result: ${4:response}'
            },
            {
                name: 'assert-step',
                description: 'Assertion step',
                snippet: '- name: "${1:Assert condition}"\n  action: assert\n  args: ["${2:actual}", "${3:==}", "${4:expected}"${5:, "${6:message}"}]'
            }
        ];

        return snippets.map(snippet => {
            const item = new vscode.CompletionItem(snippet.name, vscode.CompletionItemKind.Snippet);
            item.detail = snippet.description;
            item.insertText = new vscode.SnippetString(snippet.snippet);
            item.sortText = `9_${snippet.name}`;
            return item;
        });
    }

    /**
     * Find variable definition in document
     */
    private findVariableDefinition(document: vscode.TextDocument, variableName: string): vscode.Location | null {
        const text = document.getText();
        const lines = text.split('\n');

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const regex = new RegExp(`^\\s*${variableName}\\s*:`);
            if (regex.test(line)) {
                const range = new vscode.Range(i, 0, i, line.length);
                return new vscode.Location(document.uri, range);
            }
        }

        return null;
    }

    /**
     * Get indentation level of line
     */
    private getIndentLevel(line: string): number {
        return line.length - line.trimStart().length;
    }
}

/**
 * Completion context information
 */
interface CompletionContext {
    type: 'testcase' | 'testsuite' | 'unknown';
    section: string;
    field: string;
    indentLevel: number;
}