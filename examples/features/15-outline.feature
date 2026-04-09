@linux
Feature: echo command
    Run echo command in Linux.

    Scenario Outline: echo text
        # arrange
        Given command "echo"
        And arg "-n"
        And arg "Hello, <value>"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "Hello, <value>"

        Examples:
            | value  |
            | World! |
            | 世界   |
