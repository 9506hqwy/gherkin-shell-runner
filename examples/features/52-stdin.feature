@windows
Feature: sort command
    Run sort command in Windows.

    Scenario: sort text
        # arrange
        Given command cmd.exe
        And arg /c
        And arg sort
        And stdin
            """
            text2
            text1
            """
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            text2
            text1^Z
            text1
            text2

            """
