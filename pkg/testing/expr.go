package testing

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type TokenType int

const (
	TokenLiteral TokenType = iota
	TokenNumber
	TokenVariable
)

type token struct {
	tokenType TokenType
	value     string
}

type tokenReaderMode int

const (
	TokenReaderModeNone tokenReaderMode = iota
	TokenReaderModeLiteral
	TokenReaderModeLiteralEsc
	TokenReaderModeNumber
	TokenReaderModeVariable
)

type tokenHandlerFunc func(b byte) (bool, error)

type tokenReader struct {
	src         *bufio.Reader
	mode        tokenReaderMode
	tokens      []token
	current     []byte
	literalCode byte
	pendings    []byte
}

func newTokenReader(src io.Reader) *tokenReader {
	r := tokenReader{
		src:         bufio.NewReader(src),
		mode:        TokenReaderModeNone,
		tokens:      []token{},
		current:     []byte{},
		literalCode: KeyCodeNul,
		pendings:    []byte{},
	}

	return &r
}

func (r *tokenReader) parseInput() error {
	for {
		b, err := r.src.ReadByte()
		if comsumed(b, err) {
			err = r.addToken(io.EOF)
			return err
		}

		err = r.handleToken(b)
		if err != nil {
			return err
		}
	}
}

func (r *tokenReader) completed() error {
	if len(r.current) != ZERO {
		return fmt.Errorf("not found EOF: %s", string(r.current))
	}

	if r.literalCode != KeyCodeNul {
		return fmt.Errorf(
			"not found closing literal: %s",
			string(r.literalCode),
		)
	}

	if len(r.pendings) != ZERO {
		return fmt.Errorf(
			"not found closing escape: %s",
			string(r.pendings),
		)
	}

	return nil
}

func (r *tokenReader) handleToken(b byte) error {
	tokenHandler := []tokenHandlerFunc{
		r.handleLiteral,
		r.handleLiteralEsc,
		r.handleNumber,
		r.handleVariable,
	}

	for _, handler := range tokenHandler {
		if handled, err := handler(b); handled {
			return err
		}
	}

	r.current = append(r.current, b)
	return nil
}

func (r *tokenReader) handleLiteral(b byte) (bool, error) {
	if started, err := r.startLiteral(b); started {
		return true, err
	}

	if ended, err := r.endLiteral(b); ended {
		return true, err
	}

	return false, nil
}

func (r *tokenReader) handleLiteralEsc(b byte) (bool, error) {
	if started, err := r.startLiteralEsc(b); started {
		return true, err
	}

	if ended, err := r.endLiteralEsc(b); ended {
		return true, err
	}

	return false, nil
}

func (r *tokenReader) handleNumber(b byte) (bool, error) {
	if started, err := r.startNumber(b); started {
		r.current = append(r.current, b)
		return true, err
	}

	if ended, err := r.endNumber(b); ended {
		r.current = append(r.current, b)
		return true, err
	}

	return false, nil
}

func (r *tokenReader) handleVariable(b byte) (bool, error) {
	if started, err := r.startVariable(b); started {
		r.current = append(r.current, b)
		return true, err
	}

	if ended, err := r.endVariable(b); ended {
		r.current = append(r.current, b)
		return true, err
	}

	return false, nil
}

func (r *tokenReader) startLiteral(b byte) (bool, error) {
	if r.mode != TokenReaderModeNone {
		return false, nil
	}

	if b != KeyCodeDqt && b != KeyCodeSqt {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeLiteral
	r.literalCode = b
	return true, nil
}

func (r *tokenReader) endLiteral(b byte) (bool, error) {
	if r.mode != TokenReaderModeLiteral {
		return false, nil
	}

	if b != r.literalCode {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeNone
	r.literalCode = KeyCodeNul
	return true, nil
}

func (r *tokenReader) startLiteralEsc(b byte) (bool, error) {
	if r.mode != TokenReaderModeLiteral {
		return false, nil
	}

	if b != KeyCodeBsh {
		return false, nil
	}

	r.mode = TokenReaderModeLiteralEsc
	r.pendings = append(r.pendings, b)
	return true, nil
}

func (r *tokenReader) endLiteralEsc(b byte) (bool, error) {
	if r.mode != TokenReaderModeLiteralEsc {
		return false, nil
	}

	if b != r.literalCode {
		r.current = append(r.current, r.pendings...)
	}

	r.current = append(r.current, b)

	r.mode = TokenReaderModeLiteral
	r.pendings = []byte{}
	return true, nil
}

func (r *tokenReader) startNumber(b byte) (bool, error) {
	if r.mode != TokenReaderModeNone {
		return false, nil
	}

	if !number(b) && b != KeyCodeHyp {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeNumber
	return true, nil
}

func (r *tokenReader) endNumber(b byte) (bool, error) {
	if r.mode != TokenReaderModeNumber {
		return false, nil
	}

	if number(b) {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeNone
	return true, nil
}

func (r *tokenReader) startVariable(b byte) (bool, error) {
	if r.mode != TokenReaderModeNone {
		return false, nil
	}

	if !upperAlpha(b) && !lowerAlpha(b) {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeVariable
	return true, nil
}

func (r *tokenReader) endVariable(b byte) (bool, error) {
	if r.mode != TokenReaderModeVariable {
		return false, nil
	}

	if upperAlpha(b) || lowerAlpha(b) || number(b) || b == KeyCodeUsr {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeNone
	return true, nil
}

func (r *tokenReader) addToken(eof error) error {
	if len(r.current) == ZERO {
		return nil
	}

	tokenType := TokenLiteral
	switch r.mode {
	case TokenReaderModeLiteral:
		if eof == io.EOF {
			// Avoiding matching end of sentence without closing '"'.
			return fmt.Errorf(
				"not found closing literal: %s",
				string(r.current),
			)
		}
		tokenType = TokenLiteral
	case TokenReaderModeNumber:
		tokenType = TokenNumber
	case TokenReaderModeVariable:
		tokenType = TokenVariable
	default:
		return fmt.Errorf("not supported format: %s", string(r.current))
	}

	t := token{
		tokenType: tokenType,
		value:     string(r.current),
	}

	r.tokens = append(r.tokens, t)
	r.current = []byte{}
	return nil
}

func upperAlpha(b byte) bool {
	return b >= KeyCodeA && b <= KeyCodeZ
}

func lowerAlpha(b byte) bool {
	return b >= KeyCodea && b <= KeyCodez
}

func number(b byte) bool {
	return b >= KeyCode0 && b <= KeyCode9
}

func parseValueOne(value string) (string, error) {
	input := strings.NewReader(value)
	reader := newTokenReader(input)
	err := reader.parseInput()
	if err != nil {
		return EmptyString, err
	}

	//revive:disable:add-constant
	if len(reader.tokens) != 1 {
		return EmptyString, fmt.Errorf("not supported format: %s", value)
	}
	//revive:enable:add-constant

	token := reader.tokens[ZERO]
	return token.value, nil
}
