package testing

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

func setCommand(ctx context.Context, command string) (context.Context, error) {
	t := getTuiFeature(ctx)
	initTuituiFeature(t)

	cmd, err := parseValueOne(t, command)
	if err != nil {
		return ctx, err
	}

	t.command = os.ExpandEnv(cmd)
	return ctx, nil
}

func setWorkspace(
	ctx context.Context,
	workspace string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	ws, err := parseValueOne(t, workspace)
	if err != nil {
		return ctx, err
	}

	if !t.delWorkspace {
		t.workspace = os.ExpandEnv(ws)
	}

	return ctx, nil
}

func setEnvironment(
	ctx context.Context,
	key string,
	value string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	v, err := parseValueOne(t, value)
	if err != nil {
		return ctx, err
	}

	t.envs[key] = v
	return ctx, nil
}

func setArgument(ctx context.Context, arg string) (context.Context, error) {
	t := getTuiFeature(ctx)

	a, err := parseValueOne(t, arg)
	if err != nil {
		return ctx, err
	}

	t.args = append(t.args, a)
	return ctx, nil
}

func setStdinBlock(
	ctx context.Context,
	stdin *godog.DocString,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.stdin = stdin.Content
	return ctx, nil
}

func setStdinLine(
	ctx context.Context,
	stdin string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	in, err := parseValueOne(t, stdin)
	if err != nil {
		return ctx, err
	}

	t.stdin = in
	return ctx, nil
}

func setTimeout(ctx context.Context, timeout int) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.timeout = timeout
	return ctx, nil
}

func setSize(
	ctx context.Context,
	width int,
	height int,
) (context.Context, error) {
	size := TerminalSize{
		width:  width,
		height: height,
	}

	t := getTuiFeature(ctx)
	t.size = &size
	return ctx, nil
}

func setOutputEncoding(
	ctx context.Context,
	encoding string,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.ouputEncoding = encoding
	return ctx, nil
}

func setStdinEncoding(
	ctx context.Context,
	encoding string,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.stdinEncoding = encoding
	return ctx, nil
}

func setTempWorkspace(ctx context.Context) (context.Context, error) {
	temp, err := os.MkdirTemp("", "gsr-")
	if err != nil {
		return ctx, err
	}

	t := getTuiFeature(ctx)
	t.workspace = temp
	t.delWorkspace = true
	return ctx, nil
}

func setVariable(
	ctx context.Context,
	name string,
	value string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	v, err := parseValueOne(t, value)
	if err != nil {
		return ctx, err
	}

	t.vars[name] = v
	return ctx, nil
}
