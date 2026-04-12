@linux
Feature: echo command
    Run echo command in Linux.

    Rule: Example 1
        Background:
            # set environment variable.
            * env KEY "value"

        Example: echo text
            # arrange
            Given command "echo"
            And arg "-n"
            And arg "Hello"
            # act
            When exec
            # assert
            Then status eq 0
            And output eq "Hello"

        Example: echo text
            # arrange
            Given command "echo"
            And arg "-n"
            And arg "World"
            # act
            When exec
            # assert
            Then status eq 0
            And output eq "World"

    Rule: Example 2
        Example: echo text
            # arrange
            Given command "echo"
            And arg "-n"
            And arg "Hello, World"
            # act
            When exec
            # assert
            Then status eq 0
            And output eq "Hello, World"
