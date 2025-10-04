package mcp_server

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
		logger.SetOutput(io.Discard)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "greeter", Version: "v1.0.0"},
		nil,
	)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, sayHi)

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
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}
	s.server.Run(ctx, &mcp.StdioTransport{})

	s.running = true
	s.logger.Printf("MCP server started with transports: %v", s.config.Transport.EnabledTransports)
	return nil
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
