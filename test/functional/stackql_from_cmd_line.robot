*** Settings ***
Library    Process    

*** Settings ***
Variables    ${CURDIR}/variables/stackql_context.py

*** Test Cases *** 
Get StackQL Help
    ${result} =     Run Process     ${STACKQL_EXE}    --help 
    Log    ${result.stdout}
    Should contain    ${result.stdout}   stackql${SPACE}\[command\]


Get StackQL Providers
    ${result} =     Run Process    ${STACKQL_EXE}    exec    \-\-registry\=${RESISTRY_CFG_STR}    ${SHOW_PROVIDERS_STR} 
    Log    ${result.stdout}
    Should contain    ${result.stdout}   okta


