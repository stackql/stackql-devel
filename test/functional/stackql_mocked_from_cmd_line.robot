*** Settings ***
Library    Process
Library    OperatingSystem

*** Settings ***
Variables        ${CURDIR}/variables/stackql_context.py
Test Setup       Start Mock Server    ${JSON_INIT_FILE_PATH}    ${MOCKSERVER_JAR}    1080
Test Teardown    Terminate All Processes

*** Test Cases *** 
Get Okta Application Resources
    Set Environment Variable    OKTA_SECRET_KEY    ${OKTA_SECRET_STR}
    ${result} =     Run Process    ${STACKQL_EXE}    exec    \-\-registry\=${RESISTRY_CFG_STR}    \-\-auth\=${AUTH_CFG_STR}    \-\-tls.allowInsecure\=true    ${SELECT_CONTAINER_SUBNET_AGG} 
    Log             ${result.stdout}
    Log             ${result.stderr}
    Should contain    ${result.stdout}   ip

*** Keywords ***
Start Mock Server
    [Arguments]    ${_JSON_INIT_FILE_PATH}    ${_MOCKSERVER_JAR}    ${_MOCKSERVER_PORT}
    ${process} =    Start Process    java    \-Dfile.encoding\=UTF-8
    ...  \-Dmockserver.initializationJsonPath\=${_JSON_INIT_FILE_PATH}
    ...  \-jar    ${_MOCKSERVER_JAR}
    ...  \-serverPort    ${_MOCKSERVER_PORT}    \-logLevel    INFO
    Sleep    5s
    [Return]    ${process}
