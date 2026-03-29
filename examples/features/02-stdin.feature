@linux
Feature: cat command
    Run cat command in Linux.

    Scenario: cat text
        # arrange
        Given command "cat"
        And stdin
            """
            text1
            text2
            """
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            text1
            text2
            """
