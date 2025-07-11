import * as vscode from 'vscode';

/**
 * Manages VS Code configuration for Robogo extension
 */
export class ConfigurationManager {
    private readonly configSection = 'robogo';

    /**
     * Get Robogo executable path
     */
    getExecutablePath(): string {
        return this.getConfig<string>('executablePath', 'robogo');
    }

    /**
     * Get default output format
     */
    getOutputFormat(): 'console' | 'json' | 'markdown' {
        return this.getConfig<'console' | 'json' | 'markdown'>('outputFormat', 'console');
    }

    /**
     * Check if parallel execution is enabled
     */
    isParallelExecutionEnabled(): boolean {
        return this.getConfig<boolean>('enableParallelExecution', true);
    }

    /**
     * Get maximum concurrency for parallel execution
     */
    getMaxConcurrency(): number {
        return this.getConfig<number>('maxConcurrency', 4);
    }

    /**
     * Check if real-time validation is enabled
     */
    isRealTimeValidationEnabled(): boolean {
        return this.getConfig<boolean>('enableRealTimeValidation', true);
    }

    /**
     * Check if verbose output is enabled
     */
    isVerboseOutputEnabled(): boolean {
        return this.getConfig<boolean>('showVerboseOutput', false);
    }

    /**
     * Check if autocomplete is enabled
     */
    isAutoCompleteEnabled(): boolean {
        return this.getConfig<boolean>('enableAutoComplete', true);
    }

    /**
     * Check if hover documentation is enabled
     */
    isHoverDocumentationEnabled(): boolean {
        return this.getConfig<boolean>('enableHoverDocumentation', true);
    }

    /**
     * Check if detailed documentation should be shown
     */
    shouldShowDetailedDocumentation(): boolean {
        return this.getConfig<boolean>('showDetailedDocumentation', true);
    }

    /**
     * Check if syntax highlighting is enabled
     */
    isSyntaxHighlightingEnabled(): boolean {
        return this.getConfig<boolean>('enableSyntaxHighlighting', true);
    }

    /**
     * Check if code snippets are enabled
     */
    areCodeSnippetsEnabled(): boolean {
        return this.getConfig<boolean>('enableCodeSnippets', true);
    }

    /**
     * Get workspace folder path
     */
    getWorkspacePath(): string | undefined {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        return workspaceFolders?.[0]?.uri.fsPath;
    }

    /**
     * Update configuration value
     */
    async updateConfig<T>(key: string, value: T, scope: vscode.ConfigurationTarget = vscode.ConfigurationTarget.Workspace): Promise<void> {
        const config = vscode.workspace.getConfiguration(this.configSection);
        await config.update(key, value, scope);
    }

    /**
     * Get configuration value with fallback
     */
    private getConfig<T>(key: string, defaultValue: T): T {
        const config = vscode.workspace.getConfiguration(this.configSection);
        return config.get<T>(key, defaultValue);
    }

    /**
     * Get all configuration as object
     */
    getAllConfig(): any {
        return vscode.workspace.getConfiguration(this.configSection);
    }

    /**
     * Reset configuration to defaults
     */
    async resetToDefaults(): Promise<void> {
        const config = vscode.workspace.getConfiguration(this.configSection);
        const inspect = config.inspect('');
        
        if (inspect?.workspaceValue) {
            await config.update('', undefined, vscode.ConfigurationTarget.Workspace);
        }
        
        if (inspect?.globalValue) {
            await config.update('', undefined, vscode.ConfigurationTarget.Global);
        }
    }

    /**
     * Validate configuration
     */
    validateConfiguration(): { isValid: boolean; errors: string[] } {
        const errors: string[] = [];
        
        // Validate executable path
        const execPath = this.getExecutablePath();
        if (!execPath || execPath.trim() === '') {
            errors.push('Executable path cannot be empty');
        }

        // Validate max concurrency
        const maxConcurrency = this.getMaxConcurrency();
        if (maxConcurrency < 1 || maxConcurrency > 32) {
            errors.push('Max concurrency must be between 1 and 32');
        }

        // Validate output format
        const outputFormat = this.getOutputFormat();
        if (!['console', 'json', 'markdown'].includes(outputFormat)) {
            errors.push('Output format must be one of: console, json, markdown');
        }

        return {
            isValid: errors.length === 0,
            errors
        };
    }
}