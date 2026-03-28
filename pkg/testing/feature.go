package testing

import (
	"context"

	"github.com/cucumber/godog"
)

const ZERO = 0
const EmptyString = ""
const DefaultTimeout = 60 * 1000

type tuiFeatureKey struct{}

type tuiFeature struct {
	workspace     string
	envs          map[string]string
	command       string
	args          []string
	stdin         string
	exitCode      int
	output        []byte
	timeout       int
	size          *TerminalSize
	ouputEncoding string
	stdinEncoding string
}

type TerminalSize struct {
	width  int
	height int
}

func initTuituiFeature(t *tuiFeature) *tuiFeature {
	t.workspace = EmptyString
	t.envs = make(map[string]string)
	t.command = EmptyString
	t.args = make([]string, ZERO)
	t.stdin = EmptyString
	t.exitCode = ZERO
	t.output = make([]byte, ZERO)
	t.timeout = DefaultTimeout
	t.size = nil
	t.ouputEncoding = EmptyString
	t.stdinEncoding = EmptyString
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

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(
		ctx context.Context,
		_ *godog.Scenario,
	) (context.Context, error) {
		return setTuiFeature(ctx), nil
	})

	ctx.Given(`^command (.+)$`, setCommand)

	ctx.Step(`^workspace (.+)$`, setWorkspace)
	ctx.Step(`^env ([^ ]+) (.+)$`, setEnvironment)
	ctx.Step(`^arg (.+)$`, setArgument)
	ctx.Step(`^stdin$`, setStdinBlock)
	ctx.Step(`^stdin (.+)$`, setStdinLine)

	ctx.Step(`^timeout (\d+)$`, setTimeout)
	ctx.Step(`^size (\d+) (\d+)$`, setSize)
	ctx.Step(`^encoding output (.+)$`, setOutputEncoding)
	ctx.Step(`^encoding stdin (.+)$`, setStdinEncoding)

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
