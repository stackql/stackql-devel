*** Settings ***
Resource          ${CURDIR}/stackql.resource

*** Test Cases *** 
MCP HTTP Server List Tools
    ${serverProcess}=    Start Process    ${STACKQL_EXE}
    ...                                   mcp
    ...                                   \-\-mcp.server.type\=http 
    ...                                   \-\-mcp.config
    ...                                   {"server": {"transport": "http", "address": "127.0.0.1:9912"} }
    Sleep         5s
    ${result}=    Run Process          ${STACKQL_MCP_CLIENT_EXE}
    ...                  exec
    ...                  \-\-client\-type\=http 
    ...                  \-\-url\=http://127.0.0.1:9912
    ...                  stdout=${CURDIR}/tmp/MCP-HTTP-Server-List-Tools.txt
    ...                  stderr=${CURDIR}/tmp/MCP-HTTP-Server-List-Tools-stderr.txt
    Should Contain       ${result.stdout}       Get server information
    Should Be Equal As Integers    ${result.rc}    0
    [Teardown]    Terminate Process    ${serverProcess}   kill=True

