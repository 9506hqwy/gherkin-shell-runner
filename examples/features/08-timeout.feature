@linux
Feature: sleep command
    Run sleep command in Linux.

    Scenario: sleep timeout
        # arrange
        Given command sleep
        And arg 1
        And timeout 500
        # act
        When exec
        # assert
        Then status eq -1
        Then status not eq 0
        And output is empty
