package mcp_server

import (
	"context"
	"database/sql/driver"
)

// Backend defines the interface for executing queries from MCP clients.
// This abstraction allows for different backend implementations (in-memory, TCP, etc.)
// while maintaining compatibility with the MCP protocol.
type Backend interface {
	// Execute runs a query and returns the results.
	// The query string and parameters are provided by the MCP client.
	Execute(ctx context.Context, query string, params map[string]interface{}) (*QueryResult, error)
	
	// GetSchema returns metadata about available resources and their structure.
	// This is used by MCP clients to understand what data is available.
	GetSchema(ctx context.Context) (*Schema, error)
	
	// Ping verifies the backend connection is active.
	Ping(ctx context.Context) error
	
	// Close gracefully shuts down the backend connection.
	Close() error
}

// QueryResult represents the result of a query execution.
type QueryResult struct {
	// Columns contains metadata about each column in the result set.
	Columns []ColumnInfo `json:"columns"`
	
	// Rows contains the actual data returned by the query.
	Rows [][]interface{} `json:"rows"`
	
	// RowsAffected indicates the number of rows affected by DML operations.
	RowsAffected int64 `json:"rows_affected"`
	
	// ExecutionTime is the time taken to execute the query in milliseconds.
	ExecutionTime int64 `json:"execution_time_ms"`
}

// ColumnInfo provides metadata about a result column.
type ColumnInfo struct {
	// Name is the column name as returned by the query.
	Name string `json:"name"`
	
	// Type is the data type of the column (e.g., "string", "int64", "float64").
	Type string `json:"type"`
	
	// Nullable indicates whether the column can contain null values.
	Nullable bool `json:"nullable"`
}

// Schema represents the metadata structure of available resources.
type Schema struct {
	// Providers lists all available providers (e.g., aws, google, azure).
	Providers []Provider `json:"providers"`
}

// Provider represents a StackQL provider with its services and resources.
type Provider struct {
	// Name is the provider identifier (e.g., "aws", "google").
	Name string `json:"name"`
	
	// Version is the provider version.
	Version string `json:"version"`
	
	// Services lists all services available in this provider.
	Services []Service `json:"services"`
}

// Service represents a service within a provider.
type Service struct {
	// Name is the service identifier (e.g., "ec2", "compute").
	Name string `json:"name"`
	
	// Resources lists all resources available in this service.
	Resources []Resource `json:"resources"`
}

// Resource represents a queryable resource.
type Resource struct {
	// Name is the resource identifier (e.g., "instances", "buckets").
	Name string `json:"name"`
	
	// Methods lists the available operations for this resource.
	Methods []string `json:"methods"`
	
	// Fields describes the available fields in this resource.
	Fields []Field `json:"fields"`
}

// Field represents a field within a resource.
type Field struct {
	// Name is the field identifier.
	Name string `json:"name"`
	
	// Type is the field data type.
	Type string `json:"type"`
	
	// Required indicates if this field is mandatory for certain operations.
	Required bool `json:"required"`
	
	// Description provides human-readable documentation for the field.
	Description string `json:"description,omitempty"`
}

// BackendError represents an error that occurred in the backend.
type BackendError struct {
	// Code is a machine-readable error code.
	Code string `json:"code"`
	
	// Message is a human-readable error message.
	Message string `json:"message"`
	
	// Details contains additional context about the error.
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *BackendError) Error() string {
	return e.Message
}

// Ensure BackendError implements the driver.Valuer interface for database compatibility
func (e *BackendError) Value() (driver.Value, error) {
	return e.Message, nil
}