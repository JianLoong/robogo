import * as vscode from 'vscode';
import { RobogoLanguageServer } from '../core/languageServer';

/**
 * Provides real-time validation and error detection for Robogo files
 */
export class DiagnosticProvider {
    public readonly diagnosticCollection: vscode.DiagnosticCollection;
    private validationTimeout: NodeJS.Timeout | null = null;

    constructor(private languageServer: RobogoLanguageServer) {
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('robogo');
    }

    /**
     * Validate document and update diagnostics
     */
    validateDocument(document: vscode.TextDocument): void {
        // Debounce validation to avoid excessive calls
        if (this.validationTimeout) {
            clearTimeout(this.validationTimeout);
        }

        this.validationTimeout = setTimeout(() => {
            this.performValidation(document);
        }, 500);
    }

    /**
     * Clear diagnostics for a document
     */
    clearDiagnostics(uri: vscode.Uri): void {
        this.diagnosticCollection.delete(uri);
    }

    /**
     * Perform actual validation
     */
    private performValidation(document: vscode.TextDocument): void {
        try {
            const diagnostics = this.languageServer.validateDocument(document);
            
            // Add additional custom validations
            diagnostics.push(...this.validateCustomRules(document));
            
            this.diagnosticCollection.set(document.uri, diagnostics);
        } catch (error) {
            console.error('Error during validation:', error);
            
            // Show parsing error as diagnostic
            const diagnostic = new vscode.Diagnostic(
                new vscode.Range(0, 0, 0, 0),
                `Validation failed: ${error}`,
                vscode.DiagnosticSeverity.Error
            );
            diagnostic.source = 'robogo';
            this.diagnosticCollection.set(document.uri, [diagnostic]);
        }
    }

    /**
     * Validate custom Robogo-specific rules
     */
    private validateCustomRules(document: vscode.TextDocument): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];
        const text = document.getText();
        const lines = text.split('\n');

        // Validate step naming conventions
        diagnostics.push(...this.validateStepNaming(lines));
        
        // Validate variable usage
        diagnostics.push(...this.validateVariableUsage(lines));
        
        // Validate action arguments
        diagnostics.push(...this.validateActionArguments(lines));
        
        // Validate secrets configuration
        diagnostics.push(...this.validateSecretsConfiguration(lines));
        
        // Validate template usage
        diagnostics.push(...this.validateTemplateUsage(lines));
        
        // Validate parallel configuration
        diagnostics.push(...this.validateParallelConfiguration(lines));
        
        // Performance recommendations
        diagnostics.push(...this.generatePerformanceRecommendations(lines));

        return diagnostics;
    }

    /**
     * Validate step naming conventions
     */
    private validateStepNaming(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed.startsWith('- name:')) {
                const nameMatch = trimmed.match(/- name:\s*["']?([^"']+)["']?/);
                if (nameMatch) {
                    const stepName = nameMatch[1];
                    
                    // Check for empty or generic names
                    if (!stepName || stepName.trim() === '') {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, 0, i, line.length),
                            'Step name cannot be empty',
                            vscode.DiagnosticSeverity.Error
                        );
                        diagnostic.source = 'robogo';
                        diagnostics.push(diagnostic);
                    } else if (['step', 'test', 'action'].includes(stepName.toLowerCase())) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(stepName), i, line.indexOf(stepName) + stepName.length),
                            'Use descriptive step names instead of generic terms',
                            vscode.DiagnosticSeverity.Warning
                        );
                        diagnostic.source = 'robogo';
                        diagnostic.code = 'naming-convention';
                        diagnostics.push(diagnostic);
                    }
                    
                    // Check for overly long names
                    if (stepName.length > 100) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(stepName), i, line.indexOf(stepName) + stepName.length),
                            'Step name is too long (max 100 characters recommended)',
                            vscode.DiagnosticSeverity.Information
                        );
                        diagnostic.source = 'robogo';
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }

        return diagnostics;
    }

    /**
     * Validate variable usage patterns
     */
    private validateVariableUsage(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];
        const definedVars = new Set<string>();
        const usedVars = new Set<string>();

        // First pass: collect defined variables
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const varMatch = line.match(/^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:/);
            if (varMatch) {
                definedVars.add(varMatch[1]);
            }
        }

        // Second pass: check variable usage
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            
            // Find variable references
            const varReferences = line.match(/\$\{([^}]+)\}/g);
            if (varReferences) {
                for (const ref of varReferences) {
                    const varName = ref.slice(2, -1);
                    
                    // Handle special cases
                    if (varName.startsWith('SECRETS.') || varName === '__robogo_steps' || varName.includes('[')) {
                        continue;
                    }
                    
                    // Handle dot notation
                    const rootVar = varName.split('.')[0];
                    usedVars.add(rootVar);
                    
                    // Check for typos in common variable names
                    if (this.isPossibleTypo(rootVar, definedVars)) {
                        const suggestions = this.getSuggestions(rootVar, definedVars);
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(ref), i, line.indexOf(ref) + ref.length),
                            `Possible typo in variable name. Did you mean: ${suggestions.join(', ')}?`,
                            vscode.DiagnosticSeverity.Warning
                        );
                        diagnostic.source = 'robogo';
                        diagnostic.code = 'possible-typo';
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }

        // Check for unused variables
        for (const varName of definedVars) {
            if (!usedVars.has(varName) && !varName.startsWith('_')) {
                for (let i = 0; i < lines.length; i++) {
                    const line = lines[i];
                    if (line.includes(`${varName}:`)) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(varName), i, line.indexOf(varName) + varName.length),
                            `Variable '${varName}' is defined but never used`,
                            vscode.DiagnosticSeverity.Information
                        );
                        diagnostic.source = 'robogo';
                        diagnostic.code = 'unused-variable';
                        diagnostic.tags = [vscode.DiagnosticTag.Unnecessary];
                        diagnostics.push(diagnostic);
                        break;
                    }
                }
            }
        }

        return diagnostics;
    }

    /**
     * Validate action arguments
     */
    private validateActionArguments(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed.startsWith('action:')) {
                const actionMatch = trimmed.match(/action:\s*["']?([^"'\s]+)["']?/);
                if (actionMatch) {
                    const actionName = actionMatch[1];
                    const action = this.languageServer.getActionRegistry().getAction(actionName);
                    
                    if (action) {
                        // Check if next line has args
                        const nextLineIndex = i + 1;
                        if (nextLineIndex < lines.length) {
                            const nextLine = lines[nextLineIndex];
                            
                            // Validate argument count for specific actions
                            if (nextLine.trim().startsWith('args:')) {
                                const argsValidation = this.validateActionArgumentCount(action, nextLine, nextLineIndex);
                                if (argsValidation) {
                                    diagnostics.push(argsValidation);
                                }
                            } else if (action.parameters.some(p => p.required)) {
                                // Missing required arguments
                                const diagnostic = new vscode.Diagnostic(
                                    new vscode.Range(i, 0, i, line.length),
                                    `Action '${actionName}' requires arguments but none provided`,
                                    vscode.DiagnosticSeverity.Error
                                );
                                diagnostic.source = 'robogo';
                                diagnostics.push(diagnostic);
                            }
                        }
                    }
                }
            }
        }

        return diagnostics;
    }

    /**
     * Validate secrets configuration
     */
    private validateSecretsConfiguration(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Check for secrets in wrong places
            if ((trimmed.includes('password') || trimmed.includes('token') || trimmed.includes('key')) && 
                !trimmed.includes('SECRETS.') && 
                !line.includes('secrets:')) {
                
                const diagnostic = new vscode.Diagnostic(
                    new vscode.Range(i, 0, i, line.length),
                    'Consider using SECRETS namespace for sensitive data',
                    vscode.DiagnosticSeverity.Information
                );
                diagnostic.source = 'robogo';
                diagnostic.code = 'use-secrets';
                diagnostics.push(diagnostic);
            }

            // Check for hardcoded secrets
            if (trimmed.includes('password:') && !trimmed.includes('${') && trimmed.includes('"')) {
                const diagnostic = new vscode.Diagnostic(
                    new vscode.Range(i, 0, i, line.length),
                    'Avoid hardcoding passwords. Use variables or secrets instead',
                    vscode.DiagnosticSeverity.Warning
                );
                diagnostic.source = 'robogo';
                diagnostic.code = 'hardcoded-secret';
                diagnostics.push(diagnostic);
            }
        }

        return diagnostics;
    }

    /**
     * Validate template usage
     */
    private validateTemplateUsage(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            
            if (line.includes('template') && line.includes('.tmpl')) {
                const templateMatch = line.match(/["']([^"']*\.tmpl)["']/);
                if (templateMatch) {
                    const templatePath = templateMatch[1];
                    
                    // Check for common template files
                    const commonTemplates = ['mt103.tmpl', 'mt202.tmpl', 'sepa-credit-transfer.xml.tmpl'];
                    const templateName = templatePath.split('/').pop() || '';
                    
                    if (commonTemplates.includes(templateName)) {
                        // Suggest checking required fields
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(templatePath), i, line.indexOf(templatePath) + templatePath.length),
                            `Template '${templateName}' requires specific fields. Hover for details.`,
                            vscode.DiagnosticSeverity.Information
                        );
                        diagnostic.source = 'robogo';
                        diagnostic.code = 'template-requirements';
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }

        return diagnostics;
    }

    /**
     * Validate parallel configuration
     */
    private validateParallelConfiguration(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];
        
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            
            if (line.includes('max_concurrency:')) {
                const concurrencyMatch = line.match(/max_concurrency:\s*(\d+)/);
                if (concurrencyMatch) {
                    const concurrency = parseInt(concurrencyMatch[1]);
                    
                    if (concurrency > 16) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(concurrencyMatch[1]), i, line.indexOf(concurrencyMatch[1]) + concurrencyMatch[1].length),
                            'High concurrency values may cause resource exhaustion',
                            vscode.DiagnosticSeverity.Warning
                        );
                        diagnostic.source = 'robogo';
                        diagnostic.code = 'high-concurrency';
                        diagnostics.push(diagnostic);
                    } else if (concurrency < 1) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(i, line.indexOf(concurrencyMatch[1]), i, line.indexOf(concurrencyMatch[1]) + concurrencyMatch[1].length),
                            'Concurrency must be at least 1',
                            vscode.DiagnosticSeverity.Error
                        );
                        diagnostic.source = 'robogo';
                        diagnostics.push(diagnostic);
                    }
                }
            }
        }

        return diagnostics;
    }

    /**
     * Generate performance recommendations
     */
    private generatePerformanceRecommendations(lines: string[]): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];
        let stepCount = 0;
        let hasParallelConfig = false;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed.startsWith('- name:')) {
                stepCount++;
            }

            if (trimmed.includes('parallel:') || trimmed.includes('max_concurrency:')) {
                hasParallelConfig = true;
            }
        }

        // Suggest parallel execution for tests with many steps
        if (stepCount > 10 && !hasParallelConfig) {
            const diagnostic = new vscode.Diagnostic(
                new vscode.Range(0, 0, 0, 0),
                `Test has ${stepCount} steps. Consider enabling parallel execution for better performance`,
                vscode.DiagnosticSeverity.Information
            );
            diagnostic.source = 'robogo';
            diagnostic.code = 'performance-parallel';
            diagnostics.push(diagnostic);
        }

        return diagnostics;
    }

    /**
     * Check if a variable name is a possible typo
     */
    private isPossibleTypo(varName: string, definedVars: Set<string>): boolean {
        for (const defined of definedVars) {
            if (this.levenshteinDistance(varName, defined) === 1 && varName.length > 2) {
                return true;
            }
        }
        return false;
    }

    /**
     * Get suggestions for possible typos
     */
    private getSuggestions(varName: string, definedVars: Set<string>): string[] {
        const suggestions: string[] = [];
        for (const defined of definedVars) {
            if (this.levenshteinDistance(varName, defined) <= 2) {
                suggestions.push(defined);
            }
        }
        return suggestions.slice(0, 3); // Limit to 3 suggestions
    }

    /**
     * Calculate Levenshtein distance for typo detection
     */
    private levenshteinDistance(str1: string, str2: string): number {
        const matrix: number[][] = [];

        for (let i = 0; i <= str2.length; i++) {
            matrix[i] = [i];
        }

        for (let j = 0; j <= str1.length; j++) {
            matrix[0][j] = j;
        }

        for (let i = 1; i <= str2.length; i++) {
            for (let j = 1; j <= str1.length; j++) {
                if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
                    matrix[i][j] = matrix[i - 1][j - 1];
                } else {
                    matrix[i][j] = Math.min(
                        matrix[i - 1][j - 1] + 1,
                        matrix[i][j - 1] + 1,
                        matrix[i - 1][j] + 1
                    );
                }
            }
        }

        return matrix[str2.length][str1.length];
    }

    /**
     * Validate action argument count
     */
    private validateActionArgumentCount(action: any, argsLine: string, lineIndex: number): vscode.Diagnostic | null {
        // Extract argument count from line
        const argsMatch = argsLine.match(/args:\s*\[(.*?)\]/);
        if (argsMatch) {
            const argsContent = argsMatch[1];
            const argCount = argsContent.split(',').filter(arg => arg.trim() !== '').length;
            
            const requiredParams = action.parameters.filter((p: any) => p.required);
            
            if (argCount < requiredParams.length) {
                return new vscode.Diagnostic(
                    new vscode.Range(lineIndex, 0, lineIndex, argsLine.length),
                    `Action '${action.name}' requires at least ${requiredParams.length} arguments, but ${argCount} provided`,
                    vscode.DiagnosticSeverity.Error
                );
            }
        }
        
        return null;
    }
}