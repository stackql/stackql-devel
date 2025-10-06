package mcp_server //nolint:revive // fine for now

// create an http client that can talk to the mcp server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

const (
	MCPClientTypeHTTP  = "http"
	MCPClientTypeSTDIO = "stdio"
)

type MCPClient interface {
	InspectTools() ([]map[string]any, error)
}

func NewMCPClient(clientType string, baseURL string, logger *logrus.Logger) (MCPClient, error) {
	switch clientType {
	case MCPClientTypeHTTP:
		return newHTTPMCPClient(baseURL, logger)
	case MCPClientTypeSTDIO:
		return newStdioMCPClient(logger)
	default:
		return nil, fmt.Errorf("unknown client type: %s", clientType)
	}
}

func newHTTPMCPClient(baseURL string, logger *logrus.Logger) (MCPClient, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}
	return &httpMCPClient{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		logger:     logger,
	}, nil
}

type httpMCPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

func (c *httpMCPClient) InspectTools() ([]map[string]any, error) {
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
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: url}, nil)
	if err != nil {
		c.logger.Fatalf("Failed to connect: %v", err)
	}
	defer session.Close()

	c.logger.Infof("Connected to server (session ID: %s)", session.ID())

	// First, list available tools.
	c.logger.Infof("Listing available tools...")
	toolsResult, err := session.ListTools(ctx, nil)
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

	// Call the cityTime tool for each city.
	cities := []string{"nyc", "sf", "boston"}

	c.logger.Println("Getting time for each city...")
	for _, city := range cities {
		// Call the tool.
		result, resultErr := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "cityTime",
			Arguments: map[string]any{
				"city": city,
			},
		})
		if resultErr != nil {
			c.logger.Infof("Failed to get time for %s: %v\n", city, resultErr)
			continue
		}

		// Print the result.
		for _, content := range result.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				c.logger.Infof("  %s", textContent.Text)
			}
		}
	}

	c.logger.Infof("Client completed successfully")
	return rv, nil
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
