package actions

import (
	"context"
	"database/sql"

	"cloud.google.com/go/spanner"
	"github.com/JianLoong/robogo/internal/util"
)

// ActionContext holds all dependencies for actions to eliminate global state
type ActionContext struct {
	PostgresManager *PostgreSQLManager
	SpannerManager  *SpannerManager
	ConfigManager   *util.ConfigManager
}

// NewActionContext creates a new action context with initialized managers
func NewActionContext() *ActionContext {
	return &ActionContext{
		PostgresManager: &PostgreSQLManager{
			connections: make(map[string]*sql.DB),
		},
		SpannerManager: &SpannerManager{
			connections: make(map[string]*spanner.Client),
		},
		ConfigManager: util.NewConfigManager(),
	}
}

// GetActionContext retrieves the ActionContext from the context
func GetActionContext(ctx context.Context) *ActionContext {
	if actionCtx, ok := ctx.Value(actionContextKey).(*ActionContext); ok {
		return actionCtx
	}
	// Create a new action context if none exists
	return NewActionContext()
}

// WithActionContext adds the ActionContext to the context
func WithActionContext(ctx context.Context, actionCtx *ActionContext) context.Context {
	return context.WithValue(ctx, actionContextKey, actionCtx)
}

// contextKey type for context values
type contextKey string

const actionContextKey contextKey = "actionContext"

// Cleanup closes all connections and cleans up resources
func (ac *ActionContext) Cleanup() {
	if ac.PostgresManager != nil {
		ac.PostgresManager.CloseAll()
	}
	if ac.SpannerManager != nil {
		ac.SpannerManager.CloseAll()
	}
}