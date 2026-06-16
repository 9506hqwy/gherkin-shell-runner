//go:build windows

package testing

import (
	"bytes"
	"errors"
	"sync"

	"github.com/aymanbagabas/go-pty"
	"golang.org/x/sys/windows"
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

func waitBufferingCompleted(ptmx *pty.Pty, wg *sync.WaitGroup) error {
	// Close pty only, output EOF to output pipe.
	// When close pty and pipe, output abort, so wait.
	fd := (*ptmx).Fd()
	windows.ClosePseudoConsole(windows.Handle(fd))

	// Wait for buffering output completely.
	(*wg).Wait()

	// Close pipe after buffering.
	conPty := (*ptmx).(pty.ConPty)
	return errors.Join(
		conPty.InputPipe().Close(),
		conPty.OutputPipe().Close(),
	)
}
