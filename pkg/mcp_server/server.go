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

func sayHi(ctx context.Context, req *mcp.CallToolRequest, input GreetingInput) (
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

// // handlemcpRequest processes an MCP request and returns a response.
// func (s *mcpServer) handlemcpRequest(ctx context.Context, req *mcpRequest) *mcpResponse {
// 	// Acquire semaphore for concurrency control
// 	if err := s.requestSemaphore.Acquire(ctx, 1); err != nil {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32603,
// 				Message: "Server overloaded",
// 			},
// 		}
// 	}
// 	defer s.requestSemaphore.Release(1)

// 	// Set request timeout
// 	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.Server.RequestTimeout))
// 	defer cancel()

// 	switch req.Method {
// 	case "initialize":
// 		return s.handleInitialize(reqCtx, req)
// 	case "resources/list":
// 		return s.handleResourcesList(reqCtx, req)
// 	case "resources/read":
// 		return s.handleResourcesRead(reqCtx, req)
// 	case "tools/list":
// 		return s.handleToolsList(reqCtx, req)
// 	case "tools/call":
// 		return s.handleToolsCall(reqCtx, req)
// 	default:
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32601,
// 				Message: fmt.Sprintf("Method not found: %s", req.Method),
// 			},
// 		}
// 	}
// }

// // handleInitialize handles the MCP initialize request.
// func (s *simpleMCPServer) handleInitialize(ctx context.Context, req *mcpRequest) *mcpResponse {
// 	initResult := map[string]interface{}{
// 		"protocolVersion": "2024-11-05",
// 		"serverInfo": map[string]interface{}{
// 			"name":    s.config.Server.Name,
// 			"version": s.config.Server.Version,
// 		},
// 		"capabilities": map[string]interface{}{
// 			"resources": map[string]interface{}{
// 				"subscribe": true,
// 			},
// 			"tools": map[string]interface{}{},
// 		},
// 	}

// 	return &mcpResponse{
// 		JSONRPC: "2.0",
// 		ID:      req.ID,
// 		Result:  initResult,
// 	}
// }

// // handleResourcesList handles the MCP resources/list request.
// func (s *simpleMCPServer) handleResourcesList(ctx context.Context, req *mcpRequest) *mcpResponse {
// 	schema, err := s.backend.GetSchema(ctx)
// 	if err != nil {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32603,
// 				Message: fmt.Sprintf("Failed to get schema: %v", err),
// 			},
// 		}
// 	}

// 	var resources []map[string]interface{}

// 	// Convert schema to MCP resources format
// 	for _, provider := range schema.Providers {
// 		for _, service := range provider.Services {
// 			for _, resource := range service.Resources {
// 				mcpResource := map[string]interface{}{
// 					"uri":         fmt.Sprintf("stackql://%s/%s/%s", provider.Name, service.Name, resource.Name),
// 					"name":        fmt.Sprintf("%s.%s.%s", provider.Name, service.Name, resource.Name),
// 					"description": fmt.Sprintf("StackQL resource: %s.%s.%s", provider.Name, service.Name, resource.Name),
// 					"mimeType":    "application/json",
// 				}
// 				resources = append(resources, mcpResource)
// 			}
// 		}
// 	}

// 	return &mcpResponse{
// 		JSONRPC: "2.0",
// 		ID:      req.ID,
// 		Result: map[string]interface{}{
// 			"resources": resources,
// 		},
// 	}
// }

// // handleResourcesRead handles the MCP resources/read request.
// func (s *simpleMCPServer) handleResourcesRead(ctx context.Context, req *mcpRequest) *mcpResponse {
// 	var params struct {
// 		URI string `json:"uri"`
// 	}

// 	if err := json.Unmarshal(req.Params, &params); err != nil {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32602,
// 				Message: "Invalid parameters",
// 			},
// 		}
// 	}

// 	// For now, return resource metadata
// 	// In a full implementation, this would return actual resource data
// 	resourceContent := map[string]interface{}{
// 		"uri":      params.URI,
// 		"mimeType": "application/json",
// 		"text":     fmt.Sprintf(`{"message": "Resource data for %s would be returned here"}`, params.URI),
// 	}

// 	return &mcpResponse{
// 		JSONRPC: "2.0",
// 		ID:      req.ID,
// 		Result: map[string]interface{}{
// 			"contents": []interface{}{resourceContent},
// 		},
// 	}
// }

// // handleToolsList handles the MCP tools/list request.
// func (s *simpleMCPServer) handleToolsList(ctx context.Context, req *mcpRequest) *mcpResponse {
// 	tools := []map[string]interface{}{
// 		{
// 			"name":        "stackql_query",
// 			"description": "Execute StackQL queries against cloud provider APIs",
// 			"inputSchema": map[string]interface{}{
// 				"type": "object",
// 				"properties": map[string]interface{}{
// 					"query": map[string]interface{}{
// 						"type":        "string",
// 						"description": "The StackQL query to execute",
// 					},
// 					"parameters": map[string]interface{}{
// 						"type":        "object",
// 						"description": "Optional parameters for the query",
// 					},
// 				},
// 				"required": []string{"query"},
// 			},
// 		},
// 	}

// 	return &mcpResponse{
// 		JSONRPC: "2.0",
// 		ID:      req.ID,
// 		Result: map[string]interface{}{
// 			"tools": tools,
// 		},
// 	}
// }

// // handleToolsCall handles the MCP tools/call request.
// func (s *simpleMCPServer) handleToolsCall(ctx context.Context, req *mcpRequest) *mcpResponse {
// 	var params struct {
// 		Name      string                 `json:"name"`
// 		Arguments map[string]interface{} `json:"arguments"`
// 	}

// 	if err := json.Unmarshal(req.Params, &params); err != nil {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32602,
// 				Message: "Invalid parameters",
// 			},
// 		}
// 	}

// 	if params.Name != "stackql_query" {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32601,
// 				Message: fmt.Sprintf("Unknown tool: %s", params.Name),
// 			},
// 		}
// 	}

// 	query, ok := params.Arguments["query"].(string)
// 	if !ok {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32602,
// 				Message: "Query parameter is required and must be a string",
// 			},
// 		}
// 	}

// 	queryParams, _ := params.Arguments["parameters"].(map[string]interface{})

// 	result, err := s.backend.Execute(ctx, query, queryParams)
// 	if err != nil {
// 		return &mcpResponse{
// 			JSONRPC: "2.0",
// 			ID:      req.ID,
// 			Error: &mcpError{
// 				Code:    -32603,
// 				Message: fmt.Sprintf("Query execution failed: %v", err),
// 			},
// 		}
// 	}

// 	return &mcpResponse{
// 		JSONRPC: "2.0",
// 		ID:      req.ID,
// 		Result: map[string]interface{}{
// 			"content": []interface{}{
// 				map[string]interface{}{
// 					"type": "text",
// 					"text": fmt.Sprintf("Query executed successfully. Rows affected: %d, Execution time: %dms",
// 						result.RowsAffected, result.ExecutionTime),
// 				},
// 				map[string]interface{}{
// 					"type": "text",
// 					"text": fmt.Sprintf("Result: %+v", result),
// 				},
// 			},
// 			"isError": false,
// 		},
// 	}
// }

// // startStdioTransport starts the stdio transport (placeholder implementation).
// func (s *simpleMCPServer) startStdioTransport(ctx context.Context) error {
// 	s.logger.Printf("Stdio transport started (placeholder implementation)")
// 	// In a real implementation, this would handle stdio JSON-RPC communication
// 	return nil
// }

// // startTCPTransport starts the TCP transport.
// func (s *simpleMCPServer) startTCPTransport(ctx context.Context) error {
// 	addr := fmt.Sprintf("%s:%d", s.config.Transport.TCP.Address, s.config.Transport.TCP.Port)

// 	router := mux.NewRouter()
// 	router.HandleFunc("/mcp", s.handleHTTPMCP).Methods("POST")

// 	server := &http.Server{
// 		Addr:         addr,
// 		Handler:      router,
// 		ReadTimeout:  time.Duration(s.config.Transport.TCP.ReadTimeout),
// 		WriteTimeout: time.Duration(s.config.Transport.TCP.WriteTimeout),
// 	}

// 	listener, err := net.Listen("tcp", addr)
// 	if err != nil {
// 		return fmt.Errorf("failed to listen on %s: %w", addr, err)
// 	}

// 	go func() {
// 		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
// 			s.logger.Printf("TCP server error: %v", err)
// 		}
// 	}()

// 	s.servers = append(s.servers, server)
// 	s.logger.Printf("TCP transport started on %s", addr)
// 	return nil
// }

// // startWebSocketTransport starts the WebSocket transport (placeholder implementation).
// func (s *simpleMCPServer) startWebSocketTransport(ctx context.Context) error {
// 	addr := fmt.Sprintf("%s:%d", s.config.Transport.WebSocket.Address, s.config.Transport.WebSocket.Port)
// 	s.logger.Printf("WebSocket transport started on %s%s (placeholder implementation)", addr, s.config.Transport.WebSocket.Path)
// 	// In a real implementation, this would handle WebSocket connections
// 	return nil
// }

// // handleHTTPMCP handles HTTP-based MCP requests.
// func (s *simpleMCPServer) handleHTTPMCP(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var req mcpRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	resp := s.handleMCPRequest(r.Context(), &req)

// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(resp); err != nil {
// 		s.logger.Printf("Failed to encode response: %v", err)
// 	}
// }
