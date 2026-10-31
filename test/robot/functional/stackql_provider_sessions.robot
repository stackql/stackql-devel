*** Settings ***
Documentation     Explicit provider sessions, credential rotation and UNSET through the CLI against loopback mocks.
Library           ProviderSessionLibrary.py
Suite Setup       Prepare Session Provider    ${STACKQL_EXE}
Suite Teardown    Stop Session Provider

*** Variables ***
${STACKQL_EXE}    ${EXECDIR}/build/stackql

*** Test Cases ***
Provider Aliases Keep Their Credentials And Scope Across Session Edits
    Provider Aliases Survive Session Edits

Persistent Secrets Rotate Without Recreating The Session
    Secret Credentials Rotate And Reset

Auth Dto Fields Can Be Set And Unset In Memory
    Auth Dto Changes And Unsets

Unset Alias Auth Fails Before Any Provider Request
    Unset Alias Auth Fails Before Request
