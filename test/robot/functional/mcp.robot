*** Settings ***
Resource          ${CURDIR}${/}stackql.resource


*** Keywords ***
Start MCP HTTP Server
    Pass Execution If    "%{IS_SKIP_MCP_TEST=false}" == "true"    Some platforms do not have the MCP client available
    Start Process                         ${STACKQL_EXE}
    ...                                   mcp
    ...                                   \-\-mcp.server.type\=http
    ...                                   \-\-mcp.config
    ...                                   {"server": {"transport": "http", "address": "127.0.0.1:9912"} }
    ...                                   \-\-registry
    ...                                   ${REGISTRY_NO_VERIFY_CFG_JSON_STR}
    Sleep         5s

*** Settings ***
Suite Setup     Start MCP HTTP Server


*** Test Cases *** 
MCP HTTP Server Run List Tools
    Pass Execution If    "%{IS_SKIP_MCP_TEST=false}" == "true"    Some platforms do not have the MCP client available
    Sleep         5s
    ${result}=    Run Process          ${STACKQL_MCP_CLIENT_EXE}
    ...                  exec
    ...                  \-\-client\-type\=http 
    ...                  \-\-url\=http://127.0.0.1:9912
    ...                  stdout=${CURDIR}${/}tmp${/}MCP-HTTP-Server-Run-List-Tools.txt
    ...                  stderr=${CURDIR}${/}tmp${/}MCP-HTTP-Server-Run-List-Tools-stderr.txt
    Should Contain       ${result.stdout}       Get server information
    Should Be Equal As Integers    ${result.rc}    0


MCP HTTP Server Verify Greeting Tool
    Pass Execution If    "%{IS_SKIP_MCP_TEST=false}" == "true"    Some platforms do not have the MCP client available
    Sleep         5s
    ${result}=    Run Process          ${STACKQL_MCP_CLIENT_EXE}
    ...                  exec
    ...                  \-\-client\-type\=http 
    ...                  \-\-url\=http://127.0.0.1:9912
    ...                  \-\-exec.action      greet 
    ...                  \-\-exec.args        {"name": "JOE BLOW"}
    ...                  stdout=${CURDIR}${/}tmp${/}MCP-HTTP-Server-Verify-Greeting-Tool.txt
    ...                  stderr=${CURDIR}${/}tmp${/}MCP-HTTP-Server-Verify-Greeting-Tool-stderr.txt
    Should Contain       ${result.stdout}       JOE BLOW
    Should Be Equal As Integers    ${result.rc}    0


MCP HTTP Server List Providers Tool
    Pass Execution If    "%{IS_SKIP_MCP_TEST=false}" == "true"    Some platforms do not have the MCP client available
    Sleep         5s
    ${result}=    Run Process          ${STACKQL_MCP_CLIENT_EXE}
    ...                  exec
    ...                  \-\-client\-type\=http 
    ...                  \-\-url\=http://127.0.0.1:9912
    ...                  \-\-exec.action      list_providers 
    ...                  stdout=${CURDIR}${/}tmp${/}MCP-HTTP-Server-List-Providers.txt
    ...                  stderr=${CURDIR}${/}tmp${/}MCP-HTTP-Server-List-Providers-stderr.txt
    Should Contain       ${result.stdout}       local_openssl
    Should Be Equal As Integers    ${result.rc}    0

