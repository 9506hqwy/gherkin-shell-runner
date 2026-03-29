@linux
Feature: echo command
    Run echo command in Linux.

    Scenario: echo text
        # arrange
        Given command "bash"
        And arg "-c"
        And arg "echo 1>&2 error"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            error

            """
