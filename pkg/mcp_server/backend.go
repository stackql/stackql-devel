package mcp_server //nolint:revive // fine for now

import (
	"context"
	"database/sql/driver"
)

// Backend defines the interface for executing queries from MCP clients.
// This abstraction allows for different backend implementations (in-memory, TCP, etc.)
// while maintaining compatibility with the MCP protocol.

/*
The Backend interface should include all of the tools from the below python snippet:

```python

@mcp.tool()
def server_info() -> Dict[str, Any]:

	"""Return server and environment info useful for clients."""
	return _BACKEND.server_info()

@mcp.tool()
def db_identity() -> Dict[str, Any]:

	"""Return current DB identity details: db, user, host, port, search_path, server version, cluster name."""
	return _BACKEND.db_identity()

@mcp.tool()
def query(

	sql: str,
	parameters: Optional[List[Any]] = None,
	row_limit: int = 500,
	format: str = "markdown",

) -> str:

	"""Execute a SQL query (legacy signature). Prefer run_query with typed input."""
	return _BACKEND.query(sql, parameters, row_limit, format)

@mcp.tool()
def query_json(sql: str, parameters: Optional[List[Any]] = None, row_limit: int = 500) -> List[Dict[str, Any]]:

	"""Execute a SQL query and return JSON-serializable rows (legacy signature). Prefer run_query_json with typed input."""
	return _BACKEND.query_json(sql, parameters, row_limit)

@mcp.tool()
def run_query(input: QueryInput) -> str:

	"""Execute a SQL query with typed input (preferred)."""
	return _BACKEND.run_query(input)

@mcp.tool()
def run_query_json(input: QueryJSONInput) -> List[Dict[str, Any]]:

	"""Execute a SQL query and return JSON rows with typed input (preferred)."""
	return _BACKEND.run_query_json(input)

@mcp.tool()
def list_table_resources(schema: str = 'public') -> List[str]:

	"""List resource URIs for tables in a schema (fallback for clients without resource support)."""
	return _BACKEND.list_table_resources(schema)

@mcp.tool()
def read_table_resource(schema: str, table: str, row_limit: int = 100) -> List[Dict[str, Any]]:

	"""Read rows from a table resource (fallback)."""
	return _BACKEND.read_table_resource(schema, table, row_limit)

# Try to register proper MCP resources if available in FastMCP

try:

	resource_decorator = getattr(mcp, "resource")
	if callable(resource_decorator):
	    @resource_decorator("table://{schema}/{table}") # type: ignore
	    def table_resource(schema: str, table: str, row_limit: int = 100):
	        """Resource reader for table rows."""
	        rows = _BACKEND.read_table_resource(schema, table, row_limit=row_limit)
	        # Return as JSON string to be universally consumable
	        return json.dumps(rows, default=str)

except Exception as e:

	logger.debug(f"Resource registration skipped: {e}")

try:

	prompt_decorator = getattr(mcp, "prompt")
	if callable(prompt_decorator):
	    @prompt_decorator("write_safe_select") # type: ignore
	    def prompt_write_safe_select():
	        return _BACKEND.prompt_write_safe_select_tool()

	    @prompt_decorator("explain_plan_tips") # type: ignore
	    def prompt_explain_plan_tips():
	        return _BACKEND.prompt_explain_plan_tips_tool()

except Exception as e:

	logger.debug(f"Prompt registration skipped: {e}")

#

@mcp.tool()
def prompt_write_safe_select_tool() -> str:

	"""Prompt: guidelines for writing safe SELECT queries."""
	return _BACKEND.prompt_write_safe_select_tool()

@mcp.tool()
def prompt_explain_plan_tips_tool() -> str:

	"""Prompt: tips for reading EXPLAIN ANALYZE output."""
	return _BACKEND.prompt_explain_plan_tips_tool()

@mcp.tool()
def list_schemas_json(input: ListSchemasInput) -> List[Dict[str, Any]]:

	"""List schemas with filters and return JSON rows."""
	return _BACKEND.list_schemas_json(input)

@mcp.tool()
def list_schemas_json_page(input: ListSchemasPageInput) -> Dict[str, Any]:

	"""List schemas with pagination and filters. Returns { items: [...], next_cursor: str|null }"""
	return _BACKEND.list_schemas_json_page(input)

@mcp.tool()
def list_tables_json(input: ListTablesInput) -> List[Dict[str, Any]]:

	"""List tables in a schema with optional filters and return JSON rows."""
	return _BACKEND.list_tables_json(input)

@mcp.tool()
def list_tables_json_page(input: ListTablesPageInput) -> Dict[str, Any]:

	"""List tables with pagination and filters. Returns { items, next_cursor }."""
	return _BACKEND.list_tables_json_page(input)

@mcp.tool()
def list_schemas() -> str:

	"""List all schemas in the database."""
	return _BACKEND.list_schemas()

@mcp.tool()
def list_tables(db_schema: Optional[str] = None) -> str:

	"""List all tables in a specific schema.

	Args:
	    db_schema: The schema name to list tables from (defaults to 'public')
	"""
	return _BACKEND.list_tables(db_schema)

@mcp.tool()
def describe_table(table_name: str, db_schema: Optional[str] = None) -> str:

	"""Get detailed information about a table.
	When dealing with a stackql backend (ie: when the server is initialised to consume stackql using the 'dbapp' parameter), the required query input and returned schema can differ even across the one "resource" (table) object.
	This is because stackql has required where parameters for some access methods, where this can vary be SQL verb.
	In line with this, stackql responses will contain information about required where parameters, if applicable.

	Args:
	    table_name: The name of the table to describ
	    db_schema: The schema name (defaults to 'public')
	"""
	return _BACKEND.describe_table(table_name, db_schema=db_schema)

@mcp.tool()
def get_foreign_keys(table_name: str, db_schema: Optional[str] = None) -> str:

	"""Get foreign key information for a table.

	Args:
	    table_name: The name of the table to get foreign keys from
	    db_schema: The schema name (defaults to 'public')
	"""
	return _BACKEND.get_foreign_keys(table_name, db_schema)

@mcp.tool()
def find_relationships(table_name: str, db_schema: Optional[str] = None) -> str:

	"""Find both explicit and implied relationships for a table.

	Args:
	    table_name: The name of the table to analyze relationships for
	    db_schema: The schema name (defaults to 'public')
	"""
	return _BACKEND.find_relationships(table_name, db_schema)

```
*/
type Backend interface {

	// Ping verifies the backend connection is active.
	Ping(ctx context.Context) error

	// Close gracefully shuts down the backend connection.
	Close() error
	// Server and environment info
	ServerInfo(ctx context.Context) (map[string]interface{}, error)

	// Current DB identity details
	DBIdentity(ctx context.Context) (map[string]interface{}, error)

	// Execute a SQL query (legacy signature)
	Query(ctx context.Context, sql string, parameters []interface{}, rowLimit int, format string) (string, error)

	// Execute a SQL query and return JSON-serializable rows (legacy signature)
	QueryJSON(ctx context.Context, sql string, parameters []interface{}, rowLimit int) ([]map[string]interface{}, error)

	// Execute a SQL query with typed input (preferred)
	RunQuery(ctx context.Context, input QueryInput) (string, error)

	// Execute a SQL query and return JSON rows with typed input (preferred)
	RunQueryJSON(ctx context.Context, input QueryJSONInput) ([]map[string]interface{}, error)

	// List resource URIs for tables in a schema
	ListTableResources(ctx context.Context, schema string) ([]string, error)

	// Read rows from a table resource
	ReadTableResource(ctx context.Context, schema string, table string, rowLimit int) ([]map[string]interface{}, error)

	// Prompt: guidelines for writing safe SELECT queries
	PromptWriteSafeSelectTool(ctx context.Context) (string, error)

	// Prompt: tips for reading EXPLAIN ANALYZE output
	PromptExplainPlanTipsTool(ctx context.Context) (string, error)

	// List schemas with filters and return JSON rows
	ListSchemasJSON(ctx context.Context, input ListSchemasInput) ([]map[string]interface{}, error)

	// List schemas with pagination and filters
	ListSchemasJSONPage(ctx context.Context, input ListSchemasPageInput) (map[string]interface{}, error)

	// List tables in a schema with optional filters and return JSON rows
	ListTablesJSON(ctx context.Context, input ListTablesInput) ([]map[string]interface{}, error)

	// List tables with pagination and filters
	ListTablesJSONPage(ctx context.Context, input ListTablesPageInput) (map[string]interface{}, error)

	// List all schemas in the database
	ListSchemas(ctx context.Context) (string, error)

	// List all tables in a specific schema
	ListTables(ctx context.Context, dbSchema string) (string, error)

	// Get detailed information about a table
	DescribeTable(ctx context.Context, tableName string, dbSchema string) (string, error)

	// Get foreign key information for a table
	GetForeignKeys(ctx context.Context, tableName string, dbSchema string) (string, error)

	// Find both explicit and implied relationships for a table
	FindRelationships(ctx context.Context, tableName string, dbSchema string) (string, error)
}

// QueryResult represents the result of a query execution.
type QueryResult interface {
	// GetColumns returns metadata about each column in the result set.
	GetColumns() []ColumnInfo

	// GetRows returns the actual data returned by the query.
	GetRows() [][]interface{}

	// GetRowsAffected returns the number of rows affected by DML operations.
	GetRowsAffected() int64

	// GetExecutionTime returns the time taken to execute the query in milliseconds.
	GetExecutionTime() int64
}

// ColumnInfo provides metadata about a result column.
type ColumnInfo interface {
	// GetName returns the column name as returned by the query.
	GetName() string

	// GetType returns the data type of the column (e.g., "string", "int64", "float64").
	GetType() string

	// IsNullable indicates whether the column can contain null values.
	IsNullable() bool
}

// SchemaProvider represents the metadata structure of available resources.
type SchemaProvider interface {
	// GetProviders returns all available providers (e.g., aws, google, azure).
	GetProviders() []Provider
}

// Provider represents a StackQL provider with its services and resources.
type Provider interface {
	// GetName returns the provider identifier (e.g., "aws", "google").
	GetName() string

	// GetVersion returns the provider version.
	GetVersion() string

	// GetServices returns all services available in this provider.
	GetServices() []Service
}

// Service represents a service within a provider.
type Service interface {
	// GetName returns the service identifier (e.g., "ec2", "compute").
	GetName() string

	// GetResources returns all resources available in this service.
	GetResources() []Resource
}

// Resource represents a queryable resource.
type Resource interface {
	// GetName returns the resource identifier (e.g., "instances", "buckets").
	GetName() string

	// GetMethods returns the available operations for this resource.
	GetMethods() []string

	// GetFields returns the available fields in this resource.
	GetFields() []Field
}

// Field represents a field within a resource.
type Field interface {
	// GetName returns the field identifier.
	GetName() string

	// GetType returns the field data type.
	GetType() string

	// IsRequired indicates if this field is mandatory for certain operations.
	IsRequired() bool

	// GetDescription returns human-readable documentation for the field.
	GetDescription() string
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

// Ensure BackendError implements the driver.Valuer interface for database compatibility.
func (e *BackendError) Value() (driver.Value, error) {
	return e.Message, nil
}

// Private implementations of interfaces

type queryResult struct {
	Columns       []ColumnInfo    `json:"columns"`
	Rows          [][]interface{} `json:"rows"`
	RowsAffected  int64           `json:"rows_affected"`
	ExecutionTime int64           `json:"execution_time_ms"`
}

func (qr *queryResult) GetColumns() []ColumnInfo { return qr.Columns }
func (qr *queryResult) GetRows() [][]interface{} { return qr.Rows }
func (qr *queryResult) GetRowsAffected() int64   { return qr.RowsAffected }
func (qr *queryResult) GetExecutionTime() int64  { return qr.ExecutionTime }

type columnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

func (ci *columnInfo) GetName() string  { return ci.Name }
func (ci *columnInfo) GetType() string  { return ci.Type }
func (ci *columnInfo) IsNullable() bool { return ci.Nullable }

type schemaProvider struct {
	Providers []Provider `json:"providers"`
}

func (sp *schemaProvider) GetProviders() []Provider { return sp.Providers }

type provider struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Services []Service `json:"services"`
}

func (p *provider) GetName() string        { return p.Name }
func (p *provider) GetVersion() string     { return p.Version }
func (p *provider) GetServices() []Service { return p.Services }

type service struct {
	Name      string     `json:"name"`
	Resources []Resource `json:"resources"`
}

func (s *service) GetName() string          { return s.Name }
func (s *service) GetResources() []Resource { return s.Resources }

type resource struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Fields  []Field  `json:"fields"`
}

func (r *resource) GetName() string      { return r.Name }
func (r *resource) GetMethods() []string { return r.Methods }
func (r *resource) GetFields() []Field   { return r.Fields }

type field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

func (f *field) GetName() string        { return f.Name }
func (f *field) GetType() string        { return f.Type }
func (f *field) IsRequired() bool       { return f.Required }
func (f *field) GetDescription() string { return f.Description }

// Factory functions

// NewQueryResult creates a new QueryResult instance.
func NewQueryResult(columns []ColumnInfo, rows [][]interface{}, rowsAffected, executionTime int64) QueryResult {
	return &queryResult{
		Columns:       columns,
		Rows:          rows,
		RowsAffected:  rowsAffected,
		ExecutionTime: executionTime,
	}
}

// NewColumnInfo creates a new ColumnInfo instance.
func NewColumnInfo(name, colType string, nullable bool) ColumnInfo {
	return &columnInfo{
		Name:     name,
		Type:     colType,
		Nullable: nullable,
	}
}

// NewSchemaProvider creates a new SchemaProvider instance.
func NewSchemaProvider(providers []Provider) SchemaProvider {
	return &schemaProvider{
		Providers: providers,
	}
}

// NewProvider creates a new Provider instance.
func NewProvider(name, version string, services []Service) Provider {
	return &provider{
		Name:     name,
		Version:  version,
		Services: services,
	}
}

// NewService creates a new Service instance.
func NewService(name string, resources []Resource) Service {
	return &service{
		Name:      name,
		Resources: resources,
	}
}

// NewResource creates a new Resource instance.
func NewResource(name string, methods []string, fields []Field) Resource {
	return &resource{
		Name:    name,
		Methods: methods,
		Fields:  fields,
	}
}

// NewField creates a new Field instance.
func NewField(name, fieldType string, required bool, description string) Field {
	return &field{
		Name:        name,
		Type:        fieldType,
		Required:    required,
		Description: description,
	}
}
