@linux
Feature: echo command
    Run echo command in Linux.

    Scenario: echo text
        # arrange
        Given command echo
        And arg Hello, World!
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            Hello, World!

            """
        And output not eq
            """
            Hello, World

            """
