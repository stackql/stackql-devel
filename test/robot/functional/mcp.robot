*** Settings ***
Resource          ${CURDIR}/stackql.resource

*** Test Cases *** 
MCP HTTP Server List Tools 
    Pass Execution If    "${EXECUTION_PLATFORM}" == "docker"    Skipping MCP test in docker
    ${serverProcess}=    Start Process    ${REPOSITORY_ROOT}${/}build${/}stackql
    ...                                   \-\-mcp.server.type\=http 
    ...                                   \-\-mcp.config\='{"server": {"transport": "http", "address": "127.0.0.1:9912"}}'
    ${result}=    Run Process          ${REPOSITORY_ROOT}${/}build${/}stackql_mcp_client   
    ...                  \-\-client\-type\=http 
    ...                  \-\-url\=http://127.0.0.1:9912
    ...                  stdout=${CURDIR}/tmp/MCP-HTTP-Server-List-Tools.txt
    ...                  stderr=${CURDIR}/tmp/MCP-HTTP-Server-List-Tools-stderr.txt
    Should Contain       ${result.stdout}       Get server information
    Should Be Equal As Integers    ${result.rc}    0
    [Teardown]    Terminate Process    ${serverProcess}   kill=True

