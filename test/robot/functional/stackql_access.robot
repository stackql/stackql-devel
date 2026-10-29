*** Settings ***
Resource    ${CURDIR}/stackql.resource
Library     stackql_test_tooling.access_management_keywords.AccessManagementKeywords

*** Test Cases ***
Caller Access Uses Configured RDBMS Directory And Denies Unauthorized Effects
    [Documentation]    Signed JWTs exercise the real CLI and SQLite/PostgreSQL directory against a local provider API. Tokens remain outside arguments and Robot artifacts.
    Verify Caller Access    ${STACKQL_EXE}    ${SQL_BACKEND_CFG_STR_CANONICAL}    ${EXECUTION_PLATFORM}
