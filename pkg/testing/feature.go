package testing

import (
	"context"

	"github.com/cucumber/godog"
)

const Empty = 0

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
	//revive:disable:add-constant
	t.workspace = ""
	t.envs = make(map[string]string)
	t.command = ""
	t.args = make([]string, 0)
	t.stdin = ""
	t.exitCode = 0
	t.output = make([]byte, 0)
	t.timeout = 60 * 1000
	t.size = nil
	t.ouputEncoding = ""
	t.stdinEncoding = ""
	return t
	//revive:enable:add-constant
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

	ctx.Step(`^output is empty$`, checkOutputIsEmpty)
	ctx.Step(`^output eq$`, checkOutputEqBlock)
	ctx.Step(`^output eq (.+)$`, checkOutputEqLine)
}
