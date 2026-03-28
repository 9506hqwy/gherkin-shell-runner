@linux
@ignore
Feature: exit command
    Run exit command in Linux.

    Scenario: exit
        # arrange
        Given command bash
        And arg -c
        And arg exit 1
        # act
        When exec
        # assert
        Then status eq 0
