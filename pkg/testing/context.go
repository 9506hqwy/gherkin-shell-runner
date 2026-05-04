package testing

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cucumber/godog"
)

func setCommand(ctx context.Context, command string) (context.Context, error) {
	t := getTuiFeature(ctx)
	resetTuiFeature(t)

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

func setEnvironmentTable(
	ctx context.Context,
	table *godog.Table,
) (context.Context, error) {
	var err error

	for _, row := range table.Rows {
		//revive:disable:add-constant
		if len(row.Cells) != 2 {
			return ctx, errors.New("invalid format data table")
		}
		//revive:enable:add-constant

		ctx, err = setEnvironment(ctx, row.Cells[0].Value, row.Cells[1].Value)
		if err != nil {
			return ctx, err
		}
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

func setWait(ctx context.Context, wait int) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.wait = wait
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
	t.outputEncoding = encoding
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

func setFileEncoding(
	ctx context.Context,
	encoding string,
) (context.Context, error) {
	t := getTuiFeature(ctx)
	t.fileEncoding = encoding
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

func readFile(
	ctx context.Context,
	path string,
	variableName string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	p, err := parseValueOne(t, path)
	if err != nil {
		return ctx, err
	}

	if t.workspace != EmptyString {
		p = filepath.Join(t.workspace, p)
	}

	content, err := os.ReadFile(p)
	if err != nil {
		return ctx, err
	}

	ctx = setFile(ctx, p, content)

	value, err := decodingToBytes(t.fileEncoding, content)
	if err != nil {
		return ctx, err
	}

	t.vars[variableName] = string(value)
	return ctx, nil
}

func writeFileBlock(
	ctx context.Context,
	path string,
	content *godog.DocString,
) (context.Context, error) {
	return writeFile(ctx, path, content.Content)
}

func writeFileLine(
	ctx context.Context,
	path string,
	content string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	c, err := parseValueOne(t, content)
	if err != nil {
		return ctx, err
	}

	return writeFile(ctx, path, c)
}

func writeFile(
	ctx context.Context,
	path string,
	content string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	p, err := parseValueOne(t, path)
	if err != nil {
		return ctx, err
	}

	if t.workspace != EmptyString {
		p = filepath.Join(t.workspace, p)
	}

	ctx = setFile(ctx, p, []byte(content))

	body, err := encodingToBytes(t.fileEncoding, content)
	if err != nil {
		return ctx, err
	}

	return ctx, os.WriteFile(p, body, t.filePermission)
}

func setFile(
	ctx context.Context,
	path string,
	body []byte,
) context.Context {
	attachment := godog.Attachment{
		Body:      body,
		FileName:  path,
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}

func setFilePermission(
	ctx context.Context,
	permission string,
) (context.Context, error) {
	t := getTuiFeature(ctx)

	//revive:disable:add-constant
	perm, err := strconv.ParseInt(permission, 8, 32)
	//revive:enable:add-constant
	if err != nil {
		return ctx, err
	}

	t.filePermission = os.FileMode(perm)

	return ctx, nil
}
