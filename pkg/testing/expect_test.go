package testing

import (
	"bytes"
	gotesting "testing"
)

//revive:disable

func Test_convertNewline(t *gotesting.T) {
	cases := []struct {
		name     string
		newline  string
		input    string
		expected string
	}{
		{"crlf to lf", "\n", "a\r\nb", "a\nb"},
		{"crlf to cr", "\r", "a\r\nb", "a\rb"},
		{"cr to lf", "\n", "a\rb", "a\nb"},
		{"lf to crlf", "\r\n", "a\rb", "a\r\nb"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			v := convertNewline([]byte(c.input), []byte(c.newline))

			if !bytes.Equal(v, []byte(c.expected)) {
				t.Fatalf("no match result: %s", v)
			}
		})
	}
}

//revive:enable
