@linux
Feature: newline
    Run echo command in Linux.

    Scenario: echo crlf
        # arrange
        Given command "echo"
        And arg "-n"
        And arg "-e"
        And arg "Hello,\r\nWorld!"
        And newline output crlf
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            Hello,
            World!
            """

    Scenario: cat crlf
        # arrange
        Given command "cat"
        And newline stdin crlf
        And stdin
            """
            Hello,
            World!
            """
        And newline output crlf
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            Hello,
            World!
            """
