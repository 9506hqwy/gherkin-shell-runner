package testing

import (
	"bufio"
	"bytes"
	"io"
)

type SequenceMode int

const (
	SequenceModeNone SequenceMode = 1 << iota
	SequenceModeEsc
	SequenceModeCsi
	SequenceModeOsc
)

func (m SequenceMode) EscStartedMode() bool {
	return m == SequenceModeEsc
}

func (m SequenceMode) EscMode() bool {
	return (m & SequenceModeEsc) == SequenceModeEsc
}

func (m SequenceMode) CsiMode() bool {
	return (m & SequenceModeCsi) == SequenceModeCsi
}

func (m SequenceMode) OscMode() bool {
	return (m & SequenceModeOsc) == SequenceModeOsc
}

const (
	KeyCodeNul byte = 0x00 // NUL
	KeyCodeHt  byte = 0x09 // HT
	KeyCodeLf  byte = 0x0a // LF
	KeyCodeBel byte = 0x07 // BEL
	KeyCodeEsc byte = 0x1b // ESC
	KeyCodeUs  byte = 0x1f // US
	KeyCodeSpc byte = 0x20 // SPC
	KeyCodeDqt byte = 0x22 // "
	KeyCodeSqt byte = 0x27 // '
	KeyCodeHyp byte = 0x2d // -
	KeyCode0   byte = 0x30 // 0
	KeyCode9   byte = 0x39 // 9
	KeyCodeExc byte = 0x3f // ?
	KeyCodeAt  byte = 0x40 // @
	KeyCodeA   byte = 0x41 // A
	KeyCodeZ   byte = 0x5a // Z
	KeyCodeOsb byte = 0x5b // [
	KeyCodeBsh byte = 0x5c // \
	KeyCodeCsb byte = 0x5d // ]
	KeyCodeUsr byte = 0x5f // _
	KeyCodea   byte = 0x61 // a
	KeyCodez   byte = 0x7a // z
	KeyCodeTil byte = 0x7e // ~
	KeyCodeDel byte = 0x7f // DEL
	KeyCodeCsi byte = 0x9b // ESC [
	KeyCodeOsc byte = 0x9d // ESC ]
)

type terminal struct {
	mode   SequenceMode
	buffer *bytes.Buffer
}

func newTerminal() *terminal {
	buffer := new(bytes.Buffer)

	t := terminal{
		mode:   SequenceModeNone,
		buffer: buffer,
	}

	return &t
}

func (t *terminal) Buffer() *bytes.Buffer {
	return t.buffer
}

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

	count := int64(ZERO)
	reader := bufio.NewReader(src)

	for {
		count += t.comsumeCtrlSeq(reader)

		b, err := reader.ReadByte()
		if comsumed(b, err) {
			break
		}

		count++

		if t.startCtrlSeq(b) {
			continue
		}

		err = t.buffer.WriteByte(b)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func (t *terminal) startEsc(b byte) bool {
	if !t.mode.EscMode() && (b == KeyCodeEsc) {
		t.mode = SequenceModeEsc
		return true
	}

	return false
}

func (t *terminal) startCsi(b byte) bool {
	if t.mode.EscStartedMode() && (b == KeyCodeOsb) {
		t.mode |= SequenceModeCsi
		return true
	}

	if !t.mode.EscMode() && b == KeyCodeCsi {
		t.mode = SequenceModeEsc | SequenceModeCsi
		return true
	}

	return false
}

func (t *terminal) startOsc(b byte) bool {
	if t.mode.EscStartedMode() && (b == KeyCodeCsb) {
		t.mode |= SequenceModeOsc
		return true
	}

	if !t.mode.EscMode() && b == KeyCodeOsc {
		t.mode = SequenceModeEsc | SequenceModeOsc
		return true
	}

	return false
}

func (t *terminal) startCtrlSeq(b byte) bool {
	if t.startEsc(b) {
		return true
	}

	if t.startCsi(b) {
		return true
	}

	if t.startOsc(b) {
		return true
	}

	if nonPrintable(b) {
		return true
	}

	return false
}

func (t *terminal) comsumeCsi(reader *bufio.Reader) int64 {
	count := int64(ZERO)

	for {
		b, err := reader.ReadByte()
		if comsumed(b, err) {
			break
		}

		count++

		if csiKeyCode(b) {
			continue
		}

		if csiKeyCodeStop(b) {
			t.mode = SequenceModeNone
			break
		}
	}

	return count
}

func (t *terminal) comsumeOsc(reader *bufio.Reader) int64 {
	count := int64(ZERO)

	for {
		b, err := reader.ReadByte()
		if comsumed(b, err) {
			break
		}

		count++

		if b != KeyCodeBel {
			continue
		}

		t.mode = SequenceModeNone
		break
	}

	return count
}

func (t *terminal) comsumeCtrlSeq(reader *bufio.Reader) int64 {
	count := int64(ZERO)

	if t.mode.CsiMode() {
		n := t.comsumeCsi(reader)
		count += n
	}

	if t.mode.OscMode() {
		n := t.comsumeOsc(reader)
		count += n
	}

	return count
}

func comsumed(b byte, err error) bool {
	return b == KeyCodeNul || err == io.EOF
}

func csiKeyCode(b byte) bool {
	return b >= KeyCodeSpc && b <= KeyCodeExc
}

func csiKeyCodeStop(b byte) bool {
	return b >= KeyCodeAt && b <= KeyCodeTil
}

func nonPrintable(b byte) bool {
	if b <= KeyCodeUs || b == KeyCodeDel {
		if b != KeyCodeHt && b != KeyCodeLf {
			return true
		}
	}

	return false
}
