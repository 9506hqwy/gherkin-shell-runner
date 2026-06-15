//go:build unix

package testing

func initTuiFeatureByPlatform(t *tuiFeature) *tuiFeature {
	t.outputNewline = []byte{KeyCodeLf}
	t.stdinNewline = []byte{KeyCodeLf}

	return t
}
