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
func NewExampleBackend(connectionString string) *ExampleBackend {
	return &ExampleBackend{
		connectionString: connectionString,
		connected:        false,
	}
}

// Execute implements the Backend interface.
// This is a mock implementation that returns sample data.
func (b *ExampleBackend) Execute(ctx context.Context, query string, params map[string]interface{}) (*QueryResult, error) {
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
	var result *QueryResult
	
	if containsIgnoreCase(query, "select") {
		result = &QueryResult{
			Columns: []ColumnInfo{
				{Name: "id", Type: "int64", Nullable: false},
				{Name: "name", Type: "string", Nullable: true},
				{Name: "status", Type: "string", Nullable: false},
			},
			Rows: [][]interface{}{
				{1, "example-instance-1", "running"},
				{2, "example-instance-2", "stopped"},
				{3, "example-instance-3", "running"},
			},
			RowsAffected:  3,
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}
	} else if containsIgnoreCase(query, "show") {
		result = &QueryResult{
			Columns: []ColumnInfo{
				{Name: "resource_name", Type: "string", Nullable: false},
				{Name: "provider", Type: "string", Nullable: false},
			},
			Rows: [][]interface{}{
				{"instances", "aws"},
				{"buckets", "aws"},
				{"instances", "google"},
			},
			RowsAffected:  3,
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}
	} else {
		result = &QueryResult{
			Columns:       []ColumnInfo{{Name: "result", Type: "string", Nullable: false}},
			Rows:          [][]interface{}{{"Query executed successfully"}},
			RowsAffected:  1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}
	}
	
	return result, nil
}

// GetSchema implements the Backend interface.
// Returns a mock schema structure representing available providers and resources.
func (b *ExampleBackend) GetSchema(ctx context.Context) (*Schema, error) {
	if !b.connected {
		return nil, &BackendError{
			Code:    "NOT_CONNECTED",
			Message: "Backend is not connected",
		}
	}
	
	schema := &Schema{
		Providers: []Provider{
			{
				Name:    "aws",
				Version: "v1.0.0",
				Services: []Service{
					{
						Name: "ec2",
						Resources: []Resource{
							{
								Name:    "instances",
								Methods: []string{"select", "insert", "delete"},
								Fields: []Field{
									{Name: "instance_id", Type: "string", Required: true, Description: "EC2 instance identifier"},
									{Name: "instance_type", Type: "string", Required: false, Description: "EC2 instance type"},
									{Name: "state", Type: "string", Required: false, Description: "Instance state"},
								},
							},
						},
					},
					{
						Name: "s3",
						Resources: []Resource{
							{
								Name:    "buckets",
								Methods: []string{"select", "insert", "delete"},
								Fields: []Field{
									{Name: "bucket_name", Type: "string", Required: true, Description: "S3 bucket name"},
									{Name: "creation_date", Type: "string", Required: false, Description: "Bucket creation date"},
									{Name: "region", Type: "string", Required: false, Description: "AWS region"},
								},
							},
						},
					},
				},
			},
			{
				Name:    "google",
				Version: "v1.0.0",
				Services: []Service{
					{
						Name: "compute",
						Resources: []Resource{
							{
								Name:    "instances",
								Methods: []string{"select", "insert", "delete"},
								Fields: []Field{
									{Name: "name", Type: "string", Required: true, Description: "Instance name"},
									{Name: "machine_type", Type: "string", Required: false, Description: "Machine type"},
									{Name: "status", Type: "string", Required: false, Description: "Instance status"},
									{Name: "zone", Type: "string", Required: false, Description: "Compute zone"},
								},
							},
						},
					},
				},
			},
		},
	}
	
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
func NewMCPServerWithExampleBackend(config *Config) (*MCPServer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	backend := NewExampleBackend(config.Backend.ConnectionString)
	
	return NewMCPServer(config, backend, nil)
}