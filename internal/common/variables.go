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

// Load bulk loads variables
func (v *Variables) Load(vars map[string]any) {
	v.store.Load(vars)
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
