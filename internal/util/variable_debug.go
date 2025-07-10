package util

import (
	"fmt"
	"regexp"
	"strings"
)

// SecretManagerInterface defines methods for secret management to avoid import cycles
type SecretManagerInterface interface {
	MaskSensitiveOutput(output string) string
	IsSecretMasked(secretName string) bool
	GetSecretInfo(secretName string) (source string, masked bool, exists bool)
	ListSecrets() []string
}

// VariableResolutionDebugger helps debug variable substitution issues
type VariableResolutionDebugger struct {
	enabled bool
	context string
	secretManager SecretManagerInterface
}

// NewVariableResolutionDebugger creates a new variable resolution debugger
func NewVariableResolutionDebugger(enabled bool, context string, secretManager SecretManagerInterface) *VariableResolutionDebugger {
	return &VariableResolutionDebugger{
		enabled: enabled,
		context: context,
		secretManager: secretManager,
	}
}

// VariableResolutionResult contains the result of variable resolution analysis
type VariableResolutionResult struct {
	Original           string                    `json:"original"`
	Resolved           string                    `json:"resolved"`
	UnresolvedVars     []string                  `json:"unresolved_vars"`
	ResolvedVars       map[string]ResolvedVarInfo `json:"resolved_vars"`
	HasUnresolved      bool                      `json:"has_unresolved"`
	ResolutionWarnings []string                  `json:"resolution_warnings"`
}

// ResolvedVarInfo contains information about a resolved variable including security context
type ResolvedVarInfo struct {
	Value      string `json:"value"`
	IsSecret   bool   `json:"is_secret"`
	IsMasked   bool   `json:"is_masked"`
	Source     string `json:"source"`
	DisplayValue string `json:"display_value"`
}

// AnalyzeVariableResolution analyzes a string for variable resolution issues
func (vrd *VariableResolutionDebugger) AnalyzeVariableResolution(original, resolved string, availableVars map[string]interface{}) *VariableResolutionResult {
	result := &VariableResolutionResult{
		Original:           original,
		Resolved:           resolved,
		UnresolvedVars:     make([]string, 0),
		ResolvedVars:       make(map[string]ResolvedVarInfo, 0),
		ResolutionWarnings: make([]string, 0),
	}

	// Find all variable patterns in the original string
	varPattern := regexp.MustCompile(`\$\{([^}]+)\}`)
	originalMatches := varPattern.FindAllStringSubmatch(original, -1)
	resolvedMatches := varPattern.FindAllStringSubmatch(resolved, -1)

	// Track which variables were found in original
	originalVars := make(map[string]bool)
	for _, match := range originalMatches {
		if len(match) > 1 {
			originalVars[match[1]] = true
		}
	}

	// Track which variables are still unresolved
	unresolvedVars := make(map[string]bool)
	for _, match := range resolvedMatches {
		if len(match) > 1 {
			unresolvedVars[match[1]] = true
		}
	}

	// Identify resolved vs unresolved variables
	for varName := range originalVars {
		if unresolvedVars[varName] {
			// Variable was not resolved
			result.UnresolvedVars = append(result.UnresolvedVars, varName)
			result.HasUnresolved = true
			
			// Check if variable exists in available vars
			if _, exists := availableVars[varName]; !exists {
				result.ResolutionWarnings = append(result.ResolutionWarnings, 
					fmt.Sprintf("Variable '%s' is not defined", varName))
			} else {
				result.ResolutionWarnings = append(result.ResolutionWarnings, 
					fmt.Sprintf("Variable '%s' exists but was not substituted", varName))
			}
		} else {
			// Variable was resolved
			if value, exists := availableVars[varName]; exists {
				varInfo := vrd.createResolvedVarInfo(varName, value)
				result.ResolvedVars[varName] = varInfo
			}
		}
	}

	return result
}

// LogVariableResolution logs variable resolution analysis if debugging is enabled
func (vrd *VariableResolutionDebugger) LogVariableResolution(result *VariableResolutionResult) {
	if !vrd.enabled {
		return
	}
	
	// Always log if debugging is enabled and there are variables to resolve
	if !result.HasUnresolved && len(result.ResolvedVars) == 0 && len(result.UnresolvedVars) == 0 {
		return
	}

	fmt.Printf("🔍 Variable Resolution Debug (%s):\n", vrd.context)
	
	if len(result.ResolvedVars) > 0 {
		fmt.Printf("   ✅ Resolved Variables:\n")
		for varName, varInfo := range result.ResolvedVars {
			displayValue := truncateValue(varInfo.DisplayValue, 50)
			if varInfo.IsSecret {
				if varInfo.IsMasked {
					fmt.Printf("      ${%s} → %s (secret from %s, masked=true)\n", varName, displayValue, varInfo.Source)
				} else {
					fmt.Printf("      ${%s} → %s (secret from %s, masked=false)\n", varName, displayValue, varInfo.Source)
				}
			} else {
				fmt.Printf("      ${%s} → %s\n", varName, displayValue)
			}
		}
	}
	
	if result.HasUnresolved {
		fmt.Printf("   ❌ Unresolved Variables:\n")
		for _, varName := range result.UnresolvedVars {
			fmt.Printf("      ${%s} (not substituted)\n", varName)
		}
	}
	
	if len(result.ResolutionWarnings) > 0 {
		fmt.Printf("   ⚠️  Warnings:\n")
		for _, warning := range result.ResolutionWarnings {
			fmt.Printf("      %s\n", warning)
		}
	}
	
	fmt.Println()
}

// LogVariableSubstitution logs before/after substitution for debugging
func (vrd *VariableResolutionDebugger) LogVariableSubstitution(original, resolved string, availableVars map[string]interface{}) {
	if !vrd.enabled {
		return
	}

	result := vrd.AnalyzeVariableResolution(original, resolved, availableVars)
	vrd.LogVariableResolution(result)
}

// CreateVariableResolutionSummary creates a summary of variable resolution issues
func (vrd *VariableResolutionDebugger) CreateVariableResolutionSummary(stepResults []interface{}) string {
	if !vrd.enabled {
		return ""
	}

	var summary strings.Builder
	totalUnresolved := 0
	commonIssues := make(map[string]int)

	// This would be called with actual step results containing variable resolution data
	// For now, return a placeholder
	if totalUnresolved > 0 {
		summary.WriteString(fmt.Sprintf("🔍 Variable Resolution Summary: %d unresolved variables found\n", totalUnresolved))
		summary.WriteString("   Common Issues:\n")
		for issue, count := range commonIssues {
			summary.WriteString(fmt.Sprintf("      %s (%d occurrences)\n", issue, count))
		}
	}

	return summary.String()
}

// truncateValue truncates a string value for display
func truncateValue(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}

// GetVariablePattern returns the regex pattern used for variable detection
func GetVariablePattern() *regexp.Regexp {
	return regexp.MustCompile(`\$\{([^}]+)\}`)
}

// ExtractVariableNames extracts all variable names from a string
func ExtractVariableNames(text string) []string {
	pattern := GetVariablePattern()
	matches := pattern.FindAllStringSubmatch(text, -1)
	
	vars := make([]string, 0, len(matches))
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			vars = append(vars, match[1])
			seen[match[1]] = true
		}
	}
	
	return vars
}

// ValidateVariableAvailability checks if all required variables are available
func ValidateVariableAvailability(text string, availableVars map[string]interface{}) []string {
	requiredVars := ExtractVariableNames(text)
	missing := make([]string, 0)
	
	for _, varName := range requiredVars {
		if _, exists := availableVars[varName]; !exists {
			missing = append(missing, varName)
		}
	}
	
	return missing
}

// FormatVariableDebugInfo formats variable debug information for display
func FormatVariableDebugInfo(availableVars map[string]interface{}, requiredVars []string, secretManager SecretManagerInterface) string {
	var info strings.Builder
	
	info.WriteString("📋 Variable Information:\n")
	
	if len(availableVars) > 0 {
		info.WriteString("   Available Variables:\n")
		for name, value := range availableVars {
			displayValue := fmt.Sprintf("%v", value)
			if secretManager != nil {
				if source, masked, exists := secretManager.GetSecretInfo(name); exists {
					if masked {
						displayValue = "[MASKED]"
					}
					info.WriteString(fmt.Sprintf("      %s = %s (secret from %s, masked=%t)\n", name, truncateValue(displayValue, 40), source, masked))
					continue
				}
			}
			info.WriteString(fmt.Sprintf("      %s = %s\n", name, truncateValue(displayValue, 40)))
		}
	} else {
		info.WriteString("   No variables defined\n")
	}
	
	if len(requiredVars) > 0 {
		info.WriteString("   Required Variables:\n")
		for _, name := range requiredVars {
			if _, exists := availableVars[name]; exists {
				info.WriteString(fmt.Sprintf("      ✅ %s (available)\n", name))
			} else {
				info.WriteString(fmt.Sprintf("      ❌ %s (missing)\n", name))
			}
		}
	}
	
	return info.String()
}

// createResolvedVarInfo creates resolved variable info with security context
func (vrd *VariableResolutionDebugger) createResolvedVarInfo(varName string, value interface{}) ResolvedVarInfo {
	valueStr := fmt.Sprintf("%v", value)
	
	// Check if this is a SECRETS namespace variable
	if strings.HasPrefix(varName, "SECRETS.") {
		secretName := varName[8:] // Remove "SECRETS." prefix
		if vrd.secretManager != nil {
			if source, masked, exists := vrd.secretManager.GetSecretInfo(secretName); exists {
				displayValue := valueStr
				if masked {
					displayValue = "[MASKED]"
				}
				return ResolvedVarInfo{
					Value:        valueStr,
					IsSecret:     true,
					IsMasked:     masked,
					Source:       source,
					DisplayValue: displayValue,
				}
			}
		}
	}
	
	// Regular variable
	return ResolvedVarInfo{
		Value:        valueStr,
		IsSecret:     false,
		IsMasked:     false,
		Source:       "",
		DisplayValue: valueStr,
	}
}