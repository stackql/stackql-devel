package mcp_server

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	
	if err := config.Validate(); err != nil {
		t.Fatalf("Default config validation failed: %v", err)
	}
	
	if config.Server.Name == "" {
		t.Error("Server name should not be empty")
	}
	
	if config.Server.Version == "" {
		t.Error("Server version should not be empty")
	}
	
	if len(config.Transport.EnabledTransports) == 0 {
		t.Error("At least one transport should be enabled by default")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "valid default config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "empty server name",
			config: &Config{
				Server: ServerConfig{
					Name:                  "",
					Version:               "1.0.0",
					MaxConcurrentRequests: 100,
				},
				Backend: BackendConfig{
					Type:           "stackql",
					MaxConnections: 10,
				},
				Transport: TransportConfig{
					EnabledTransports: []string{"stdio"},
				},
			},
			wantError: true,
		},
		{
			name: "invalid transport",
			config: &Config{
				Server: ServerConfig{
					Name:                  "Test Server",
					Version:               "1.0.0",
					MaxConcurrentRequests: 100,
				},
				Backend: BackendConfig{
					Type:           "stackql",
					MaxConnections: 10,
				},
				Transport: TransportConfig{
					EnabledTransports: []string{"invalid"},
				},
			},
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Config.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestExampleBackend(t *testing.T) {
	backend := NewExampleBackend("test://localhost")
	ctx := context.Background()
	
	// Test Ping
	if err := backend.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	
	// Test GetSchema
	schema, err := backend.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}
	
	if len(schema.Providers) == 0 {
		t.Error("Schema should contain at least one provider")
	}
	
	// Test Execute with SELECT query
	result, err := backend.Execute(ctx, "SELECT * FROM aws.ec2.instances", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	if len(result.Columns) == 0 {
		t.Error("Result should contain columns")
	}
	
	if len(result.Rows) == 0 {
		t.Error("Result should contain rows")
	}
	
	// Test Close
	if err := backend.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestMCPServerCreation(t *testing.T) {
	config := DefaultConfig()
	backend := NewExampleBackend("test://localhost")
	
	server, err := NewMCPServer(config, backend, nil)
	if err != nil {
		t.Fatalf("NewMCPServer failed: %v", err)
	}
	
	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestMCPRequestHandling(t *testing.T) {
	config := DefaultConfig()
	backend := NewExampleBackend("test://localhost")
	server, err := NewMCPServer(config, backend, nil)
	if err != nil {
		t.Fatalf("NewMCPServer failed: %v", err)
	}
	
	ctx := context.Background()
	
	// Test initialize request
	initReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{}`),
	}
	
	resp := server.handleMCPRequest(ctx, initReq)
	if resp.Error != nil {
		t.Fatalf("Initialize request failed: %v", resp.Error)
	}
	
	if resp.Result == nil {
		t.Error("Initialize response should contain result")
	}
	
	// Test resources/list request
	resourcesReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "resources/list",
		Params:  json.RawMessage(`{}`),
	}
	
	resp = server.handleMCPRequest(ctx, resourcesReq)
	if resp.Error != nil {
		t.Fatalf("Resources/list request failed: %v", resp.Error)
	}
	
	// Test tools/list request
	toolsReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/list",
		Params:  json.RawMessage(`{}`),
	}
	
	resp = server.handleMCPRequest(ctx, toolsReq)
	if resp.Error != nil {
		t.Fatalf("Tools/list request failed: %v", resp.Error)
	}
	
	// Test tools/call request
	toolsCallReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "stackql_query", "arguments": {"query": "SELECT * FROM aws.ec2.instances"}}`),
	}
	
	resp = server.handleMCPRequest(ctx, toolsCallReq)
	if resp.Error != nil {
		t.Fatalf("Tools/call request failed: %v", resp.Error)
	}
	
	// Test unknown method
	unknownReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "unknown/method",
		Params:  json.RawMessage(`{}`),
	}
	
	resp = server.handleMCPRequest(ctx, unknownReq)
	if resp.Error == nil {
		t.Error("Unknown method should return error")
	}
	
	if resp.Error.Code != -32601 {
		t.Errorf("Expected method not found error code -32601, got %d", resp.Error.Code)
	}
}

func TestDurationMarshaling(t *testing.T) {
	d := Duration(30 * time.Second)
	
	// Test JSON marshaling
	jsonData, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	
	var d2 Duration
	if err := json.Unmarshal(jsonData, &d2); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	
	if time.Duration(d) != time.Duration(d2) {
		t.Errorf("Duration mismatch after JSON round-trip: %v != %v", d, d2)
	}
}

func TestBackendError(t *testing.T) {
	err := &BackendError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Details: map[string]interface{}{"field": "value"},
	}
	
	if err.Error() != "Test error message" {
		t.Errorf("Expected error message 'Test error message', got '%s'", err.Error())
	}
	
	// Test Value() method for database compatibility
	val, dbErr := err.Value()
	if dbErr != nil {
		t.Fatalf("Value() failed: %v", dbErr)
	}
	
	if val != "Test error message" {
		t.Errorf("Expected value 'Test error message', got '%v'", val)
	}
}

func TestNewMCPServerWithExampleBackend(t *testing.T) {
	server, err := NewMCPServerWithExampleBackend(nil)
	if err != nil {
		t.Fatalf("NewMCPServerWithExampleBackend failed: %v", err)
	}
	
	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestConfigLoading(t *testing.T) {
	// Test JSON config loading
	jsonConfig := `{
		"server": {
			"name": "Test Server",
			"version": "1.0.0",
			"max_concurrent_requests": 50,
			"request_timeout": "15s"
		},
		"backend": {
			"type": "stackql",
			"max_connections": 5
		},
		"transport": {
			"enabled_transports": ["stdio"]
		},
		"logging": {
			"level": "debug"
		}
	}`
	
	config, err := LoadFromJSON([]byte(jsonConfig))
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}
	
	if config.Server.Name != "Test Server" {
		t.Errorf("Expected server name 'Test Server', got '%s'", config.Server.Name)
	}
	
	if config.Server.MaxConcurrentRequests != 50 {
		t.Errorf("Expected max concurrent requests 50, got %d", config.Server.MaxConcurrentRequests)
	}
	
	// Test YAML config loading
	yamlConfig := `
server:
  name: "YAML Test Server"
  version: "2.0.0"
  max_concurrent_requests: 75
backend:
  type: "stackql"
  max_connections: 8
transport:
  enabled_transports: ["tcp"]
logging:
  level: "warn"
`
	
	config, err = LoadFromYAML([]byte(yamlConfig))
	if err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}
	
	if config.Server.Name != "YAML Test Server" {
		t.Errorf("Expected server name 'YAML Test Server', got '%s'", config.Server.Name)
	}
	
	if config.Server.MaxConcurrentRequests != 75 {
		t.Errorf("Expected max concurrent requests 75, got %d", config.Server.MaxConcurrentRequests)
	}
}