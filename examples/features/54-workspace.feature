@windows
Feature: workspace
    Run cd command in Windows.

    Scenario: show current directory
        # arrange
        Given command ${SYSTEMROOT}\System32\cmd.exe
        And workspace C:\
        And arg /c
        And arg cd
        # act
        When exec
        # assert
        Then status eq 0
        And output eq
            """
            C:\

            """
