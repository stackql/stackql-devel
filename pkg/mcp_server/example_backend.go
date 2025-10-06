package mcp_server //nolint:revive // fine for now

import (
	"context"
	"time"
)

const (
	ExplainerForeignKeyStackql = "At present, foreign keys are not meaningfully supported in stackql."
	ExplainerFindRelationships = "At present, relationship finding is not meaningfully supported in stackql."
)

// ExampleBackend is a simple implementation of the Backend interface for demonstration purposes.
// This shows how to implement the Backend interface without depending on StackQL internals.
type ExampleBackend struct {
	connectionString string
	connected        bool
}

// Stub all Backend interface methods below

func (b *ExampleBackend) Greet(ctx context.Context, args greetInput) (string, error) {
	return "Hi " + args.Name, nil
}

func (b *ExampleBackend) ServerInfo(ctx context.Context, _ any) (serverInfoOutput, error) {
	return serverInfoOutput{
		Name:       "Stackql explorer",
		Info:       "This is an example server.",
		IsReadOnly: false,
	}, nil
}

// Please adjust all below to sensible signatures in keeping with what is above.
// Do it now!
func (b *ExampleBackend) DBIdentity(ctx context.Context, _ any) (map[string]any, error) {
	return map[string]any{
		"identity": "stub",
	}, nil
}

func (b *ExampleBackend) Query(ctx context.Context, sql string, parameters []interface{}, rowLimit int, format string) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) QueryJSON(ctx context.Context, sql string, parameters []interface{}, rowLimit int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (b *ExampleBackend) RunQuery(ctx context.Context, args queryInput) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) RunQueryJSON(ctx context.Context, input queryJSONInput) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (b *ExampleBackend) ListTableResources(ctx context.Context, schema string) ([]string, error) {
	return []string{}, nil
}

func (b *ExampleBackend) ReadTableResource(ctx context.Context, schema string, table string, rowLimit int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (b *ExampleBackend) PromptWriteSafeSelectTool(ctx context.Context) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) PromptExplainPlanTipsTool(ctx context.Context) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) ListTablesJSON(ctx context.Context, input listTablesInput) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (b *ExampleBackend) ListTablesJSONPage(ctx context.Context, input listTablesPageInput) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (b *ExampleBackend) ListTables(ctx context.Context, hI hierarchyInput) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) DescribeTable(ctx context.Context, hI hierarchyInput) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) GetForeignKeys(ctx context.Context, hI hierarchyInput) (string, error) {
	return ExplainerForeignKeyStackql, nil
}

func (b *ExampleBackend) FindRelationships(ctx context.Context, hI hierarchyInput) (string, error) {
	return ExplainerFindRelationships, nil
}

func (b *ExampleBackend) ListProviders(ctx context.Context) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) ListServices(ctx context.Context, hI hierarchyInput) (string, error) {
	return "stub", nil
}

func (b *ExampleBackend) ListResources(ctx context.Context, hI hierarchyInput) (string, error) {
	return "stub", nil
}

// NewExampleBackend creates a new example backend instance.
func NewExampleBackend(connectionString string) Backend {
	return &ExampleBackend{
		connectionString: connectionString,
		connected:        false,
	}
}

// Ping implements the Backend interface.
func (b *ExampleBackend) Ping(ctx context.Context) error {
	if !b.connected {
		// Simulate connection establishment
		b.connected = true
	}

	// Simulate a ping operation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return nil
	}
}

// Close implements the Backend interface.
func (b *ExampleBackend) Close() error {
	b.connected = false
	return nil
}

// NewMCPServerWithExampleBackend creates a new MCP server with an example backend.
// This is a convenience function for testing and demonstration purposes.
func NewMCPServerWithExampleBackend(config *Config) (MCPServer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	backend := NewExampleBackend(config.Backend.ConnectionString)

	return NewMCPServer(config, backend, nil)
}
