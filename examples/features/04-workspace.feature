@linux
Feature: workspace
    Run pwd command in Linux.

    Scenario: show current directory
        # arrange
        Given command pwd
        And workspace /tmp
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            /tmp

            """
