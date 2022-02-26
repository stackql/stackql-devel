*** Settings ***
Library    Process     

*** Test Cases *** 
Get StackQL Help
    ${result} =     Run Process     ../../build/stackql    --help 
    Log    ${result.stdout}
    Should contain    ${result.stdout}   stackql${SPACE}\[command\]


