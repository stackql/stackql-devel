
## Running the MCP server


```bash




```


## Using the MCP Client

This is very much a development tool, not currently recommended for production.

Build:

```bash
python cicd/python/build.py --build-mcp-client
```

Then, assuming you have a `stackql` MCP server serving streamable HTTP on port `9992`: 


```bash

./build/stackql_mcp_client exec --client-type=http  --url=http://127.0.0.1:9992 --exec.action      list_services --exec.args '{"provider": "google"}'

```
