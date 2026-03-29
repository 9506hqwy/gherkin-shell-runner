package testing

import (
	"fmt"
	"strings"
	gotesting "testing"
)

//revive:disable

func Test_parseInput_Variable(t *gotesting.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"a", "a", "a"},
		{"z", "z", "z"},
		{"A", "A", "A"},
		{"Z", "Z", "Z"},
		{"two", "ab", "ab"},
		{"number0", "a0", "a0"},
		{"number9", "a9", "a9"},
		{"under score", "a_", "a_"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			input := strings.NewReader(c.input)

			r := newTokenReader(input)

			err := r.parseInput()
			if err != nil {
				t.Fatal(err)
			}

			if len(r.tokens) != 1 {
				t.Fatalf("count of tokens: %d", len(r.tokens))
			}

			token := r.tokens[0]
			if token.tokenType != TokenVariable || token.value != c.expected {
				t.Fatalf("no match token: %v", r.tokens[0])
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_parseInput_Literal(t *gotesting.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quoted", "\"a\"", "a"},
		{"escaped double quoted", "\"a\\\"b\"", "a\"b"},
		{"non-escaped singled quoted", "\"a'b\"", "a'b"},
		{"single quoted", "'a'", "a"},
		{"escaped single quoted", "'a\\'b'", "a'b"},
		{"non-escaped double quoted", "'a\"b'", "a\"b"},
		{"non-escaped", "'a\\\\b'", "a\\\\b"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			input := strings.NewReader(c.input)

			r := newTokenReader(input)

			err := r.parseInput()
			if err != nil {
				t.Fatal(err)
			}

			if len(r.tokens) != 1 {
				t.Fatalf("count of tokens: %d", len(r.tokens))
			}

			token := r.tokens[0]
			if token.tokenType != TokenLiteral || token.value != c.expected {
				t.Fatalf("no match token: %v", r.tokens[0])
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_parseInput_Literal_is_NotClosed(t *gotesting.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quoted", "\"a\\\"", "a\""},
		{"single quoted", "'a\\'", "a'"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			input := strings.NewReader(c.input)

			r := newTokenReader(input)

			err := r.parseInput()
			if err == nil {
				t.Fatalf("no occurred error.")
			}

			if err.Error() != fmt.Sprintf("not found closing literal: %s", c.expected) {
				t.Fatal(err)
			}
		})
	}
}

func Test_parseInput_Number(t *gotesting.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"positive number 0", "0", "0"},
		{"positive number 1", "1", "1"},
		{"positive number 11", "11", "11"},
		{"positive number -1", "-1", "-1"},
		{"positive number -11", "-11", "-11"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			input := strings.NewReader(c.input)

			r := newTokenReader(input)

			err := r.parseInput()
			if err != nil {
				t.Fatal(err)
				return
			}

			if len(r.tokens) != 1 {
				t.Fatalf("count of tokens: %d", len(r.tokens))
			}

			token := r.tokens[0]
			if token.tokenType != TokenNumber || token.value != c.expected {
				t.Fatalf("no match token: %v", r.tokens[0])
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

//revive:enable
