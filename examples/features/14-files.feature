@linux
Feature: files
    Run cat command in Linux.

    Scenario: read file
        # arrange
        Given command "cat"
        And read file "/etc/os-release" to content
        And arg "/etc/os-release"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq content

    Scenario: write content tofile
        # arrange
        Given command "cat"
        And use temp workspace
        And chmod 0755 file
        And write file "text.txt"
            """
            Hello, World!
            """
        And arg "text.txt"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq "Hello, World!"

    Scenario: write variable to file
        # arrange
        Given command "cat"
        And use temp workspace
        And set value "Hello, World!"
        And write file "text.txt" from value
        And arg "text.txt"
        # act
        When exec
        # assert
        Then status eq 0
        And output eq value
