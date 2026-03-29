@linux
Feature: ls and iconv
    Run ls and iconv command in Linux.

    Scenario: ls directory
        # arrange
        Given command "ls"
        And arg "-l"
        And arg "./examples/features"
        # act
        When exec
        # assert
        Then status eq 0
        And output regex
            """
            (?m)12-regex.feature$
            """
        And output not regex
            """
            (?m)99-unknown.feature$
            """

    Scenario: iconv text
        # arrange
        Given command "iconv"
        And arg "-f"
        And arg "utf8"
        And arg "-t"
        And arg "sjis"
        And stdin "こんにちは"
        # act
        When exec
        # assert
        Then status eq 0
        And encoding output sjis
        And output regex "んにち"
