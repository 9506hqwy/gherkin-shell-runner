package testing

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

func setCommand(ctx context.Context, command string) (context.Context, error) {
	t := getTuiFeature(ctx)
	initTuituiFeature(t)
	t.command = os.ExpandEnv(command)
	return ctx, nil
}

func setWorkspace(
	ctx context.Context,
	workspace string,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.workspace = os.ExpandEnv(workspace)
	return ctx, nil
}

func setEnvironment(
	ctx context.Context,
	key string,
	value string,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.envs[key] = value
	return ctx, nil
}

func setArgument(ctx context.Context, arg string) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.args = append(t.args, arg)
	return ctx, nil
}

func setStdinBlock(
	ctx context.Context,
	stdin *godog.DocString,
) (context.Context, error) {
	return setStdinLine(ctx, stdin.Content)
}

func setStdinLine(
	ctx context.Context,
	stdin string,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.stdin = stdin
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
