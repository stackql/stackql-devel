package mcp_server //nolint:revive // fine for now

// create an http client that can talk to the mcp server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

const (
	MCPClientTypeHTTP  = "http"
	MCPClientTypeSTDIO = "stdio"
)

type MCPClient interface {
	InspectTools() ([]map[string]any, error)
	CallToolText(toolName string, args map[string]any) (string, error)
}

func NewMCPClient(clientType string, baseURL string, clientCfgMap map[string]any, logger *logrus.Logger) (MCPClient, error) {
	switch clientType {
	case MCPClientTypeHTTP:
		return newHTTPMCPClient(baseURL, clientCfgMap, logger)
	case MCPClientTypeSTDIO:
		return newStdioMCPClient(logger)
	default:
		return nil, fmt.Errorf("unknown client type: %s", clientType)
	}
}

func getHTTPClient(clientCfgMap map[string]any) (*http.Client, error) {
	if clientCfgMap != nil && clientCfgMap["ca_file"] != nil {
		caFile, isString := clientCfgMap["ca_file"].(string)
		if !isString {
			return nil, fmt.Errorf("ca_file must be a string")
		}
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Create a TLS configuration
		tlsConfig := &tls.Config{
			RootCAs:    caCertPool,       // Trust custom CA certificates
			MinVersion: tls.VersionTLS12, // Enforce minimum TLS version
			// InsecureSkipVerify: false,            // Set to true to skip server certificate verification (NOT recommended for production)
		}

		// Create a custom HTTP transport
		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		return &http.Client{Transport: tr}, nil
	}
	return http.DefaultClient, nil
}

func newHTTPMCPClient(baseURL string, clientCfgMap map[string]any, logger *logrus.Logger) (MCPClient, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}
	httpClient, httpClientErr := getHTTPClient(clientCfgMap)
	if httpClientErr != nil {
		return nil, fmt.Errorf("error creating HTTP client: %w", httpClientErr)
	}
	return &httpMCPClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
		clientCfg:  clientCfgMap,
	}, nil
}

type httpMCPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	clientCfg  map[string]any
}

func (c *httpMCPClient) connect() (*mcp.ClientSession, error) {
	url := c.baseURL
	ctx := context.Background()

	// Create the URL for the server.
	c.logger.Infof("Connecting to MCP server at %s", url)

	// Create an MCP client.
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "stackql-client",
		Version: "1.0.0",
	}, nil)

	// Connect to the server.
	return client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: url}, nil)
}

func (c *httpMCPClient) connectOrDie() *mcp.ClientSession {
	session, err := c.connect()
	if err != nil {
		c.logger.Fatalf("Failed to connect: %v", err)
	}
	return session
}

func (c *httpMCPClient) InspectTools() ([]map[string]any, error) {
	session := c.connectOrDie()
	defer session.Close()

	c.logger.Infof("Connected to server (session ID: %s)", session.ID())

	// First, list available tools.
	c.logger.Infof("Listing available tools...")
	toolsResult, err := session.ListTools(context.Background(), nil)
	if err != nil {
		c.logger.Fatalf("Failed to list tools: %v", err)
	}
	var rv []map[string]any
	for _, tool := range toolsResult.Tools {
		c.logger.Infof("  - %s: %s\n", tool.Name, tool.Description)
		toolInfo := map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
		}
		rv = append(rv, toolInfo)
	}

	c.logger.Infof("Client completed successfully")
	return rv, nil
}

func (c *httpMCPClient) callTool(toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	session := c.connectOrDie()
	defer session.Close()

	c.logger.Infof("Connected to server (session ID: %s)", session.ID())

	c.logger.Infof("Calling tool %s...", toolName)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		c.logger.Errorf("Failed to call tool %s: %v\n", toolName, err)
		return result, err
	}

	c.logger.Infof("Client completed successfully")
	return result, nil
}

func (c *httpMCPClient) CallToolText(toolName string, args map[string]any) (string, error) {
	toolCall, toolCallErr := c.callTool(toolName, args)
	if toolCallErr != nil {
		return "", toolCallErr
	}
	var result string
	for _, content := range toolCall.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			result += textContent.Text + "\n"
		}
	}
	return result, nil
}

type stdioMCPClient struct {
	logger *logrus.Logger
}

func newStdioMCPClient(logger *logrus.Logger) (MCPClient, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}
	return &stdioMCPClient{
		logger: logger,
	}, nil
}

func (c *stdioMCPClient) InspectTools() ([]map[string]any, error) {
	c.logger.Infof("stdio MCP client not implemented yet")
	return nil, nil
}

func (c *stdioMCPClient) CallToolText(toolName string, args map[string]any) (string, error) {
	c.logger.Infof("stdio MCP client not implemented yet")
	return "", nil
}
