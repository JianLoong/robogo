import * as vscode from 'vscode';
import { ConfigurationManager } from '../core/configurationManager';
import { RobogoLanguageServer } from '../core/languageServer';
import { TestExecutor } from './testExecutor';
import { TemplateGenerator } from './templateGenerator';
import { DocumentationProvider } from './documentationProvider';

/**
 * Manages all Robogo extension commands
 */
export class CommandManager {
    private testExecutor: TestExecutor;
    private templateGenerator: TemplateGenerator;
    private documentationProvider: DocumentationProvider;

    constructor(
        private config: ConfigurationManager,
        private languageServer: RobogoLanguageServer
    ) {
        this.testExecutor = new TestExecutor(config);
        this.templateGenerator = new TemplateGenerator();
        this.documentationProvider = new DocumentationProvider(languageServer);
    }

    /**
     * Register all commands
     */
    registerCommands(context: vscode.ExtensionContext): void {
        // Test execution commands
        context.subscriptions.push(
            vscode.commands.registerCommand('robogo.runTest', () => this.runTest()),
            vscode.commands.registerCommand('robogo.runTestParallel', () => this.runTestParallel()),
            vscode.commands.registerCommand('robogo.runTestSuite', () => this.runTestSuite()),
            vscode.commands.registerCommand('robogo.runWithOutput', () => this.runWithOutput()),
            vscode.commands.registerCommand('robogo.debugTest', (uri?: vscode.Uri) => this.debugTest(uri))
        );

        // Utility commands
        context.subscriptions.push(
            vscode.commands.registerCommand('robogo.listActions', () => this.listActions()),
            vscode.commands.registerCommand('robogo.validateSyntax', () => this.validateSyntax()),
            vscode.commands.registerCommand('robogo.validateTDM', () => this.validateTDM())
        );

        // Template and generation commands
        context.subscriptions.push(
            vscode.commands.registerCommand('robogo.generateTemplate', () => this.generateTemplate()),
            vscode.commands.registerCommand('robogo.generateTestCase', () => this.generateTestCase()),
            vscode.commands.registerCommand('robogo.generateTestSuite', () => this.generateTestSuite())
        );

        // Documentation commands
        context.subscriptions.push(
            vscode.commands.registerCommand('robogo.showDocumentation', () => this.showDocumentation()),
            vscode.commands.registerCommand('robogo.showActionHelp', (actionName?: string) => this.showActionHelp(actionName)),
            vscode.commands.registerCommand('robogo.showExamples', () => this.showExamples())
        );

        // Configuration commands
        context.subscriptions.push(
            vscode.commands.registerCommand('robogo.openSettings', () => this.openSettings()),
            vscode.commands.registerCommand('robogo.resetConfiguration', () => this.resetConfiguration()),
            vscode.commands.registerCommand('robogo.showConfiguration', () => this.showConfiguration())
        );

        // Status bar commands
        this.createStatusBarItems(context);
    }

    /**
     * Run current test file
     */
    private async runTest(): Promise<void> {
        const editor = vscode.window.activeTextEditor;
        if (!editor || !this.isRobogoFile(editor.document)) {
            vscode.window.showErrorMessage('Please open a .robogo file to run tests.');
            return;
        }

        await this.testExecutor.runTest(editor.document.uri, {
            parallel: false,
            outputFormat: this.config.getOutputFormat(),
            verbose: this.config.isVerboseOutputEnabled()
        });
    }

    /**
     * Run test with parallel execution
     */
    private async runTestParallel(): Promise<void> {
        const editor = vscode.window.activeTextEditor;
        if (!editor || !this.isRobogoFile(editor.document)) {
            vscode.window.showErrorMessage('Please open a .robogo file to run tests.');
            return;
        }

        await this.testExecutor.runTest(editor.document.uri, {
            parallel: true,
            maxConcurrency: this.config.getMaxConcurrency(),
            outputFormat: this.config.getOutputFormat(),
            verbose: this.config.isVerboseOutputEnabled()
        });
    }

    /**
     * Run test suite
     */
    private async runTestSuite(): Promise<void> {
        const editor = vscode.window.activeTextEditor;
        if (!editor || !this.isRobogoFile(editor.document)) {
            vscode.window.showErrorMessage('Please open a .robogo file to run test suite.');
            return;
        }

        await this.testExecutor.runTestSuite(editor.document.uri, {
            parallel: this.config.isParallelExecutionEnabled(),
            maxConcurrency: this.config.getMaxConcurrency(),
            outputFormat: this.config.getOutputFormat(),
            verbose: this.config.isVerboseOutputEnabled()
        });
    }

    /**
     * Run test with custom output format
     */
    private async runWithOutput(): Promise<void> {
        const outputFormat = await vscode.window.showQuickPick(
            ['console', 'json', 'markdown'],
            { placeHolder: 'Select output format' }
        );

        if (!outputFormat) return;

        const editor = vscode.window.activeTextEditor;
        if (!editor || !this.isRobogoFile(editor.document)) {
            vscode.window.showErrorMessage('Please open a .robogo file to run tests.');
            return;
        }

        await this.testExecutor.runTest(editor.document.uri, {
            parallel: this.config.isParallelExecutionEnabled(),
            outputFormat: outputFormat as 'console' | 'json' | 'markdown',
            verbose: this.config.isVerboseOutputEnabled()
        });
    }

    /**
     * Debug test with variable inspection
     */
    private async debugTest(uri?: vscode.Uri): Promise<void> {
        const targetUri = uri || vscode.window.activeTextEditor?.document.uri;
        if (!targetUri) {
            vscode.window.showErrorMessage('No file selected for debugging.');
            return;
        }

        await this.testExecutor.debugTest(targetUri);
    }

    /**
     * List all available actions
     */
    private async listActions(): Promise<void> {
        const actions = this.languageServer.getActionRegistry().getAllActions();
        const actionsByCategory = this.groupByCategory(actions);

        const panel = vscode.window.createWebviewPanel(
            'robogoActions',
            'Robogo Actions',
            vscode.ViewColumn.Beside,
            { enableScripts: true }
        );

        panel.webview.html = this.generateActionsHTML(actionsByCategory);
    }

    /**
     * Validate current file syntax
     */
    private async validateSyntax(): Promise<void> {
        const editor = vscode.window.activeTextEditor;
        if (!editor || !this.isRobogoFile(editor.document)) {
            vscode.window.showErrorMessage('Please open a .robogo file to validate syntax.');
            return;
        }

        const diagnostics = this.languageServer.validateDocument(editor.document);
        
        if (diagnostics.length === 0) {
            vscode.window.showInformationMessage('✅ Syntax validation passed - no issues found!');
        } else {
            const errorCount = diagnostics.filter(d => d.severity === vscode.DiagnosticSeverity.Error).length;
            const warningCount = diagnostics.filter(d => d.severity === vscode.DiagnosticSeverity.Warning).length;
            
            vscode.window.showWarningMessage(
                `Validation found ${errorCount} errors and ${warningCount} warnings. Check Problems panel for details.`
            );
        }
    }

    /**
     * Validate TDM configuration
     */
    private async validateTDM(): Promise<void> {
        vscode.window.showInformationMessage('TDM validation is not yet implemented in this version.');
    }

    /**
     * Generate new template
     */
    private async generateTemplate(): Promise<void> {
        const templateType = await vscode.window.showQuickPick([
            { label: 'Test Case', value: 'testcase', description: 'Basic test case template' },
            { label: 'Test Suite', value: 'testsuite', description: 'Test suite with multiple test cases' },
            { label: 'HTTP API Test', value: 'http', description: 'API testing template' },
            { label: 'Database Test', value: 'database', description: 'Database operations template' },
            { label: 'Integration Test', value: 'integration', description: 'Complex integration test template' }
        ], { placeHolder: 'Select template type' });

        if (!templateType) return;

        await this.templateGenerator.generateTemplate(templateType.value);
    }

    /**
     * Generate test case
     */
    private async generateTestCase(): Promise<void> {
        await this.templateGenerator.generateTestCase();
    }

    /**
     * Generate test suite
     */
    private async generateTestSuite(): Promise<void> {
        await this.templateGenerator.generateTestSuite();
    }

    /**
     * Show documentation
     */
    private async showDocumentation(): Promise<void> {
        await this.documentationProvider.showDocumentation();
    }

    /**
     * Show action-specific help
     */
    private async showActionHelp(actionName?: string): Promise<void> {
        if (!actionName) {
            const actions = this.languageServer.getActionRegistry().getAllActions();
            const selected = await vscode.window.showQuickPick(
                actions.map(action => ({
                    label: action.name,
                    description: action.description,
                    detail: action.category
                })),
                { placeHolder: 'Select action for help' }
            );
            
            if (!selected) return;
            actionName = selected.label;
        }

        await this.documentationProvider.showActionHelp(actionName);
    }

    /**
     * Show examples
     */
    private async showExamples(): Promise<void> {
        await this.documentationProvider.showExamples();
    }

    /**
     * Open extension settings
     */
    private async openSettings(): Promise<void> {
        await vscode.commands.executeCommand('workbench.action.openSettings', 'robogo');
    }

    /**
     * Reset configuration to defaults
     */
    private async resetConfiguration(): Promise<void> {
        const confirm = await vscode.window.showWarningMessage(
            'Reset all Robogo configuration to defaults?',
            'Reset',
            'Cancel'
        );

        if (confirm === 'Reset') {
            await this.config.resetToDefaults();
            vscode.window.showInformationMessage('Robogo configuration reset to defaults.');
        }
    }

    /**
     * Show current configuration
     */
    private async showConfiguration(): Promise<void> {
        const config = this.config.getAllConfig();
        const validation = this.config.validateConfiguration();
        
        const panel = vscode.window.createWebviewPanel(
            'robogoConfig',
            'Robogo Configuration',
            vscode.ViewColumn.Beside,
            { enableScripts: true }
        );

        panel.webview.html = this.generateConfigHTML(config, validation);
    }

    /**
     * Create status bar items
     */
    private createStatusBarItems(context: vscode.ExtensionContext): void {
        // Run test status bar item
        const runTestStatusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
        runTestStatusBar.text = '$(play) Run Test';
        runTestStatusBar.command = 'robogo.runTest';
        runTestStatusBar.tooltip = 'Run current Robogo test';
        context.subscriptions.push(runTestStatusBar);

        // Show status bar only for Robogo files
        const updateStatusBar = () => {
            const editor = vscode.window.activeTextEditor;
            if (editor && this.isRobogoFile(editor.document)) {
                runTestStatusBar.show();
            } else {
                runTestStatusBar.hide();
            }
        };

        context.subscriptions.push(
            vscode.window.onDidChangeActiveTextEditor(updateStatusBar),
            vscode.workspace.onDidOpenTextDocument(updateStatusBar)
        );

        updateStatusBar();
    }

    /**
     * Check if document is a Robogo file
     */
    private isRobogoFile(document: vscode.TextDocument): boolean {
        return document.languageId === 'robogo' || document.fileName.endsWith('.robogo');
    }

    /**
     * Group actions by category
     */
    private groupByCategory(actions: any[]): { [category: string]: any[] } {
        return actions.reduce((groups, action) => {
            const category = action.category || 'Other';
            if (!groups[category]) {
                groups[category] = [];
            }
            groups[category].push(action);
            return groups;
        }, {});
    }

    /**
     * Generate HTML for actions list
     */
    private generateActionsHTML(actionsByCategory: { [category: string]: any[] }): string {
        let html = `
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; padding: 20px; }
                .category { margin-bottom: 30px; }
                .category h2 { color: #007ACC; border-bottom: 2px solid #007ACC; padding-bottom: 10px; }
                .action { margin: 15px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
                .action h3 { margin: 0 0 10px 0; color: #333; }
                .action .description { color: #666; margin-bottom: 10px; }
                .action .example { background: #f5f5f5; padding: 10px; border-radius: 3px; font-family: monospace; font-size: 12px; }
                .parameters { margin: 10px 0; }
                .parameter { margin: 5px 0; font-size: 14px; }
                .required { color: #d73a49; font-weight: bold; }
                .optional { color: #6f42c1; }
            </style>
        </head>
        <body>
            <h1>🎯 Robogo Actions Reference</h1>
        `;

        for (const [category, actions] of Object.entries(actionsByCategory)) {
            html += `<div class="category"><h2>${category}</h2>`;
            
            for (const action of actions) {
                html += `
                <div class="action">
                    <h3>${action.name}</h3>
                    <div class="description">${action.description}</div>
                    <div class="parameters">
                        <strong>Parameters:</strong>
                        ${action.parameters.map((p: any) => `
                            <div class="parameter">
                                <span class="${p.required ? 'required' : 'optional'}">
                                    ${p.name} (${p.type})${p.required ? ' *' : ''}
                                </span>
                                : ${p.description}
                            </div>
                        `).join('')}
                    </div>
                    ${action.examples.length > 0 ? `
                        <div class="example">
                            <strong>Example:</strong><br>
                            <pre>${action.examples[0]}</pre>
                        </div>
                    ` : ''}
                </div>
                `;
            }
            
            html += '</div>';
        }

        html += '</body></html>';
        return html;
    }

    /**
     * Generate HTML for configuration display
     */
    private generateConfigHTML(config: any, validation: { isValid: boolean; errors: string[] }): string {
        return `
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; padding: 20px; }
                .status { padding: 10px; border-radius: 5px; margin-bottom: 20px; }
                .valid { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
                .invalid { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
                .config-item { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 3px; }
                .config-key { font-weight: bold; color: #007ACC; }
                .config-value { font-family: monospace; color: #333; }
                .error { color: #d73a49; margin: 5px 0; }
            </style>
        </head>
        <body>
            <h1>⚙️ Robogo Configuration</h1>
            
            <div class="status ${validation.isValid ? 'valid' : 'invalid'}">
                ${validation.isValid ? '✅ Configuration is valid' : '❌ Configuration has errors'}
                ${validation.errors.map(error => `<div class="error">${error}</div>`).join('')}
            </div>
            
            <h2>Current Settings</h2>
            ${Object.entries(config).map(([key, value]) => `
                <div class="config-item">
                    <div class="config-key">${key}</div>
                    <div class="config-value">${JSON.stringify(value, null, 2)}</div>
                </div>
            `).join('')}
        </body>
        </html>
        `;
    }
}