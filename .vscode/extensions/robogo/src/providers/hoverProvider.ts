import * as vscode from 'vscode';
import { RobogoLanguageServer } from '../core/languageServer';

/**
 * Provides hover documentation for Robogo actions and syntax
 */
export class HoverProvider implements vscode.HoverProvider {

    constructor(private languageServer: RobogoLanguageServer) {}

    /**
     * Provide hover information
     */
    async provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Hover | null> {
        const wordRange = document.getWordRangeAtPosition(position);
        if (!wordRange) return null;

        const word = document.getText(wordRange);
        const line = document.lineAt(position);
        const lineText = line.text;

        // Action documentation
        if (lineText.includes('action:') && lineText.includes(word)) {
            return this.getActionHover(word, wordRange);
        }

        // Variable documentation
        if (lineText.includes('${' + word + '}') || lineText.includes('SECRETS.' + word)) {
            return this.getVariableHover(document, word, wordRange);
        }

        // Field documentation
        if (lineText.includes(word + ':')) {
            return this.getFieldHover(word, wordRange);
        }

        // Template documentation
        if (lineText.includes('template') && lineText.includes(word)) {
            return this.getTemplateHover(word, wordRange);
        }

        // Special syntax documentation
        return this.getSpecialSyntaxHover(word, lineText, wordRange);
    }

    /**
     * Get hover documentation for actions
     */
    private getActionHover(actionName: string, range: vscode.Range): vscode.Hover | null {
        const action = this.languageServer.getActionRegistry().getAction(actionName);
        if (!action) return null;

        const markdown = new vscode.MarkdownString();
        markdown.isTrusted = true;
        markdown.supportHtml = true;

        // Action header
        markdown.appendMarkdown(`### 🎯 **${action.name}** Action\n\n`);
        markdown.appendMarkdown(`*${action.description}*\n\n`);

        // Parameters
        if (action.parameters.length > 0) {
            markdown.appendMarkdown(`**Parameters:**\n\n`);
            for (const param of action.parameters) {
                const required = param.required ? '*(required)*' : '*(optional)*';
                const defaultValue = param.default !== undefined ? ` - Default: \`${param.default}\`` : '';
                markdown.appendMarkdown(`- **${param.name}** (${param.type}) ${required}: ${param.description}${defaultValue}\n`);
            }
            markdown.appendMarkdown(`\n`);
        }

        // Examples
        if (action.examples.length > 0) {
            markdown.appendMarkdown(`**Example:**\n\n`);
            markdown.appendCodeblock(action.examples[0], 'yaml');
        }

        // Category
        markdown.appendMarkdown(`\n**Category:** ${action.category}`);

        return new vscode.Hover(markdown, range);
    }

    /**
     * Get hover documentation for variables
     */
    private getVariableHover(document: vscode.TextDocument, variableName: string, range: vscode.Range): vscode.Hover | null {
        const isSecret = variableName.startsWith('SECRETS.');
        const actualVarName = isSecret ? variableName.replace('SECRETS.', '') : variableName;

        const markdown = new vscode.MarkdownString();
        markdown.isTrusted = true;

        if (isSecret) {
            markdown.appendMarkdown(`### 🔒 **Secret Variable: ${actualVarName}**\n\n`);
            markdown.appendMarkdown(`*Secure variable with output masking*\n\n`);
            markdown.appendMarkdown(`**Usage:**\n`);
            markdown.appendMarkdown(`- Access with: \`\${SECRETS.${actualVarName}}\`\n`);
            markdown.appendMarkdown(`- Automatically masked in output logs\n`);
            markdown.appendMarkdown(`- Can be loaded from files or defined inline\n\n`);
            markdown.appendMarkdown(`**Security:** ⚠️ Values are masked in output but stored in memory during execution.`);
        } else {
            markdown.appendMarkdown(`### 📦 **Variable: ${variableName}**\n\n`);
            
            // Try to find variable definition and value
            const varInfo = this.findVariableInfo(document, variableName);
            if (varInfo) {
                markdown.appendMarkdown(`**Defined at:** Line ${varInfo.line + 1}\n\n`);
                if (varInfo.value) {
                    markdown.appendMarkdown(`**Value:** \`${varInfo.value}\`\n\n`);
                }
                if (varInfo.type) {
                    markdown.appendMarkdown(`**Type:** ${varInfo.type}\n\n`);
                }
            }

            markdown.appendMarkdown(`**Usage:**\n`);
            markdown.appendMarkdown(`- Access with: \`\${${variableName}}\`\n`);
            markdown.appendMarkdown(`- Supports dot notation: \`\${${variableName}.property}\`\n`);
            markdown.appendMarkdown(`- Can be used in any string value\n`);
        }

        return new vscode.Hover(markdown, range);
    }

    /**
     * Get hover documentation for fields
     */
    private getFieldHover(fieldName: string, range: vscode.Range): vscode.Hover | null {
        const fieldDocs: { [key: string]: { description: string; type: string; examples?: string[] } } = {
            'testcase': {
                description: 'Name of the test case',
                type: 'string',
                examples: ['"API Login Test"', '"Database Connection Test"']
            },
            'testsuite': {
                description: 'Name of the test suite',
                type: 'string',
                examples: ['"User Management Suite"', '"Integration Test Suite"']
            },
            'description': {
                description: 'Description of the test case or suite',
                type: 'string',
                examples: ['"Test user authentication flow"', '"Validate database operations"']
            },
            'variables': {
                description: 'Variable definitions for the test',
                type: 'object',
                examples: ['vars:', 'secrets:']
            },
            'steps': {
                description: 'List of test steps to execute',
                type: 'array',
                examples: ['- name: "Step 1"']
            },
            'name': {
                description: 'Name of the test step',
                type: 'string',
                examples: ['"HTTP GET request"', '"Validate response"']
            },
            'action': {
                description: 'Action to execute in this step',
                type: 'string',
                examples: ['http', 'assert', 'postgres', 'log']
            },
            'args': {
                description: 'Arguments passed to the action',
                type: 'array',
                examples: ['["GET", "https://api.example.com"]', '["assert", "${response.status}", "==", "200"]']
            },
            'result': {
                description: 'Variable name to store the action result',
                type: 'string',
                examples: ['response', 'user_data', 'query_result']
            },
            'if': {
                description: 'Conditional execution expression',
                type: 'string',
                examples: ['"${response.status} == 200"', '"${user.active} == true"']
            },
            'retry': {
                description: 'Retry configuration for the step',
                type: 'object',
                examples: ['attempts: 3', 'delay: "1s"']
            },
            'timeout': {
                description: 'Timeout for step execution',
                type: 'string',
                examples: ['"30s"', '"5m"', '"1h"']
            },
            'parallel': {
                description: 'Parallel execution configuration',
                type: 'object',
                examples: ['enabled: true', 'max_concurrency: 4']
            },
            'templates': {
                description: 'Template definitions for the test',
                type: 'object',
                examples: ['greeting: "Hello ${name}"']
            },
            'setup': {
                description: 'Steps to run before test cases in a suite',
                type: 'array'
            },
            'teardown': {
                description: 'Steps to run after test cases in a suite',
                type: 'array'
            },
            'testcases': {
                description: 'List of test case files in a suite',
                type: 'array',
                examples: ['- test1.robogo', '- test2.robogo']
            }
        };

        const fieldDoc = fieldDocs[fieldName];
        if (!fieldDoc) return null;

        const markdown = new vscode.MarkdownString();
        markdown.isTrusted = true;

        markdown.appendMarkdown(`### 📋 **${fieldName}**\n\n`);
        markdown.appendMarkdown(`**Description:** ${fieldDoc.description}\n\n`);
        markdown.appendMarkdown(`**Type:** \`${fieldDoc.type}\`\n\n`);

        if (fieldDoc.examples && fieldDoc.examples.length > 0) {
            markdown.appendMarkdown(`**Examples:**\n\n`);
            for (const example of fieldDoc.examples) {
                markdown.appendMarkdown(`- \`${example}\`\n`);
            }
        }

        return new vscode.Hover(markdown, range);
    }

    /**
     * Get hover documentation for templates
     */
    private getTemplateHover(templateName: string, range: vscode.Range): vscode.Hover | null {
        const templateDocs: { [key: string]: { description: string; fields: string[] } } = {
            'mt103.tmpl': {
                description: 'SWIFT MT103 Customer Transfer template',
                fields: ['TransactionID', 'Timestamp', 'Date', 'Currency', 'Amount', 'Reference', 'Sender.BIC', 'Sender.Account', 'Sender.Name', 'Beneficiary.Account', 'Beneficiary.Name']
            },
            'mt202.tmpl': {
                description: 'SWIFT MT202 Financial Institution Transfer template',
                fields: ['TransactionID', 'Date', 'Currency', 'Amount', 'Reference', 'Sender.BIC', 'Receiver.BIC']
            },
            'sepa-credit-transfer.xml.tmpl': {
                description: 'SEPA Credit Transfer XML template',
                fields: ['MessageID', 'CreationDateTime', 'InitiatingParty', 'Amount', 'Currency', 'Debtor', 'Creditor']
            }
        };

        const templateDoc = templateDocs[templateName];
        if (!templateDoc) {
            // Generic template documentation
            const markdown = new vscode.MarkdownString();
            markdown.appendMarkdown(`### 📄 **Template: ${templateName}**\n\n`);
            markdown.appendMarkdown(`Template file for generating formatted output.\n\n`);
            markdown.appendMarkdown(`**Usage:**\n`);
            markdown.appendCodeblock(`action: template\nargs: ["${templateName}", {"field": "value"}]`, 'yaml');
            return new vscode.Hover(markdown, range);
        }

        const markdown = new vscode.MarkdownString();
        markdown.isTrusted = true;

        markdown.appendMarkdown(`### 📄 **${templateDoc.description}**\n\n`);
        markdown.appendMarkdown(`**Required Fields:**\n\n`);
        
        for (const field of templateDoc.fields) {
            markdown.appendMarkdown(`- \`${field}\`\n`);
        }

        markdown.appendMarkdown(`\n**Usage Example:**\n\n`);
        markdown.appendCodeblock(`action: template\nargs: ["${templateName}", {\n  "${templateDoc.fields[0]}": "value1",\n  "${templateDoc.fields[1]}": "value2"\n}]\nresult: generated_message`, 'yaml');

        return new vscode.Hover(markdown, range);
    }

    /**
     * Get hover documentation for special syntax
     */
    private getSpecialSyntaxHover(word: string, lineText: string, range: vscode.Range): vscode.Hover | null {
        const specialSyntax: { [key: string]: { description: string; usage: string } } = {
            '__robogo_steps': {
                description: 'Special variable containing execution history of all previous steps',
                usage: '${__robogo_steps[0].result} - Access result of first step\n${__robogo_steps[0].error} - Access error of first step'
            },
            'SECRETS': {
                description: 'Namespace for secure variables with output masking',
                usage: '${SECRETS.api_key} - Access secret variable\nAutomatically masked in logs and output'
            }
        };

        const syntax = specialSyntax[word];
        if (!syntax) return null;

        const markdown = new vscode.MarkdownString();
        markdown.isTrusted = true;

        markdown.appendMarkdown(`### ⚡ **Special Syntax: ${word}**\n\n`);
        markdown.appendMarkdown(`${syntax.description}\n\n`);
        markdown.appendMarkdown(`**Usage:**\n\n`);
        markdown.appendCodeblock(syntax.usage, 'yaml');

        return new vscode.Hover(markdown, range);
    }

    /**
     * Find variable information in document
     */
    private findVariableInfo(document: vscode.TextDocument, variableName: string): { line: number; value?: string; type?: string } | null {
        const text = document.getText();
        const lines = text.split('\n');

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const regex = new RegExp(`^\\s*${variableName}\\s*:\\s*(.+)?`);
            const match = regex.exec(line);
            
            if (match) {
                let value = match[1]?.trim();
                let type = 'string';

                if (value) {
                    // Remove quotes
                    value = value.replace(/^["']|["']$/g, '');
                    
                    // Determine type
                    if (value.startsWith('{') || value.startsWith('[')) {
                        type = value.startsWith('{') ? 'object' : 'array';
                    } else if (!isNaN(Number(value))) {
                        type = 'number';
                    } else if (value === 'true' || value === 'false') {
                        type = 'boolean';
                    }
                }

                return { line: i, value, type };
            }
        }

        return null;
    }
}