package mcp_server

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the complete configuration for the MCP server.
type Config struct {
	// Server contains server-specific configuration.
	Server ServerConfig `json:"server" yaml:"server"`
	
	// Backend contains backend-specific configuration.
	Backend BackendConfig `json:"backend" yaml:"backend"`
	
	// Transport contains transport layer configuration.
	Transport TransportConfig `json:"transport" yaml:"transport"`
	
	// Logging contains logging configuration.
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// ServerConfig contains configuration for the MCP server itself.
type ServerConfig struct {
	// Name is the server name advertised to clients.
	Name string `json:"name" yaml:"name"`
	
	// Version is the server version advertised to clients.
	Version string `json:"version" yaml:"version"`
	
	// Description is a human-readable description of the server.
	Description string `json:"description" yaml:"description"`
	
	// MaxConcurrentRequests limits the number of concurrent client requests.
	MaxConcurrentRequests int `json:"max_concurrent_requests" yaml:"max_concurrent_requests"`
	
	// RequestTimeout specifies the timeout for individual requests.
	RequestTimeout Duration `json:"request_timeout" yaml:"request_timeout"`
}

// BackendConfig contains configuration for the backend connection.
type BackendConfig struct {
	// Type specifies the backend type ("stackql", "tcp", "memory").
	Type string `json:"type" yaml:"type"`
	
	// ConnectionString contains the connection details for the backend.
	// Format depends on the backend type.
	ConnectionString string `json:"connection_string" yaml:"connection_string"`
	
	// MaxConnections limits the number of backend connections.
	MaxConnections int `json:"max_connections" yaml:"max_connections"`
	
	// ConnectionTimeout specifies the timeout for backend connections.
	ConnectionTimeout Duration `json:"connection_timeout" yaml:"connection_timeout"`
	
	// QueryTimeout specifies the timeout for individual queries.
	QueryTimeout Duration `json:"query_timeout" yaml:"query_timeout"`
	
	// RetryConfig contains retry policy configuration.
	Retry RetryConfig `json:"retry" yaml:"retry"`
}

// TransportConfig contains configuration for MCP transport layers.
type TransportConfig struct {
	// EnabledTransports lists which transports to enable (stdio, tcp, websocket).
	EnabledTransports []string `json:"enabled_transports" yaml:"enabled_transports"`
	
	// StdioConfig contains stdio transport configuration.
	Stdio StdioTransportConfig `json:"stdio" yaml:"stdio"`
	
	// TCPConfig contains TCP transport configuration.
	TCP TCPTransportConfig `json:"tcp" yaml:"tcp"`
	
	// WebSocketConfig contains WebSocket transport configuration.
	WebSocket WebSocketTransportConfig `json:"websocket" yaml:"websocket"`
}

// StdioTransportConfig contains configuration for stdio transport.
type StdioTransportConfig struct {
	// BufferSize specifies the buffer size for stdio operations.
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`
}

// TCPTransportConfig contains configuration for TCP transport.
type TCPTransportConfig struct {
	// Address specifies the TCP listen address.
	Address string `json:"address" yaml:"address"`
	
	// Port specifies the TCP listen port.
	Port int `json:"port" yaml:"port"`
	
	// MaxConnections limits the number of concurrent TCP connections.
	MaxConnections int `json:"max_connections" yaml:"max_connections"`
	
	// ReadTimeout specifies the timeout for read operations.
	ReadTimeout Duration `json:"read_timeout" yaml:"read_timeout"`
	
	// WriteTimeout specifies the timeout for write operations.
	WriteTimeout Duration `json:"write_timeout" yaml:"write_timeout"`
}

// WebSocketTransportConfig contains configuration for WebSocket transport.
type WebSocketTransportConfig struct {
	// Address specifies the WebSocket listen address.
	Address string `json:"address" yaml:"address"`
	
	// Port specifies the WebSocket listen port.
	Port int `json:"port" yaml:"port"`
	
	// Path specifies the WebSocket endpoint path.
	Path string `json:"path" yaml:"path"`
	
	// MaxConnections limits the number of concurrent WebSocket connections.
	MaxConnections int `json:"max_connections" yaml:"max_connections"`
	
	// MaxMessageSize limits the size of WebSocket messages.
	MaxMessageSize int64 `json:"max_message_size" yaml:"max_message_size"`
}

// RetryConfig contains retry policy configuration.
type RetryConfig struct {
	// Enabled determines whether retries are enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// MaxAttempts specifies the maximum number of retry attempts.
	MaxAttempts int `json:"max_attempts" yaml:"max_attempts"`
	
	// InitialDelay specifies the initial delay between retries.
	InitialDelay Duration `json:"initial_delay" yaml:"initial_delay"`
	
	// MaxDelay specifies the maximum delay between retries.
	MaxDelay Duration `json:"max_delay" yaml:"max_delay"`
	
	// Multiplier specifies the backoff multiplier.
	Multiplier float64 `json:"multiplier" yaml:"multiplier"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	// Level specifies the log level (debug, info, warn, error).
	Level string `json:"level" yaml:"level"`
	
	// Format specifies the log format (text, json).
	Format string `json:"format" yaml:"format"`
	
	// Output specifies the log output (stdout, stderr, file path).
	Output string `json:"output" yaml:"output"`
	
	// EnableRequestLogging enables detailed request/response logging.
	EnableRequestLogging bool `json:"enable_request_logging" yaml:"enable_request_logging"`
}

// Duration is a wrapper around time.Duration that can be marshaled to/from JSON and YAML.
type Duration time.Duration

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name:                  "StackQL MCP Server",
			Version:               "1.0.0",
			Description:           "Model Context Protocol server for StackQL",
			MaxConcurrentRequests: 100,
			RequestTimeout:        Duration(30 * time.Second),
		},
		Backend: BackendConfig{
			Type:              "stackql",
			ConnectionString:  "stackql://localhost",
			MaxConnections:    10,
			ConnectionTimeout: Duration(10 * time.Second),
			QueryTimeout:      Duration(30 * time.Second),
			Retry: RetryConfig{
				Enabled:      true,
				MaxAttempts:  3,
				InitialDelay: Duration(100 * time.Millisecond),
				MaxDelay:     Duration(5 * time.Second),
				Multiplier:   2.0,
			},
		},
		Transport: TransportConfig{
			EnabledTransports: []string{"stdio"},
			Stdio: StdioTransportConfig{
				BufferSize: 4096,
			},
			TCP: TCPTransportConfig{
				Address:        "localhost",
				Port:           8080,
				MaxConnections: 100,
				ReadTimeout:    Duration(30 * time.Second),
				WriteTimeout:   Duration(30 * time.Second),
			},
			WebSocket: WebSocketTransportConfig{
				Address:        "localhost",
				Port:           8081,
				Path:           "/mcp",
				MaxConnections: 100,
				MaxMessageSize: 1024 * 1024, // 1MB
			},
		},
		Logging: LoggingConfig{
			Level:                "info",
			Format:               "text",
			Output:               "stdout",
			EnableRequestLogging: false,
		},
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.Server.Name == "" {
		return fmt.Errorf("server.name is required")
	}
	if c.Server.Version == "" {
		return fmt.Errorf("server.version is required")
	}
	if c.Server.MaxConcurrentRequests <= 0 {
		return fmt.Errorf("server.max_concurrent_requests must be greater than 0")
	}
	if c.Backend.Type == "" {
		return fmt.Errorf("backend.type is required")
	}
	if c.Backend.MaxConnections <= 0 {
		return fmt.Errorf("backend.max_connections must be greater than 0")
	}
	if len(c.Transport.EnabledTransports) == 0 {
		return fmt.Errorf("at least one transport must be enabled")
	}
	
	// Validate enabled transports
	validTransports := map[string]bool{
		"stdio":     true,
		"tcp":       true,
		"websocket": true,
	}
	for _, transport := range c.Transport.EnabledTransports {
		if !validTransports[transport] {
			return fmt.Errorf("invalid transport: %s", transport)
		}
	}
	
	// Validate TCP config if TCP transport is enabled
	for _, transport := range c.Transport.EnabledTransports {
		if transport == "tcp" {
			if c.Transport.TCP.Port <= 0 || c.Transport.TCP.Port > 65535 {
				return fmt.Errorf("tcp.port must be between 1 and 65535")
			}
		}
		if transport == "websocket" {
			if c.Transport.WebSocket.Port <= 0 || c.Transport.WebSocket.Port > 65535 {
				return fmt.Errorf("websocket.port must be between 1 and 65535")
			}
		}
	}
	
	// Validate logging config
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}
	
	return nil
}

// LoadFromJSON loads configuration from JSON data.
func LoadFromJSON(data []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return config, nil
}

// LoadFromYAML loads configuration from YAML data.
func LoadFromYAML(data []byte) (*Config, error) {
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return config, nil
}