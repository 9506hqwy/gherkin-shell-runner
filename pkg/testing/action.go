package testing

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/aymanbagabas/go-pty"
	"github.com/cucumber/godog"
)

const attachmentMime = "text/plain"

func execCommand(ctx context.Context) (context.Context, error) {
	t := getTuiFeature(ctx)
	return runFeature(ctx, t)
}

func runFeature(
	ctx context.Context,
	t *tuiFeature,
) (context.Context, error) {
	ptmx, err := pty.New()
	if err != nil {
		return ctx, err
	}

	defer ptmx.Close()

	err = setPty(t, &ptmx)
	if err != nil {
		return ctx, err
	}

	cmd, cancel := createComand(ctx, t, &ptmx)
	defer cancel()

	output, err := runCommand(t, &ptmx, cmd)

	if err == context.DeadlineExceeded {
		ctx = setError(ctx, t, err)
		err = nil
	}

	exitErr, ok := err.(*exec.ExitError)
	if ok {
		ctx = setFail(ctx, t, exitErr)
		err = nil
	}

	if err != nil {
		ctx = setError(ctx, t, err)
		return ctx, err
	}

	t.exitCode = cmd.ProcessState.ExitCode()

	ctx = setOutput(ctx, t, output)

	return ctx, err
}

func createComand(
	ctx context.Context,
	t *tuiFeature,
	ptmx *pty.Pty,
) (*pty.Cmd, context.CancelFunc) {
	deadline, cancel := context.WithTimeout(
		ctx,
		time.Duration(t.timeout)*time.Millisecond,
	)

	cmd := (*ptmx).CommandContext(deadline, t.command, t.args...)

	cmd.Env = os.Environ()
	for key, value := range t.envs {
		env := fmt.Sprintf("%s=%s", key, value)
		cmd.Env = append(cmd.Env, env)
	}

	if t.workspace != EmptyString {
		cmd.Dir = t.workspace
	}

	return cmd, cancel
}

func runCommand(
	t *tuiFeature,
	ptmx *pty.Pty,
	cmd *pty.Cmd,
) (*bytes.Buffer, error) {
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	terminal := newTerminal()
	go terminal.Copy(*ptmx)

	if t.stdin != EmptyString {
		err = inputStdin(t, ptmx)
		if err != nil {
			return nil, err
		}
	}

	err = cmd.Wait()

	// TODO: wait for complete output correctly.
	time.Sleep(time.Duration(t.wait) * time.Millisecond)

	return terminal.Buffer(), err
}

func setOutput(
	ctx context.Context,
	t *tuiFeature,
	output *bytes.Buffer,
) context.Context {
	body := output.Bytes()

	t.output = body

	attachment := godog.Attachment{
		Body:      body,
		FileName:  "output",
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}

func setFail(
	ctx context.Context,
	_ *tuiFeature,
	err *exec.ExitError,
) context.Context {
	body := err.Stderr

	attachment := godog.Attachment{
		Body:      body,
		FileName:  "fail",
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}

func setError(
	ctx context.Context,
	_ *tuiFeature,
	err error,
) context.Context {
	body := err.Error()

	attachment := godog.Attachment{
		Body:      []byte(body),
		FileName:  "error",
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}
