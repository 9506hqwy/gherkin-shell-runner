package testing

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func checkStatusEq(ctx context.Context, expect int) (context.Context, error) {
	t := getTuiFeature(ctx)
	if t.exitCode != expect {
		return ctx, fmt.Errorf(
			"expected exit code to be: %d, but actual is: %d",
			expect,
			t.exitCode)
	}

	return ctx, nil
}

func checkOutputIsEmpty(ctx context.Context) (context.Context, error) {
	return checkOutputEqLine(ctx, "")
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
	t := getTuiFeature(ctx)
	expectedBytes, err := encodingToBytes(t.ouputEncoding, expect)
	if err != nil {
		return ctx, err
	}

	if !bytes.Equal(t.output, expectedBytes) {
		return ctx, fmt.Errorf(
			"expected Output to be: '%s', but actual is: '%s'",
			expect,
			t.output)
	}

	return ctx, nil
}

func encodingToBytes(encoding string, value string) ([]byte, error) {
	if encoding == "" {
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
