package common

import (
	"fmt"
	"os"
	"strings"
)

// TemplateSubstitution handles variable substitution in templates
type TemplateSubstitution struct {
	storage VariableStorage
}

// NewTemplateSubstitution creates a new template substitution engine
func NewTemplateSubstitution(storage VariableStorage) *TemplateSubstitution {
	return &TemplateSubstitution{
		storage: storage,
	}
}

// SubstitutionResult contains the result of template substitution
type SubstitutionResult struct {
	// Result is the final substituted string
	Result string
	// UnresolvedVariables contains variables that could not be resolved
	UnresolvedVariables []string
	// HasUnresolved indicates if any variables were unresolved
	HasUnresolved bool
}

// Substitute performs variable substitution using ${variable} syntax
func (ts *TemplateSubstitution) Substitute(template string) SubstitutionResult {
	result := template
	var unresolved []string
	
	// Keep substituting until no more variables found or max iterations reached
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		newResult, newUnresolved := ts.substituteOnce(result)
		
		// If no changes made, we're done
		if newResult == result {
			break
		}
		
		result = newResult
		unresolved = append(unresolved[:0], newUnresolved...) // Reset and add new unresolved
	}
	
	return SubstitutionResult{
		Result:              result,
		UnresolvedVariables: unresolved,
		HasUnresolved:       len(unresolved) > 0,
	}
}

// substituteOnce performs a single pass of variable substitution
func (ts *TemplateSubstitution) substituteOnce(template string) (string, []string) {
	result := template
	var unresolved []string
	
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start
		
		// Extract variable name
		varName := result[start+2 : end]
		varPlaceholder := result[start : end+1]
		
		// Resolve variable
		value, resolved := ts.resolveVariable(varName)
		
		if resolved {
			// Replace with resolved value
			result = strings.Replace(result, varPlaceholder, fmt.Sprintf("%v", value), 1)
		} else {
			// Track unresolved variable and move past it
			unresolved = append(unresolved, varName)
			// Replace with a marker to avoid infinite loops
			result = strings.Replace(result, varPlaceholder, "{{UNRESOLVED:"+varName+"}}", 1)
		}
	}
	
	return result, unresolved
}

// resolveVariable attempts to resolve a single variable
func (ts *TemplateSubstitution) resolveVariable(varName string) (any, bool) {
	// Handle environment variables
	if strings.HasPrefix(varName, "ENV:") {
		envVarName := varName[4:]
		if envValue := os.Getenv(envVarName); envValue != "" {
			return envValue, true
		}
		return nil, false
	}
	
	// Handle regular variables
	if ts.storage.Has(varName) {
		return ts.storage.Get(varName), true
	}
	
	return nil, false
}

// CleanUnresolvedMarkers removes unresolved markers from final result
func (ts *TemplateSubstitution) CleanUnresolvedMarkers(text string) string {
	// Replace markers back to original variable syntax for display
	result := text
	for strings.Contains(result, "{{UNRESOLVED:") {
		start := strings.Index(result, "{{UNRESOLVED:")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2
		
		// Extract variable name from marker
		marker := result[start:end]
		varName := marker[13 : len(marker)-2] // Remove "{{UNRESOLVED:" and "}}"
		
		// Replace with original variable syntax
		result = strings.Replace(result, marker, "${"+varName+"}", 1)
	}
	return result
}