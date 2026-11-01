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

Explicit Session Scope Supports Set And Unset
    Explicit Session Scope

Global Configuration Scope Fails Before Provider Execution
    Unsupported Configuration Scope

Dollar Is Rejected On Assignment Targets
    Dollar Assignment Target Is Rejected

Scalar Variables Preserve Functions Operators And Types
    Scalar Variables Preserve Sql Expressions

Variable Predicates Resolve Before Provider Dispatch
    Variable Predicate Resolves Before Provider Dispatch

Missing Unset And Secret Variables Fail Before Dispatch
    Variable Errors Prevent Provider Dispatch

Secret Dereferencing Requires Trusted Runtime Opt In
    Secret Variables Require Runtime Opt In
