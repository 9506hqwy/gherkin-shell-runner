@linux
Feature: stty command
    Run stty command in Linux.

    Scenario: stty size
        # arrange
        Given command stty
        And arg size
        And size 1 1
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            1 1

            """
