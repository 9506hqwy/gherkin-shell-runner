@linux
Feature: expression
    Run echo command in Linux.

    Scenario: add number
        # arrange
        Given command "echo"
        And arg "-n"
        And arg 1 + 2
        # act
        When exec
        # assert
        Then status eq 0
        And output eq 1 + 2

    Scenario: add literal
        # arrange
        Given command "echo"
        And arg "-n"
        And arg "a" + "b"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "a" + "b"

    Scenario: add variable
        # arrange
        Given command "echo"
        And set x 1
        And set y "a"
        And arg "-n"
        And arg x + y
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "1a"
