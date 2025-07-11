import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';
import { ConfigurationManager } from '../core/configurationManager';

const execAsync = promisify(exec);

/**
 * Handles test execution with various output formats and options
 */
export class TestExecutor {
    constructor(private config: ConfigurationManager) {}

    /**
     * Run a single test file
     */
    async runTest(uri: vscode.Uri, options: TestExecutionOptions): Promise<void> {
        const filePath = uri.fsPath;
        await this.executeTest(filePath, options, 'run');
    }

    /**
     * Run a test suite
     */
    async runTestSuite(uri: vscode.Uri, options: TestExecutionOptions): Promise<void> {
        const filePath = uri.fsPath;
        await this.executeTest(filePath, options, 'run-suite');
    }

    /**
     * Debug a test with variable inspection
     */
    async debugTest(uri: vscode.Uri): Promise<void> {
        const filePath = uri.fsPath;
        await this.executeTest(filePath, {
            outputFormat: 'console',
            verbose: true,
            debug: true
        }, 'run');
    }

    /**
     * Run a specific step (simulated by running test with step context)
     */
    async runStep(uri: vscode.Uri, line: number): Promise<void> {
        vscode.window.showInformationMessage(`Running step at line ${line + 1}...`);
        
        // For now, run the entire test with verbose output
        // In a full implementation, this would extract and run just the specific step
        await this.runTest(uri, {
            outputFormat: 'console',
            verbose: true
        });
    }

    /**
     * Execute test command
     */
    private async executeTest(filePath: string, options: TestExecutionOptions, command: 'run' | 'run-suite'): Promise<void> {
        try {
            // Show progress
            const progressOptions = {
                location: vscode.ProgressLocation.Notification,
                title: `Running ${command === 'run-suite' ? 'test suite' : 'test'}...`,
                cancellable: true
            };

            await vscode.window.withProgress(progressOptions, async (progress, token) => {
                // Build command
                const cmd = this.buildCommand(filePath, options, command);
                
                progress.report({ message: 'Executing test...', increment: 20 });

                // Execute command
                const { stdout, stderr } = await execAsync(cmd, { 
                    cwd: this.config.getWorkspacePath(),
                    timeout: 300000 // 5 minutes timeout
                });

                if (token.isCancellationRequested) {
                    vscode.window.showWarningMessage('Test execution was cancelled.');
                    return;
                }

                progress.report({ message: 'Processing results...', increment: 80 });

                // Display results based on output format
                await this.displayResults(stdout, stderr, options.outputFormat);

                progress.report({ message: 'Complete', increment: 100 });
            });

        } catch (error: any) {
            const errorMessage = error.message || 'Unknown error occurred';
            
            if (errorMessage.includes('ENOENT')) {
                vscode.window.showErrorMessage(
                    'Robogo executable not found. Please check the executable path in settings.',
                    'Open Settings'
                ).then(selection => {
                    if (selection === 'Open Settings') {
                        vscode.commands.executeCommand('robogo.openSettings');
                    }
                });
            } else if (errorMessage.includes('timeout')) {
                vscode.window.showErrorMessage('Test execution timed out. Consider optimizing your test or increasing timeout.');
            } else {
                vscode.window.showErrorMessage(`Test execution failed: ${errorMessage}`);
                
                // Show detailed error in output channel
                const outputChannel = vscode.window.createOutputChannel('Robogo Test Execution');
                outputChannel.appendLine('=== Test Execution Error ===');
                outputChannel.appendLine(`Command: ${this.buildCommand(filePath, options, command)}`);
                outputChannel.appendLine(`Error: ${errorMessage}`);
                outputChannel.appendLine(`Stderr: ${error.stderr || 'None'}`);
                outputChannel.show();
            }
        }
    }

    /**
     * Build command for test execution
     */
    private buildCommand(filePath: string, options: TestExecutionOptions, command: 'run' | 'run-suite'): string {
        const executable = this.config.getExecutablePath();
        let cmd = `"${executable}" ${command} "${filePath}"`;

        // Add output format
        if (options.outputFormat) {
            cmd += ` --output ${options.outputFormat}`;
        }

        // Add parallel options
        if (options.parallel) {
            cmd += ' --parallel';
            if (options.maxConcurrency) {
                cmd += ` --max-concurrency ${options.maxConcurrency}`;
            }
        }

        // Add verbose flag
        if (options.verbose) {
            cmd += ' --verbose';
        }

        // Add debug flag
        if (options.debug) {
            cmd += ' --debug';
        }

        return cmd;
    }

    /**
     * Display test results based on format
     */
    private async displayResults(stdout: string, stderr: string, format?: 'console' | 'json' | 'markdown'): Promise<void> {
        if (stderr && stderr.trim() !== '') {
            vscode.window.showErrorMessage('Test execution completed with warnings. Check output for details.');
        }

        switch (format) {
            case 'json':
                await this.displayJSONResults(stdout);
                break;
            case 'markdown':
                await this.displayMarkdownResults(stdout);
                break;
            default:
                await this.displayConsoleResults(stdout, stderr);
                break;
        }
    }

    /**
     * Display console output results
     */
    private async displayConsoleResults(stdout: string, stderr: string): Promise<void> {
        const outputChannel = vscode.window.createOutputChannel('Robogo Test Results');
        outputChannel.clear();
        
        outputChannel.appendLine('=== ROBOGO TEST EXECUTION RESULTS ===\n');
        
        if (stdout) {
            outputChannel.appendLine('--- Output ---');
            outputChannel.appendLine(stdout);
        }
        
        if (stderr && stderr.trim() !== '') {
            outputChannel.appendLine('\n--- Warnings/Errors ---');
            outputChannel.appendLine(stderr);
        }
        
        outputChannel.show(true);

        // Parse results for summary
        const summary = this.parseTestSummary(stdout);
        if (summary) {
            if (summary.failed > 0) {
                vscode.window.showErrorMessage(`❌ Tests failed: ${summary.failed}/${summary.total} tests failed`);
            } else {
                vscode.window.showInformationMessage(`✅ All tests passed: ${summary.passed}/${summary.total} tests successful`);
            }
        }
    }

    /**
     * Display JSON formatted results
     */
    private async displayJSONResults(stdout: string): Promise<void> {
        try {
            const results = JSON.parse(stdout);
            
            // Create a new document with formatted JSON
            const doc = await vscode.workspace.openTextDocument({
                content: JSON.stringify(results, null, 2),
                language: 'json'
            });
            
            await vscode.window.showTextDocument(doc, { viewColumn: vscode.ViewColumn.Beside });
            
            // Show summary
            if (results.summary) {
                const { passed, failed, total } = results.summary;
                if (failed > 0) {
                    vscode.window.showErrorMessage(`❌ Tests failed: ${failed}/${total} tests failed`);
                } else {
                    vscode.window.showInformationMessage(`✅ All tests passed: ${passed}/${total} tests successful`);
                }
            }
            
        } catch (error) {
            vscode.window.showErrorMessage('Failed to parse JSON results');
            await this.displayConsoleResults(stdout, '');
        }
    }

    /**
     * Display markdown formatted results
     */
    private async displayMarkdownResults(stdout: string): Promise<void> {
        // Create a new markdown document
        const doc = await vscode.workspace.openTextDocument({
            content: stdout,
            language: 'markdown'
        });
        
        await vscode.window.showTextDocument(doc, { viewColumn: vscode.ViewColumn.Beside });
        
        // Parse for summary
        const summary = this.parseTestSummary(stdout);
        if (summary) {
            if (summary.failed > 0) {
                vscode.window.showErrorMessage(`❌ Tests failed: ${summary.failed}/${summary.total} tests failed`);
            } else {
                vscode.window.showInformationMessage(`✅ All tests passed: ${summary.passed}/${summary.total} tests successful`);
            }
        }
    }

    /**
     * Parse test summary from output
     */
    private parseTestSummary(output: string): { passed: number; failed: number; total: number } | null {
        // Try to parse common result patterns
        const patterns = [
            /(\d+)\s+passed.*?(\d+)\s+failed.*?(\d+)\s+total/i,
            /passed:\s*(\d+).*?failed:\s*(\d+).*?total:\s*(\d+)/i,
            /✅.*?(\d+).*?❌.*?(\d+).*?total.*?(\d+)/i
        ];

        for (const pattern of patterns) {
            const match = output.match(pattern);
            if (match) {
                return {
                    passed: parseInt(match[1]),
                    failed: parseInt(match[2]),
                    total: parseInt(match[3])
                };
            }
        }

        // Fallback: count success/failure indicators
        const successCount = (output.match(/✅|PASSED|Success:/gi) || []).length;
        const failureCount = (output.match(/❌|FAILED|Error:|Failed:/gi) || []).length;
        
        if (successCount > 0 || failureCount > 0) {
            return {
                passed: successCount,
                failed: failureCount,
                total: successCount + failureCount
            };
        }

        return null;
    }
}

/**
 * Test execution options
 */
export interface TestExecutionOptions {
    outputFormat?: 'console' | 'json' | 'markdown';
    parallel?: boolean;
    maxConcurrency?: number;
    verbose?: boolean;
    debug?: boolean;
}