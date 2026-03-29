@linux
Feature: exit code
    Run exit command in Linux.

    Scenario: exit 1
        # arrange
        Given command "bash"
        And arg "-c"
        And arg "exit 1"
        # act
        When exec
        # assert
        Then status eq 1
        And output is empty
