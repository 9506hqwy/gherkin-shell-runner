@windows
Feature: timeout command
    Run timeout command in Windows.

    Scenario: timeout
        # arrange
        Given command cmd.exe
        And arg /c
        And arg timeout /t 1
        And timeout 500
        # act
        When exec
        # assert
        Then status eq 1
        Then status not eq 0
        And output is not empty
