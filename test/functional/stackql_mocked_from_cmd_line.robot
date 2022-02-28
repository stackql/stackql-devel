*** Settings ***
Library    Process
Library    OperatingSystem

*** Settings ***
Variables        ${CURDIR}/variables/stackql_context.py
Suite Setup       Start Mock Server    ${JSON_INIT_FILE_PATH}    ${MOCKSERVER_JAR}    1080
Suite Teardown    Terminate All Processes

*** Test Cases *** 
Google Container Agg Desc
    Set Environment Variable    OKTA_SECRET_KEY    ${OKTA_SECRET_STR}
    ${result} =     Run StackQL Exec Command
                    ...  ${SELECT_CONTAINER_SUBNET_AGG_DESC}
    Should Be Equal    ${result.stdout}   ${SELECT_CONTAINER_SUBNET_AGG_DESC_EXPECTED}

Google Container Agg Asc
    Set Environment Variable    OKTA_SECRET_KEY    ${OKTA_SECRET_STR}
    ${result} =     Run StackQL Exec Command
                    ...  ${SELECT_CONTAINER_SUBNET_AGG_ASC}
    Should Be Equal    ${result.stdout}   ${SELECT_CONTAINER_SUBNET_AGG_ASC_EXPECTED}

*** Keywords ***
Start Mock Server
    [Arguments]    ${_JSON_INIT_FILE_PATH}    ${_MOCKSERVER_JAR}    ${_MOCKSERVER_PORT}
    ${process} =    Start Process    java    \-Dfile.encoding\=UTF-8
    ...  \-Dmockserver.initializationJsonPath\=${_JSON_INIT_FILE_PATH}
    ...  \-jar    ${_MOCKSERVER_JAR}
    ...  \-serverPort    ${_MOCKSERVER_PORT}    \-logLevel    INFO
    Sleep    5s
    [Return]    ${process}

Run StackQL Exec Command
    [Arguments]    ${_EXEC_CMD_STR}
    Set Environment Variable    OKTA_SECRET_KEY    ${OKTA_SECRET_STR}
    ${result} =     Run Process    
                    ...  ${STACKQL_EXE}
                    ...  exec    \-\-registry\=${RESISTRY_CFG_STR}
                    ...  \-\-auth\=${AUTH_CFG_STR}
                    ...  \-\-tls.allowInsecure\=true
                    ...  ${_EXEC_CMD_STR} 
    Log             ${result.stdout}
    Log             ${result.stderr}
    [Return]    ${result}
