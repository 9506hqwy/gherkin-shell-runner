@linux
Feature: iconv command
    Run iconv command in Linux.

    Scenario: iconv text
        # arrange
        Given command iconv
        And arg -f
        And arg utf8
        And arg -t
        And arg sjis
        And stdin こんにちは
        # act
        When exec
        # assert
        Then status eq 0
        And encoding output sjis
        And output eq こんにちは
