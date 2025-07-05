package actions

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// SecretManager handles secret variable resolution (inline or file)
type SecretManager struct {
	secrets map[string]string
	masked  map[string]bool
}

// NewSecretManager creates a new secret manager
func NewSecretManager() *SecretManager {
	return &SecretManager{
		secrets: make(map[string]string),
		masked:  make(map[string]bool),
	}
}

// ResolveSecrets resolves all secrets for a test case (inline or file)
func (sm *SecretManager) ResolveSecrets(variables map[string]interface{}) error {
	for name, secretInterface := range variables {
		if secretMap, ok := secretInterface.(map[string]interface{}); ok {
			var secretValue string
			maskOutput := true
			if mask, hasMask := secretMap["mask_output"]; hasMask {
				maskOutput = fmt.Sprintf("%v", mask) == "true"
			}
			if value, hasValue := secretMap["value"]; hasValue && fmt.Sprintf("%v", value) != "" {
				secretValue = fmt.Sprintf("%v", value)
			} else if file, hasFile := secretMap["file"]; hasFile && fmt.Sprintf("%v", file) != "" {
				filePath := fmt.Sprintf("%v", file)
				data, err := ioutil.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read secret file '%s': %w", filePath, err)
				}
				secretValue = strings.TrimSpace(string(data))
			} else {
				return fmt.Errorf("secret '%s' must have either a 'value' or 'file' field", name)
			}
			sm.secrets[name] = secretValue
			sm.masked[name] = maskOutput
		}
	}
	return nil
}

// GetSecret retrieves a secret value
func (sm *SecretManager) GetSecret(name string) (string, bool) {
	value, exists := sm.secrets[name]
	return value, exists
}

// IsMasked checks if a secret should be masked in output
func (sm *SecretManager) IsMasked(name string) bool {
	return sm.masked[name]
}

// MaskValue masks a value for secure output
func (sm *SecretManager) MaskValue(value string) string {
	if len(value) == 0 {
		return "****"
	}
	if len(value) <= 4 {
		return "****"
	}
	return "****" + fmt.Sprintf("(%d chars)", len(value))
}

// MaskSecretsInString masks all secrets in a string
func (sm *SecretManager) MaskSecretsInString(s string) string {
	result := s
	for name, value := range sm.secrets {
		if sm.IsMasked(name) {
			maskedValue := sm.MaskValue(value)
			result = strings.ReplaceAll(result, value, maskedValue)
		}
	}
	return result
}
