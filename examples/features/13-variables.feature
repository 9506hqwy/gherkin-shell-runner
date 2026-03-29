@linux
Feature: variables
    Run echo command in Linux.

    Scenario: arg and output
        # arrange
        Given command "echo"
        And set value 1
        And arg "-n"
        And arg value
        # act
        When exec
        # assert
        Then status eq 0
        And output eq value
        And output eq
            """
            1
            """

        # arrange
        Given command "echo"
        And arg "-n"
        And arg value
        # act
        When exec
        # assert
        Then status eq 0
        And output eq value

    Scenario: env and output
        # arrange
        Given command "bash"
        And set value "abc"
        And env key value
        And arg "-c"
        And arg "echo ${key}"
        # act
        When exec
        # assert
        Then status eq 0
        And output regex value
        And output regex
            """
            abc
            """
