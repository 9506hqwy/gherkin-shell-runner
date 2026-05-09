package testing

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

type TokenType int

const (
	TokenLiteral TokenType = iota
	TokenNumber
	TokenOperator
	TokenVariable
	TokenWhitespace
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
	TokenReaderModeOperator
	TokenReaderModeVariable
	TokenReaderModeWhitespace
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

func (r *tokenReader) startTokenHandler() []tokenHandlerFunc {
	return []tokenHandlerFunc{
		r.startLiteral,
		r.startLiteralEsc,
		r.startNumber,
		r.startOperator,
		r.startVariable,
		r.startWhitespace,
	}
}

func (r *tokenReader) endTokenHandler() []tokenHandlerFunc {
	return []tokenHandlerFunc{
		r.endLiteral,
		r.endLiteralEsc,
		r.endNumber,
		r.endOperator,
		r.endVariable,
		r.endWhitespace,
	}
}

func (r *tokenReader) handleToken(b byte) error {
	for _, handler := range r.endTokenHandler() {
		if next, err := handler(b); next {
			return err
		}
	}

	for _, handler := range r.startTokenHandler() {
		if next, err := handler(b); next {
			return err
		}
	}

	r.current = append(r.current, b)
	return nil
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
	return false, nil
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
	return false, nil
}

func (r *tokenReader) startOperator(b byte) (bool, error) {
	if r.mode != TokenReaderModeNone {
		return false, nil
	}

	if !operator(b) {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeOperator
	return false, nil
}

func (r *tokenReader) endOperator(b byte) (bool, error) {
	if r.mode != TokenReaderModeOperator {
		return false, nil
	}

	if operator(b) {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeNone
	return false, nil
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
	return false, nil
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
	return false, nil
}

func (r *tokenReader) startWhitespace(b byte) (bool, error) {
	if r.mode != TokenReaderModeNone {
		return false, nil
	}

	if !whitespace(b) {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeWhitespace
	return false, nil
}

func (r *tokenReader) endWhitespace(b byte) (bool, error) {
	if r.mode != TokenReaderModeWhitespace {
		return false, nil
	}

	if whitespace(b) {
		return false, nil
	}

	err := r.addToken(nil)
	if err != nil {
		return true, err
	}

	r.mode = TokenReaderModeNone
	return false, nil
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
	case TokenReaderModeOperator:
		tokenType = TokenOperator
	case TokenReaderModeVariable:
		tokenType = TokenVariable
	case TokenReaderModeWhitespace:
		tokenType = TokenWhitespace
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

func operator(b byte) bool {
	return b == KeyCodePls
}

func whitespace(b byte) bool {
	return b == KeyCodeSpc
}

func parseValueOne(t *tuiFeature, value string) (string, error) {
	input := strings.NewReader(value)
	reader := newTokenReader(input)
	err := reader.parseInput()
	if err != nil {
		return EmptyString, err
	}

	//revive:disable:add-constant
	if len(reader.tokens) != 1 {
		cal := &calculator{}
		err = cal.compile(reader.tokens)
		if err != nil {
			return EmptyString, err
		}

		return cal.calculate(t)
	}
	//revive:enable:add-constant

	token := reader.tokens[ZERO]

	switch token.tokenType {
	case TokenVariable:
		value, ok := t.vars[token.value]
		if !ok {
			return EmptyString, fmt.Errorf(
				"not found variable: %s",
				token.value,
			)
		}

		return value, nil
	default:
		return token.value, nil
	}
}

// ---------------------------------------------------------------------------

type InstructionType int

const (
	InstructionTypeVal InstructionType = iota
	InstructionTypeAdd
)

type instruction struct {
	instType InstructionType
	token    token
}

type calculator struct {
	instructions []instruction
}

func (r *calculator) compile(tokens []token) error {
	var stack []instruction

	for _, token := range tokens {
		switch token.tokenType {
		case TokenLiteral, TokenNumber, TokenVariable:
			instruction := instruction{
				instType: InstructionTypeVal,
				token:    token,
			}
			r.instructions = append(r.instructions, instruction)
		case TokenOperator:
			instruction := instruction{
				instType: InstructionTypeAdd,
				token:    token,
			}
			stack = append(stack, instruction)
		case TokenWhitespace:
			continue
		default:
			return fmt.Errorf("not supported token: %v", token)
		}
	}

	slices.Reverse(stack)
	r.instructions = append(r.instructions, stack...)

	return nil
}

//revive:disable:add-constant

func (r *calculator) calculate(t *tuiFeature) (string, error) {
	var stack []instruction

	for _, inst := range r.instructions {
		switch inst.instType {
		case InstructionTypeVal:
			stack = append(stack, inst)
		case InstructionTypeAdd:
			var err error
			stack, err = calculateAdd(t, stack)
			if err != nil {
				return EmptyString, err
			}
		default:
			return EmptyString, errors.New("invalid expression")
		}
	}

	if len(stack) != 1 {
		return EmptyString, errors.New("invalid expression")
	}

	return stack[0].token.value, nil
}

func calculateAdd(t *tuiFeature, stack []instruction) ([]instruction, error) {
	if len(stack) < 2 {
		return nil, errors.New("invalid expression")
	}

	y := stack[len(stack)-1]
	x := stack[len(stack)-2]
	stack = stack[:len(stack)-2]

	result, err := calculateTokenAdd(t, x.token, y.token)
	if err != nil {
		return nil, err
	}

	stack = append(stack, instruction{
		instType: InstructionTypeVal,
		token:    result,
	})

	return stack, nil
}

func calculateTokenAdd(t *tuiFeature, x token, y token) (token, error) {
	xToken, err := getVariableToken(t, x)
	if err != nil {
		return token{}, err
	}

	yToken, err := getVariableToken(t, y)
	if err != nil {
		return token{}, err
	}

	switch {
	case xToken.tokenType == TokenNumber || yToken.tokenType == TokenNumber:
		return calculateTokenAddNumber(xToken, yToken)
	default:
		return token{
			tokenType: TokenLiteral,
			value:     xToken.value + yToken.value,
		}, nil
	}
}

func getVariableToken(t *tuiFeature, value token) (token, error) {
	if value.tokenType == TokenVariable {
		v, ok := t.vars[value.value]
		if !ok {
			return token{}, fmt.Errorf("not found variable: %s", value.value)
		}

		value = token{
			tokenType: TokenLiteral,
			value:     v,
		}
	}

	return value, nil
}

func calculateTokenAddNumber(x token, y token) (token, error) {
	xValue, err := strconv.ParseInt(x.value, 10, 64)
	if err != nil {
		return token{}, fmt.Errorf("invalid number: %s", x.value)
	}

	yValue, err := strconv.ParseInt(y.value, 10, 64)
	if err != nil {
		return token{}, fmt.Errorf("invalid number: %s", y.value)
	}

	return token{
		tokenType: TokenNumber,
		value:     fmt.Sprintf("%d", xValue+yValue),
	}, nil
}

//revive:enable:add-constant
