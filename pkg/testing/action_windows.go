//go:build windows

package testing

import (
	"bytes"

	"github.com/aymanbagabas/go-pty"
)

func setPty(_ *tuiFeature, _ *pty.Pty) error {
	return nil
}

func inputStdin(
	t *tuiFeature,
	ptmx *pty.Pty,
) error {
	inputBytes := []byte(t.stdin)

	// replace LF to CR.
	encodedBytes := bytes.ReplaceAll(inputBytes, []byte{0x0a}, []byte{0x0d})

	_, err := (*ptmx).Write(encodedBytes)
	if err != nil {
		return err
	}

	// CTRL+Z (\x1a) + CR
	_, err = (*ptmx).Write([]byte("\x1a\r"))
	if err != nil {
		return err
	}

	return nil
}
