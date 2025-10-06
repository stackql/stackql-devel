package mcp_server //nolint:revive // fine for now

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverTransportStdIO = "stdio"
	serverTransportHTTP  = "http"
	serverTransportSSE   = "sse"
	DefaultHTTPServerURL = "http://127.0.0.1:9876"
)

type MCPServer interface {
	Start(context.Context) error
	Stop() error
}

// simpleMCPServer implements the Model Context Protocol server for StackQL.
type simpleMCPServer struct {
	config  *Config
	backend Backend
	logger  *logrus.Logger

	server *mcp.Server

	// Concurrency control
	requestSemaphore *semaphore.Weighted

	// Server state
	mu      sync.RWMutex
	running bool
	servers []io.Closer // Track all running servers for cleanup
}

func sayHi(_ context.Context, _ *mcp.CallToolRequest, input GreetingInput) (
	*mcp.CallToolResult,
	GreetingOutput,
	error,
) {
	return nil, GreetingOutput{Greeting: "Hi " + input.Name}, nil
}

func (s *simpleMCPServer) runHTTPServer(server *mcp.Server, url string) error {
	// Create the streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	handlerWithLogging := loggingHandler(handler, s.logger)

	s.logger.Debugf("MCP server listening on %s", url)
	s.logger.Debugf("Available tool: cityTime (cities: nyc, sf, boston)")

	// Start the HTTP server with logging handler.
	if err := http.ListenAndServe(url, handlerWithLogging); err != nil {
		s.logger.Errorf("Server failed: %v", err)
		return err
	}
	return nil
}

func NewExampleBackendServer(config *Config, logger *logrus.Logger) (MCPServer, error) {
	backend := NewExampleBackend("example-connection-string")
	return NewMCPServer(config, backend, logger)
}

func NewExampleHTTPBackendServer(logger *logrus.Logger) (MCPServer, error) {
	backend := NewExampleBackend("example-connection-string")
	config := DefaultHTTPConfig()
	return NewMCPServer(config, backend, logger)
}

// NewMCPServer creates a new MCP server with the provided configuration and backend.
func NewMCPServer(config *Config, backend Backend, logger *logrus.Logger) (MCPServer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	if backend == nil {
		return nil, fmt.Errorf("backend is required")
	}
	if logger == nil {
		logger = logrus.New()
		// logger.SetOutput(io.Discard)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "greeter", Version: "v0.1.0"},
		nil,
	)
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "greet",
			Description: "Say hi.  A simple livenesss check",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, args greetInput) (*mcp.CallToolResult, any, error) {
			greeting, greetingErr := backend.Greet(ctx, args)
			if greetingErr != nil {
				return nil, nil, greetingErr
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: greeting},
				},
			}, nil, nil
		},
	)
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "server_info",
			Description: "Get server information",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, args greetInput) (*mcp.CallToolResult, serverInfoOutput, error) {
			rv, rvErr := backend.ServerInfo(ctx, args)
			if rvErr != nil {
				return nil, serverInfoOutput{}, rvErr
			}
			return nil, rv, nil
		},
	)
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "db_identity",
			Description: "get current database identity",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, args greetInput) (*mcp.CallToolResult, map[string]any, error) {
			rv, rvErr := backend.DBIdentity(ctx, args)
			if rvErr != nil {
				return nil, nil, rvErr
			}
			return nil, rv, nil
		},
	)
	// mcp.AddTool(
	// 	server,
	// 	&mcp.Tool{
	// 		Name:        "query",
	// 		Description: "execute a SQL query",
	// 		// Input and output schemas can be defined here if needed.
	// 	},
	// 	func(ctx context.Context, req *mcp.CallToolRequest, args queryInput) (*mcp.CallToolResult, any, error) {
	// 		rv, rvErr := backend.RunQuery(ctx, args)
	// 		if rvErr != nil {
	// 			return nil, nil, rvErr
	// 		}
	// 		return &mcp.CallToolResult{
	// 			Content: []mcp.Content{
	// 				&mcp.TextContent{Text: rv},
	// 			},
	// 		}, nil, nil
	// 	},
	// )
	// mcp.AddTool(
	// 	server,
	// 	&mcp.Tool{
	// 		Name:        "query_json",
	// 		Description: "execute a SQL query and return a JSON array of rows",
	// 		// Input and output schemas can be defined here if needed.
	// 	},
	// 	func(ctx context.Context, req *mcp.CallToolRequest, args queryJSONInput) (*mcp.CallToolResult, any, error) {
	// 		arr, err := backend.RunQueryJSON(ctx, args)
	// 		if err != nil {
	// 			return nil, nil, err
	// 		}
	// 		bytesArr, marshalErr := json.Marshal(arr)
	// 		if marshalErr != nil {
	// 			return nil, nil, fmt.Errorf("failed to marshal query result to JSON: %w", marshalErr)
	// 		}
	// 		return &mcp.CallToolResult{
	// 			Content: []mcp.Content{
	// 				&mcp.TextContent{Text: string(bytesArr)},
	// 			},
	// 		}, nil, nil
	// 	},
	// )

	return &simpleMCPServer{
		config:           config,
		backend:          backend,
		logger:           logger,
		server:           server,
		requestSemaphore: semaphore.NewWeighted(int64(config.Server.MaxConcurrentRequests)),
		servers:          make([]io.Closer, 0),
	}, nil
}

// Start starts the MCP server with all configured transports.
//
//nolint:errcheck // ok for now
func (s *simpleMCPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		s.running = false
	}()
	if s.running {
		return fmt.Errorf("server is already running")
	}
	s.running = true
	return s.run(ctx)
}

// Synchronous server run.
func (s *simpleMCPServer) run(ctx context.Context) error {
	switch s.config.Server.Transport {
	case serverTransportHTTP:
		return s.runHTTPServer(s.server, s.config.Server.URL)
	case serverTransportSSE:
		return fmt.Errorf("SSE transport not yet implemented")
	case serverTransportStdIO:
		// Default to stdio transport
		return s.server.Run(ctx, &mcp.StdioTransport{})
	default:
		return fmt.Errorf("unsupported transport: %s", s.config.Server.Transport)
	}
}

// Stop gracefully stops the MCP server and all transports.
func (s *simpleMCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Close all servers
	var errs []error
	for _, server := range s.servers {
		if err := server.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close backend
	if err := s.backend.Close(); err != nil {
		errs = append(errs, err)
	}

	s.running = false
	s.servers = s.servers[:0]

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	s.logger.Printf("MCP server stopped")
	return nil
}
