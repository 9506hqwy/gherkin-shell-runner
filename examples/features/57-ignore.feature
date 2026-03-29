@windows
@ignore
Feature: exit command
    Run exit command in Windows.

    Scenario: exit
        # arrange
        Given command "cmd.exe"
        And arg "/c"
        And arg "exit 1"
        # act
        When exec
        # assert
        Then status eq 0
