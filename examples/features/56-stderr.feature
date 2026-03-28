@windows
Feature: echo command
    Run echo command in Windows.

    Scenario: echo text
        # arrange
        Given command cmd.exe
        And arg /c
        And arg echo error 1>&2
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            error

            """
