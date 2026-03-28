@windows
Feature: environment variables
    Run echo command in Windows.

    Scenario: echo env
        # arrange
        Given command cmd.exe
        And env KEY VALUE
        And arg /c
        And arg echo %KEY%
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            VALUE

            """
