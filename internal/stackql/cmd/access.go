package cmd

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/stackql-labs/omnisdk/public/authentication"
	"github.com/stackql/stackql/internal/stackql/accessmanagement"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/pkg/mcp_server"
)

func configureCallerMCP(host handler.HandlerContext, cfg *mcp_server.Config) error {
	manager := host.GetAccessManager()
	if manager == nil {
		return nil
	}
	if mcpServerType == "reverse_proxy" {
		return errors.New("caller access does not support the MCP reverse proxy")
	}
	if cfg.GetServerTransport() != "http" {
		return nil
	}
	if cfg.Server.TLSCertFile == "" || cfg.Server.TLSKeyFile == "" {
		return errors.New("caller access requires HTTPS for MCP HTTP")
	}
	cfg.AuthenticateRequest = func(req *http.Request) (context.Context, error) {
		fields := strings.Fields(req.Header.Get("Authorization"))
		if req.TLS == nil || len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			return nil, errors.New("TLS bearer credential required")
		}
		evidence := accessmanagement.WithEvidence(req.Context(), authentication.Evidence{Token: fields[1], TLS: req.TLS})
		_, _, cancel, err := manager.Establish(evidence)
		if err != nil {
			return nil, err
		}
		cancel()
		return evidence, nil
	}
	return nil
}
