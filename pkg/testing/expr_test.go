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
				t.Fatalf("no match token: %v", token)
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
				t.Fatalf("no match token: %v", token)
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
				t.Fatalf("no match token: %v", token)
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_parseInput_Operator(t *gotesting.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"plus operator", "+", "+"},
		{"plusplus operator", "++", "++"},
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
			if token.tokenType != TokenOperator || token.value != c.expected {
				t.Fatalf("no match token: %v", token)
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_parseInput_Whitespace(t *gotesting.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"space", " ", " "},
		{"spacespace", "  ", "  "},
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
			if token.tokenType != TokenWhitespace || token.value != c.expected {
				t.Fatalf("no match token: %v", token)
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_parseInput_Expr(t *gotesting.T) {
	cases := []struct {
		name         string
		input        string
		expected     string
		tokens       int
		instructions int
	}{
		{"plus number1", "1+1", "2", 3, 3},
		{"plus number2", "1 + 1", "2", 5, 3},
		{"plus number3", "1 + 1 + 1", "3", 9, 5},
		{"plus literal", "\"a\"+\"b\"", "ab", 3, 3},
		{"plus literal", "\"a\" + \"b\"", "ab", 5, 3},
		{"plus variable", "1 + v1", "3", 5, 3},
		{"plus variable", "\"a\" + v2", "ac", 5, 3},
	}

	feat := &tuiFeature{
		vars: map[string]string{
			"v1": "2",
			"v2": "c",
		},
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

			if len(r.tokens) != c.tokens {
				t.Fatalf("count of tokens: %d", len(r.tokens))
			}

			if err := r.completed(); err != nil {
				t.Fatal(err)
			}

			cal := &calculator{}
			err = cal.compile(r.tokens)
			if err != nil {
				t.Fatal(err)
				return
			}

			if len(cal.instructions) != c.instructions {
				t.Fatalf("count of instructions: %d", len(cal.instructions))
			}

			instruction := cal.instructions[len(cal.instructions)-1]
			if instruction.instType != InstructionTypeAdd {
				t.Fatalf("no match instruction: %v", instruction)
			}

			result, err := cal.calculate(feat)
			if err != nil {
				t.Fatal(err)
				return
			}

			if result != c.expected {
				t.Fatalf("no match result: %s", result)
			}
		})
	}
}

//revive:enable
