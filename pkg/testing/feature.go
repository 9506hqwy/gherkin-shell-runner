package testing

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

const ZERO = 0
const EmptyString = ""
const DefaultWait = 30
const DefaultTimeout = 60 * 1000
const DefaultFilePerm os.FileMode = 0644

type tuiFeatureKey struct{}

type tuiFeature struct {
	workspace      string
	envs           map[string]string
	command        string
	args           []string
	stdin          string
	exitCode       int
	output         []byte
	wait           int
	timeout        int
	size           *TerminalSize
	ouputEncoding  string
	stdinEncoding  string
	fileEncoding   string
	filePermission os.FileMode
	delWorkspace   bool
	vars           map[string]string
}

type TerminalSize struct {
	width  int
	height int
}

func initTuituiFeature(t *tuiFeature) *tuiFeature {
	_ = cleanTuiFeature(t)

	t.workspace = EmptyString
	t.envs = make(map[string]string)
	t.command = EmptyString
	t.args = make([]string, ZERO)
	t.stdin = EmptyString
	t.exitCode = ZERO
	t.output = make([]byte, ZERO)
	t.wait = DefaultWait
	t.timeout = DefaultTimeout
	t.size = nil
	t.ouputEncoding = EmptyString
	t.stdinEncoding = EmptyString
	t.fileEncoding = EmptyString
	t.filePermission = DefaultFilePerm
	t.delWorkspace = false

	if t.vars == nil {
		t.vars = make(map[string]string)
	}

	return t
}

func resetTuituiFeature(t *tuiFeature) *tuiFeature {
	t.args = make([]string, ZERO)
	t.stdin = EmptyString
	t.exitCode = ZERO
	t.output = make([]byte, ZERO)
	return t
}

func getTuiFeature(ctx context.Context) *tuiFeature {
	t, ok := ctx.Value(tuiFeatureKey{}).(*tuiFeature)
	if !ok {
		panic("Not found TUI feature in context.")
	}

	return t
}

func setTuiFeature(ctx context.Context) context.Context {
	t := tuiFeature{}

	initTuituiFeature(&t)

	return context.WithValue(ctx, tuiFeatureKey{}, &t)
}

func cleanTuiFeature(t *tuiFeature) error {
	var err error
	if t.delWorkspace {
		err = os.RemoveAll(t.workspace)
		if err == nil {
			t.delWorkspace = false
		}
	}

	return err
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(
		ctx context.Context,
		_ *godog.Scenario,
	) (context.Context, error) {
		return setTuiFeature(ctx), nil
	})

	ctx.After(func(
		ctx context.Context,
		_ *godog.Scenario,
		_ error,
	) (context.Context, error) {
		t := getTuiFeature(ctx)
		return ctx, cleanTuiFeature(t)
	})

	ctx.Given(`^command (.+)$`, setCommand)

	ctx.Step(`^workspace (.+)$`, setWorkspace)
	ctx.Step(`^env$`, setEnvironmentTable)
	ctx.Step(`^env ([^ ]+) (.+)$`, setEnvironment)
	ctx.Step(`^arg (.+)$`, setArgument)
	ctx.Step(`^stdin$`, setStdinBlock)
	ctx.Step(`^stdin (.+)$`, setStdinLine)

	ctx.Step(`^wait (\d+)$`, setWait)
	ctx.Step(`^timeout (\d+)$`, setTimeout)
	ctx.Step(`^size (\d+) (\d+)$`, setSize)
	ctx.Step(`^encoding output (.+)$`, setOutputEncoding)
	ctx.Step(`^encoding stdin (.+)$`, setStdinEncoding)
	ctx.Step(`^encoding file (.+)$`, setFileEncoding)
	ctx.Step(`^use temp workspace$`, setTempWorkspace)
	ctx.Step(`^set ([A-Za-z][0-9A-Za-z_]*) (.+)$`, setVariable)

	ctx.Step(`^read file (.+) to ([A-Za-z][0-9A-Za-z_]*)$`, readFile)
	ctx.Step(`^write file (.+) from ([A-Za-z][0-9A-Za-z_]*)$`, writeFileLine)
	ctx.Step(`^write file (.+)$`, writeFileBlock)
	ctx.Step(`^chmod (\d+) file$`, setFilePermission)

	ctx.When(`^exec$`, execCommand)

	ctx.Then(`^status eq (-?\d+)$`, checkStatusEq)
	ctx.Then(`^status not eq (-?\d+)$`, checkStatusNotEq)

	ctx.Step(`^output is empty$`, checkOutputIsEmpty)
	ctx.Step(`^output is not empty$`, checkOutputIsNotEmpty)
	ctx.Step(`^output eq$`, checkOutputEqBlock)
	ctx.Step(`^output eq (.+)$`, checkOutputEqLine)
	ctx.Step(`^output not eq$`, checkOutputNotEqBlock)
	ctx.Step(`^output not eq (.+)$`, checkOutputNotEqLine)
	ctx.Step(`^output regex$`, checkOutputRegexBlock)
	ctx.Step(`^output regex (.+)$`, checkOutputRegexLine)
	ctx.Step(`^output not regex$`, checkOutputNotRegexBlock)
	ctx.Step(`^output not regex (.+)$`, checkOutputNotRegexLine)
}
