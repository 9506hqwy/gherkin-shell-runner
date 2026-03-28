package testing

import (
	"bytes"
	"io"
)

type SequenceMode int

const (
	None SequenceMode = 1 << iota
	Esc
	Csi
	Osc
)

type terminal struct {
	mode   SequenceMode
	buffer *bytes.Buffer
}

func (t *terminal) Buffer() *bytes.Buffer {
	return t.buffer
}

//revive:disable

func (t *terminal) Copy(src io.Reader) (written int64, err error) {
	// Remove CSI (Control Sequence Introducer)
	// - 7bit sequence
	//   0x1b 0x5b [0x20 - 0x3f]+ [0x40 - 0x7e]
	// - 8bit sequence
	//   0x9b [0x20 - 0x3f]+ [0x40 - 0x7e]
	// Remove OSC (Operating System Command)
	// - 7bit sequence
	//   0x1b 0x5d [^0x07]+ 0x07
	// - 8bit sequence
	//   0x9d [^0x07]+ 0x07
	// Remove control sequence
	//   0x00 - 0x1f, 0x7f

	count := int64(0)
	buffer := make([]byte, 1)
	for {
		n, err := src.Read(buffer)
		if err == io.EOF || n == 0 {
			break
		}

		count += int64(n)
		b := buffer[0]

		if (t.mode&Esc) != Esc && b == 0x1b { // ESC
			t.mode = Esc
			continue
		}

		if t.mode == Esc && (b == 0x5b) { // [
			t.mode |= Csi
			continue
		}

		if t.mode == Esc && (b == 0x5d) { // ]
			t.mode |= Osc
			continue
		}

		if (t.mode&Esc) != Esc && b == 0x9b { // ESC [
			t.mode = Esc | Csi
			continue
		}

		if (t.mode&Esc) != Esc && b == 0x9d { // ESC ]
			t.mode = Esc | Osc
			continue
		}

		if (t.mode & Csi) == Csi {
			if b >= 0x20 && b <= 0x3f {
				continue
			} else if b >= 0x40 && b <= 0x7e {
				t.mode = None
				continue
			}
		}

		if (t.mode & Osc) == Osc {
			if b != 0x07 {
				continue
			} else {
				t.mode = None
				continue
			}
		}

		if b <= 0x1f || b == 0x7f { // Control Sequence
			if b != 0x09 && b != 0x0a { // HT, LF
				continue
			}
		}

		_, err = t.buffer.Write(buffer)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

//revive:enable

func newTerminal() *terminal {
	buffer := new(bytes.Buffer)

	t := terminal{
		mode:   None,
		buffer: buffer,
	}

	return &t
}
