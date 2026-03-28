@linux
Feature: environment variables
    Run echo command in Linux.

    Scenario: echo env
        # arrange
        Given command bash
        And env KEY VALUE
        And arg -c
        And arg echo -n $KEY
        # act
        When exec
        # assert
        Then status eq 0
        And output is not empty
        And output eq VALUE
