*** Settings ***
Library    Process
Library    OperatingSystem


*** Settings ***
Variables        ${CURDIR}/variables/stackql_context.py
Test Setup       Start Mock Server
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
    Start Process    java    \-Dfile.encoding\=UTF-8
    ...  \-Dmockserver.initializationJsonPath\=${REPOSITORY_ROOT}/test/server/expectations/static-gcp-expectations.json
    ...  \-jar    /usr/local/lib/mockserver/mockserver-netty-jar-with-dependencies.jar
    ...  \-serverPort    1080    \-logLevel    INFO
    Sleep    5s
