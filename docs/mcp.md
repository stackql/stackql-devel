
## Running the MCP server

If necessary, rebuild stackql with:

```bash
python cicd/python/build.py --build
```

**Note**: before starting an MCP server, remember to export all appropriate auth env vars.

We have a nice debug config for running an MCP server with `vscode`, please see [the `vscode` debug launch config](/.vscode/launch.json) for that.  Otherwise, you can run with stackql (assuming locally built into `./build/stackql`):


```bash

./build/stackql mcp --mcp.server.type=http --mcp.config '{"server": {"transport": "http", "address": "127.0.0.1:9992"} }'


```


## Using the MCP Client

This is very much a development tool, not currently recommended for production.

Build:

```bash
python cicd/python/build.py --build-mcp-client
```

Then, assuming you have a `stackql` MCP server serving streamable HTTP on port `9992`: 


```bash

./build/stackql_mcp_client exec --client-type=http  --url=http://127.0.0.1:9992 --exec.action      list_providers

./build/stackql_mcp_client exec --client-type=http  --url=http://127.0.0.1:9992 --exec.action      list_services --exec.args '{"provider": "google"}'

./build/stackql_mcp_client exec --client-type=http  --url=http://127.0.0.1:9992 --exec.action      list_resources --exec.args '{"provider": "google", "service": "compute"}'


./build/stackql_mcp_client exec --client-type=http  --url=http://127.0.0.1:9992 --exec.action      list_methods --exec.args '{"provider": "google", "service": "compute", "resource": "networks"}'

./build/stackql_mcp_client exec --client-type=http  --url=http://127.0.0.1:9992 --exec.action query_json_v2      --exec.args '{"sql": "select name from google.compute.networks where project = '"'"'stackql-demo'"'"';"}'

```
