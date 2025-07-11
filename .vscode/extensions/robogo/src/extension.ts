import * as vscode from 'vscode';
import { CompletionProvider } from './providers/completionProvider';
import { HoverProvider } from './providers/hoverProvider';
import { DiagnosticProvider } from './providers/diagnosticProvider';
import { CommandManager } from './commands/commandManager';
import { ConfigurationManager } from './core/configurationManager';
import { RobogoLanguageServer } from './core/languageServer';

/**
 * VS Code Extension for Robogo Test Automation Framework
 * 
 * Features:
 * - Intelligent syntax highlighting and autocomplete
 * - Real-time validation and error detection
 * - Integrated test execution with multiple output formats
 * - Action documentation and hover tooltips
 * - Code snippets and templates
 * - Test suite management
 * - Parallel execution support
 */
export function activate(context: vscode.ExtensionContext) {
    console.log('🚀 Robogo Extension v2.0 - Activating...');

    try {
        // Initialize core services
        const config = new ConfigurationManager();
        const languageServer = new RobogoLanguageServer(config);
        
        // Initialize providers
        const completionProvider = new CompletionProvider(languageServer);
        const hoverProvider = new HoverProvider(languageServer);
        const diagnosticProvider = new DiagnosticProvider(languageServer);
        
        // Initialize command manager
        const commandManager = new CommandManager(config, languageServer);

        // Register language providers
        registerLanguageProviders(context, completionProvider, hoverProvider, diagnosticProvider);
        
        // Register commands
        commandManager.registerCommands(context);
        
        // Setup file watchers and event handlers
        setupEventHandlers(context, diagnosticProvider);

        console.log('✅ Robogo Extension v2.0 - Successfully activated!');
        
        // Show welcome message on first activation
        showWelcomeMessage(context);
        
    } catch (error) {
        console.error('❌ Failed to activate Robogo extension:', error);
        vscode.window.showErrorMessage(`Failed to activate Robogo extension: ${error}`);
    }
}

/**
 * Register all language providers
 */
function registerLanguageProviders(
    context: vscode.ExtensionContext,
    completionProvider: CompletionProvider,
    hoverProvider: HoverProvider,
    diagnosticProvider: DiagnosticProvider
) {
    const robogoSelector = [
        { scheme: 'file', language: 'robogo' },
        { scheme: 'file', pattern: '**/*.robogo' }
    ];

    // Completion provider with enhanced triggers
    context.subscriptions.push(
        vscode.languages.registerCompletionItemProvider(
            robogoSelector,
            completionProvider,
            ':', ' ', '-', '$', '{', '"', "'", '\n', '\t'
        )
    );

    // Hover provider for documentation
    context.subscriptions.push(
        vscode.languages.registerHoverProvider(robogoSelector, hoverProvider)
    );

    // Document symbol provider for outline
    context.subscriptions.push(
        vscode.languages.registerDocumentSymbolProvider(robogoSelector, completionProvider)
    );

    // Definition provider for navigation
    context.subscriptions.push(
        vscode.languages.registerDefinitionProvider(robogoSelector, completionProvider)
    );

    // Code lens provider for test execution
    context.subscriptions.push(
        vscode.languages.registerCodeLensProvider(robogoSelector, completionProvider)
    );

    // Diagnostic collection for validation
    context.subscriptions.push(diagnosticProvider.diagnosticCollection);
}

/**
 * Setup file watchers and event handlers
 */
function setupEventHandlers(context: vscode.ExtensionContext, diagnosticProvider: DiagnosticProvider) {
    // Real-time validation on file changes
    const onDidChangeTextDocument = vscode.workspace.onDidChangeTextDocument((event) => {
        if (event.document.languageId === 'robogo' || event.document.fileName.endsWith('.robogo')) {
            diagnosticProvider.validateDocument(event.document);
        }
    });

    // Validation on file open
    const onDidOpenTextDocument = vscode.workspace.onDidOpenTextDocument((document) => {
        if (document.languageId === 'robogo' || document.fileName.endsWith('.robogo')) {
            diagnosticProvider.validateDocument(document);
        }
    });

    // Clear diagnostics on file close
    const onDidCloseTextDocument = vscode.workspace.onDidCloseTextDocument((document) => {
        if (document.languageId === 'robogo' || document.fileName.endsWith('.robogo')) {
            diagnosticProvider.clearDiagnostics(document.uri);
        }
    });

    // Configuration changes
    const onDidChangeConfiguration = vscode.workspace.onDidChangeConfiguration((event) => {
        if (event.affectsConfiguration('robogo')) {
            vscode.window.showInformationMessage('Robogo configuration updated. Some changes may require restart.');
        }
    });

    context.subscriptions.push(
        onDidChangeTextDocument,
        onDidOpenTextDocument,
        onDidCloseTextDocument,
        onDidChangeConfiguration
    );
}

/**
 * Show welcome message for new users
 */
function showWelcomeMessage(context: vscode.ExtensionContext) {
    const hasShownWelcome = context.globalState.get('robogo.hasShownWelcome', false);
    
    if (!hasShownWelcome) {
        const welcomeMessage = '🎉 Welcome to Robogo VS Code Extension! Get started by opening a .robogo file or run "Robogo: Generate Test Template" to create your first test.';
        
        vscode.window.showInformationMessage(
            welcomeMessage,
            'Show Documentation',
            'Generate Template',
            'Dismiss'
        ).then(selection => {
            switch (selection) {
                case 'Show Documentation':
                    vscode.commands.executeCommand('robogo.showDocumentation');
                    break;
                case 'Generate Template':
                    vscode.commands.executeCommand('robogo.generateTemplate');
                    break;
            }
        });

        context.globalState.update('robogo.hasShownWelcome', true);
    }
}

/**
 * Deactivate extension
 */
export function deactivate() {
    console.log('👋 Robogo Extension - Deactivating...');
    // Cleanup if needed
}