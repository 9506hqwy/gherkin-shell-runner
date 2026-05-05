//go:build unix

package testing

import (
	"github.com/aymanbagabas/go-pty"
	"github.com/u-root/u-root/pkg/termios"
)

func setPty(t *tuiFeature, ptmx *pty.Pty) error {
	term, err := termios.GTTY(int((*ptmx).Fd()))
	if err != nil {
		return err
	}

	// disable input echo.
	term.Opts["echo"] = false
	// disable `onlcr`.
	term.Opts["opost"] = false

	_, err = term.STTY(int((*ptmx).Fd()))
	if err != nil {
		return err
	}

	if t.size != nil {
		err = (*ptmx).Resize(t.size.width, t.size.height)
		if err != nil {
			return err
		}
	}

	return nil
}

func inputStdin(
	t *tuiFeature,
	ptmx *pty.Pty,
) error {
	_, err := (*ptmx).Write(t.stdin)
	if err != nil {
		return err
	}

	// CTRL+D (\x04)
	_, err = (*ptmx).Write([]byte("\x04\x04"))
	if err != nil {
		return err
	}

	return nil
}
