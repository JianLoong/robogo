package common

// Variables - variable storage and substitution (simplified, no complex expressions)
type Variables struct {
	store  *VariableStore
	engine *SubstitutionEngine
}

// NewVariables creates a Variables instance with simple substitution
func NewVariables() *Variables {
	store := NewVariableStore()
	engine := NewSubstitutionEngine(store)

	return &Variables{
		store:  store,
		engine: engine,
	}
}

// Set stores a variable
func (v *Variables) Set(key string, value any) {
	v.store.Set(key, value)
}

// Get retrieves a variable
func (v *Variables) Get(key string) any {
	return v.store.Get(key)
}

// Load bulk loads variables with environment variable substitution
func (v *Variables) Load(vars map[string]any) {
	// Substitute environment variables in the values before storing
	substitutedVars := make(map[string]any)
	for key, value := range vars {
		if str, ok := value.(string); ok {
			// Perform substitution on string values
			substitutedVars[key] = v.engine.Substitute(str)
		} else {
			substitutedVars[key] = value
		}
	}
	v.store.Load(substitutedVars)
}

// GetSnapshot returns a copy of all current variables for context enrichment
func (v *Variables) GetSnapshot() map[string]interface{} {
	return v.store.GetSnapshot()
}

// Substitute performs variable substitution using ${variable} syntax
func (v *Variables) Substitute(template string) string {
	return v.engine.Substitute(template)
}

// SubstituteArgs performs variable substitution on arguments
func (v *Variables) SubstituteArgs(args []any) []any {
	return v.engine.SubstituteArgs(args)
}

// Clone creates a copy of the Variables with the same data
func (v *Variables) Clone() *Variables {
	// Create new Variables instance
	newVars := NewVariables()
	
	// Copy all variables from current instance
	snapshot := v.GetSnapshot()
	for key, value := range snapshot {
		newVars.Set(key, value)
	}
	
	return newVars
}
