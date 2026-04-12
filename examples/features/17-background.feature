@linux
Feature: echo command
    Run echo command in Linux.

    Background:
        # set environment variable.
        * env KEY "value"

    Scenario: echo text
        # arrange
        Given command "bash"
        And arg "-c"
        And arg "echo -n $KEY"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "value"

    Scenario: echo text
        # arrange
        Given command "bash"
        And arg "-c"
        And arg "echo -n $KEY"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "value"
