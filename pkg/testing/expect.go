package testing

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type CompareMode int

const (
	Equal CompareMode = iota
	NotEqual
)

func checkStatusEq(ctx context.Context, expect int) (context.Context, error) {
	return checkStatus(ctx, expect, Equal)
}

func checkStatusNotEq(
	ctx context.Context,
	expect int,
) (context.Context, error) {
	return checkStatus(ctx, expect, NotEqual)
}

func checkStatus(
	ctx context.Context,
	expect int,
	mode CompareMode,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	if (t.exitCode == expect) != (mode == Equal) {
		return ctx, fmt.Errorf(
			"expected exit code to be: %d, but actual is: %d",
			expect,
			t.exitCode)
	}

	return ctx, nil
}

func checkOutputIsEmpty(ctx context.Context) (context.Context, error) {
	return checkOutputEqLine(ctx, EmptyString)
}

func checkOutputIsNotEmpty(ctx context.Context) (context.Context, error) {
	t := getTuiFeature(ctx)
	if len(t.output) == ZERO {
		return ctx, fmt.Errorf(
			"expected Output to be: '', but actual is: '%s'",
			t.output)
	}

	return ctx, nil
}

func checkOutputEqBlock(
	ctx context.Context,
	expect *godog.DocString,
) (context.Context, error) {
	return checkOutputEqLine(ctx, expect.Content)
}

func checkOutputEqLine(
	ctx context.Context,
	expect string,
) (context.Context, error) {
	return checkOutputEq(ctx, expect, Equal)
}

func checkOutputNotEqBlock(
	ctx context.Context,
	expect *godog.DocString,
) (context.Context, error) {
	return checkOutputNotEqLine(ctx, expect.Content)
}

func checkOutputNotEqLine(
	ctx context.Context,
	expect string,
) (context.Context, error) {
	return checkOutputEq(ctx, expect, NotEqual)
}

func checkOutputEq(
	ctx context.Context,
	expect string,
	mode CompareMode,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	expectedBytes, err := encodingToBytes(t.ouputEncoding, expect)
	if err != nil {
		return ctx, err
	}

	if bytes.Equal(t.output, expectedBytes) != (mode == Equal) {
		return ctx, fmt.Errorf(
			"expected Output to be: '%s', but actual is: '%s'",
			expect,
			t.output)
	}

	return ctx, nil
}

func checkOutputRegexBlock(
	ctx context.Context,
	pattern *godog.DocString,
) (context.Context, error) {
	return checkOutputRegexLine(ctx, pattern.Content)
}

func checkOutputRegexLine(
	ctx context.Context,
	pattern string,
) (context.Context, error) {
	return checkOutputRegex(ctx, pattern, Equal)
}

func checkOutputNotRegexBlock(
	ctx context.Context,
	pattern *godog.DocString,
) (context.Context, error) {
	return checkOutputNotRegexLine(ctx, pattern.Content)
}

func checkOutputNotRegexLine(
	ctx context.Context,
	pattern string,
) (context.Context, error) {
	return checkOutputRegex(ctx, pattern, NotEqual)
}

func checkOutputRegex(
	ctx context.Context,
	pattern string,
	mode CompareMode,
) (context.Context, error) {
	re := regexp.MustCompile(pattern)

	t := getTuiFeature(ctx)
	outputBytes, err := decodingToBytes(t.ouputEncoding, t.output)
	if err != nil {
		return ctx, err
	}

	if re.Match(outputBytes) != (mode == Equal) {
		return ctx, fmt.Errorf(
			"expected Output to match: '%s', but actual is: '%s'",
			pattern,
			t.output)
	}

	return ctx, nil
}

func decodingToBytes(encoding string, value []byte) ([]byte, error) {
	if encoding == EmptyString {
		return value, nil
	}

	if encoding == "sjis" {
		return decodingSjis(value)
	}

	return nil, fmt.Errorf("not suported output encoding. %s", encoding)
}

func decodingSjis(value []byte) ([]byte, error) {
	t := japanese.ShiftJIS.NewDecoder()
	toBytes, _, err := transform.Bytes(t, value)

	return toBytes, err
}

func encodingToBytes(encoding string, value string) ([]byte, error) {
	if encoding == EmptyString {
		return []byte(value), nil
	}

	if encoding == "sjis" {
		return encodingSjis(value)
	}

	return nil, fmt.Errorf("not suported output encoding. %s", encoding)
}

func encodingSjis(value string) ([]byte, error) {
	fromBytes := []byte(value)

	t := japanese.ShiftJIS.NewEncoder()
	toBytes, _, err := transform.Bytes(t, fromBytes)

	return toBytes, err
}
