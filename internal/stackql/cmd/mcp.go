/*
Copyright © 2019 stackql info@stackql.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/stackql/any-sdk/pkg/logging"
	"github.com/stackql/stackql/internal/stackql/entryutil"
	"github.com/stackql/stackql/internal/stackql/iqlerror"
	"github.com/stackql/stackql/pkg/mcp_server"
)

//nolint:gochecknoglobals // cobra pattern
var mcpSrvCmd = &cobra.Command{
	Use:   "mcp",
	Short: "run mcp server",
	Long: `
	Run a MCP protocol server.
	Supports MCP client connections from all manner or libs.
  `,
	//nolint:revive // acceptable for now
	Run: func(cmd *cobra.Command, args []string) {
		flagErr := dependentFlagHandler(&runtimeCtx)
		iqlerror.PrintErrorAndExitOneIfError(flagErr)
		inputBundle, err := entryutil.BuildInputBundle(runtimeCtx)
		iqlerror.PrintErrorAndExitOneIfError(err)
		handlerCtx, err := entryutil.BuildHandlerContext(runtimeCtx, nil, queryCache, inputBundle, false)
		iqlerror.PrintErrorAndExitOneIfError(err)
		iqlerror.PrintErrorAndExitOneIfNil(handlerCtx, "handler context is unexpectedly nil")
		server, serverErr := mcp_server.NewExampleBackendServer(
			nil,
			logging.GetLogger(),
		)
		// server, serverErr := mcp_server.NewExampleHTTPBackendServer(
		// 	logging.GetLogger(),
		// )
		iqlerror.PrintErrorAndExitOneIfError(serverErr)
		server.Start(context.Background()) //nolint:errcheck // TODO: investigate
	},
}
