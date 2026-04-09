@linux
Feature: echo command
    Run echo command in Linux.

    Scenario Outline: echo text
        # arrange
        Given command "bash"
        And env
            | KEY1 | "value1" |
            | KEY2 | "value2" |
        And arg "-c"
        And arg "echo -n $KEY1 $KEY2"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "value1 value2"
