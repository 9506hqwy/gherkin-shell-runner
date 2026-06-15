//go:build windows

package testing

func initTuiFeatureByPlatform(t *tuiFeature) *tuiFeature {
	t.outputNewline = []byte{KeyCodeCr, KeyCodeLf}
	t.stdinNewline = []byte{KeyCodeCr, KeyCodeLf}

	return t
}
