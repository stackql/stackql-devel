package mcp_server

import (
	"context"
	"fmt"
	"time"
)

// ExampleBackend is a simple implementation of the Backend interface for demonstration purposes.
// This shows how to implement the Backend interface without depending on StackQL internals.
type ExampleBackend struct {
	connectionString string
	connected        bool
}

// NewExampleBackend creates a new example backend instance.
func NewExampleBackend(connectionString string) Backend {
	return &ExampleBackend{
		connectionString: connectionString,
		connected:        false,
	}
}

// Execute implements the Backend interface.
// This is a mock implementation that returns sample data.
func (b *ExampleBackend) Execute(ctx context.Context, query string, params map[string]interface{}) (QueryResult, error) {
	if !b.connected {
		return nil, &BackendError{
			Code:    "NOT_CONNECTED",
			Message: "Backend is not connected",
		}
	}
	
	startTime := time.Now()
	
	// Simulate query processing delay
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(50 * time.Millisecond):
		// Continue processing
	}
	
	// Mock response based on query content
	var result QueryResult
	
	if containsIgnoreCase(query, "select") {
		columns := []ColumnInfo{
			NewColumnInfo("id", "int64", false),
			NewColumnInfo("name", "string", true),
			NewColumnInfo("status", "string", false),
		}
		rows := [][]interface{}{
			{1, "example-instance-1", "running"},
			{2, "example-instance-2", "stopped"},
			{3, "example-instance-3", "running"},
		}
		result = NewQueryResult(columns, rows, 3, time.Since(startTime).Milliseconds())
	} else if containsIgnoreCase(query, "show") {
		columns := []ColumnInfo{
			NewColumnInfo("resource_name", "string", false),
			NewColumnInfo("provider", "string", false),
		}
		rows := [][]interface{}{
			{"instances", "aws"},
			{"buckets", "aws"},
			{"instances", "google"},
		}
		result = NewQueryResult(columns, rows, 3, time.Since(startTime).Milliseconds())
	} else {
		columns := []ColumnInfo{NewColumnInfo("result", "string", false)}
		rows := [][]interface{}{{"Query executed successfully"}}
		result = NewQueryResult(columns, rows, 1, time.Since(startTime).Milliseconds())
	}
	
	return result, nil
}

// GetSchema implements the Backend interface.
// Returns a mock schema structure representing available providers and resources.
func (b *ExampleBackend) GetSchema(ctx context.Context) (SchemaProvider, error) {
	if !b.connected {
		return nil, &BackendError{
			Code:    "NOT_CONNECTED",
			Message: "Backend is not connected",
		}
	}
	
	// Build AWS EC2 instances resource
	ec2Fields := []Field{
		NewField("instance_id", "string", true, "EC2 instance identifier"),
		NewField("instance_type", "string", false, "EC2 instance type"),
		NewField("state", "string", false, "Instance state"),
	}
	ec2Instances := NewResource("instances", []string{"select", "insert", "delete"}, ec2Fields)
	ec2Service := NewService("ec2", []Resource{ec2Instances})
	
	// Build AWS S3 buckets resource
	s3Fields := []Field{
		NewField("bucket_name", "string", true, "S3 bucket name"),
		NewField("creation_date", "string", false, "Bucket creation date"),
		NewField("region", "string", false, "AWS region"),
	}
	s3Buckets := NewResource("buckets", []string{"select", "insert", "delete"}, s3Fields)
	s3Service := NewService("s3", []Resource{s3Buckets})
	
	// Build AWS provider
	awsProvider := NewProvider("aws", "v1.0.0", []Service{ec2Service, s3Service})
	
	// Build Google Compute instances resource
	gceFields := []Field{
		NewField("name", "string", true, "Instance name"),
		NewField("machine_type", "string", false, "Machine type"),
		NewField("status", "string", false, "Instance status"),
		NewField("zone", "string", false, "Compute zone"),
	}
	gceInstances := NewResource("instances", []string{"select", "insert", "delete"}, gceFields)
	computeService := NewService("compute", []Resource{gceInstances})
	
	// Build Google provider
	googleProvider := NewProvider("google", "v1.0.0", []Service{computeService})
	
	// Create schema
	schema := NewSchemaProvider([]Provider{awsProvider, googleProvider})
	
	return schema, nil
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

// containsIgnoreCase checks if a string contains a substring (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase without using strings package to avoid dependencies.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + ('a' - 'A')
		} else {
			result[i] = b
		}
	}
	return string(result)
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