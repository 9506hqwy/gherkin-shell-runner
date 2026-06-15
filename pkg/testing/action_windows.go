//go:build windows

package testing

import (
	"bytes"

	"github.com/aymanbagabas/go-pty"
)

func setPty(_ *tuiFeature, _ *pty.Pty) error {
	// Need to configure ENABLE_ECHO_INPUT to ptyIn ?
	// https://github.com/aymanbagabas/go-pty/blob/v0.2.3/pty_windows.go#L52
	// https://devblogs.microsoft.com/commandline/windows-command-line-introducing-the-windows-pseudo-console-conpty/
	return nil
}

func inputStdin(
	t *tuiFeature,
	ptmx *pty.Pty,
) error {
	// replace LF to CR.
	encodedBytes := bytes.ReplaceAll(t.stdin, []byte{KeyCodeCr, KeyCodeLf}, []byte{KeyCodeLf})
	encodedBytes = bytes.ReplaceAll(encodedBytes, []byte{KeyCodeLf}, []byte{KeyCodeCr})

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
